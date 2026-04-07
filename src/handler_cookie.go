package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Cookie struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate,omitempty"`
	HostOnly       bool    `json:"hostOnly"`
	HTTPOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	StoreID        string  `json:"storeId"`
	Value          string  `json:"value"`
}

const chromiumEpochOffset = 11644473600

func unixToChromium(unixSec float64) int64 {
	return int64((unixSec + chromiumEpochOffset) * 1e6)
}

func chromiumToUnix(chromiumUsec int64) float64 {
	return float64(chromiumUsec)/1e6 - chromiumEpochOffset
}

func chromiumNow() int64 {
	return int64((float64(time.Now().Unix()) + chromiumEpochOffset) * 1e6)
}

func sameSiteToInt(s string) int {
	switch strings.ToLower(s) {
	case "no_restriction", "none":
		return 0
	case "lax":
		return 1
	case "strict":
		return 2
	default:
		return -1
	}
}

func sameSiteToStr(i int) string {
	switch i {
	case 0:
		return "no_restriction"
	case 1:
		return "lax"
	case 2:
		return "strict"
	default:
		return "unspecified"
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func exportCookies(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	rawID := decodeID(id)
	if rawID <= 0 {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid request body"})
		return
	}
	var cookieStr string
	err := db.QueryRow("SELECT cookie FROM profiles WHERE id=?", rawID).Scan(&cookieStr)
	if err != nil {
		writeJSON(w, Response[any]{Code: 404, Message: "profile not found"})
		return
	}
	var cookies []Cookie
	if cookieStr != "" {
		json.Unmarshal([]byte(cookieStr), &cookies)
	}
	if cookies == nil {
		cookies = []Cookie{}
	}
	writeJSON(w, Response[[]Cookie]{Code: 200, Message: "success", Data: cookies})
}

func importCookies(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      string   `json:"id"`
		Cookies []Cookie `json:"cookies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid request body"})
		return
	}
	if len(req.Cookies) == 0 {
		writeJSON(w, Response[any]{Code: 200, Message: "no cookies to import"})
		return
	}

	rawID := decodeID(req.ID)
	var fp FingerprintConfig
	if err := db.QueryRow("SELECT fingerprint FROM profiles WHERE id=?", rawID).Scan(&fp); err != nil {
		writeJSON(w, Response[any]{Code: 404, Message: "profile not found"})
		return
	}
	userDataDir := filepath.Join(dataDir, "profiles", encodeID(int64(fp.Seed)))
	absUserDataDir, _ := filepath.Abs(userDataDir)
	cookiePath := filepath.Join(absUserDataDir, "Default", "Cookies")

	writeCookiesToFile(cookiePath, req.Cookies)
	if data, err := json.Marshal(req.Cookies); err == nil {
		db.Exec("UPDATE profiles SET cookie=? WHERE id=?", string(data), rawID)
	}
	writeJSON(w, Response[any]{Code: 200, Message: "success"})
}

func readCookiesFromFile(cookiePath string) []Cookie {
	if _, err := os.Stat(cookiePath); os.IsNotExist(err) {
		return nil
	}
	cookieDB, err := sql.Open("sqlite", cookiePath)
	if err != nil {
		log.Printf("[Cookie] failed to open %s: %v", cookiePath, err)
		return nil
	}
	defer cookieDB.Close()

	rows, err := cookieDB.Query(`SELECT host_key, name, value, path, expires_utc,
		is_secure, is_httponly, has_expires, samesite FROM cookies`)
	if err != nil {
		log.Printf("[Cookie] query error: %v", err)
		return nil
	}
	defer rows.Close()

	var cookies []Cookie
	for rows.Next() {
		var hostKey, name, value, path string
		var expiresUTC int64
		var isSecure, isHTTPOnly, hasExpires, sameSite int

		if err := rows.Scan(&hostKey, &name, &value, &path, &expiresUTC,
			&isSecure, &isHTTPOnly, &hasExpires, &sameSite); err != nil {
			continue
		}

		c := Cookie{
			Domain:   hostKey,
			Name:     name,
			Value:    value,
			Path:     path,
			Secure:   isSecure == 1,
			HTTPOnly: isHTTPOnly == 1,
			Session:  hasExpires == 0,
			HostOnly: !strings.HasPrefix(hostKey, "."),
			SameSite: sameSiteToStr(sameSite),
			StoreID:  "0",
		}
		if hasExpires == 1 && expiresUTC > 0 {
			c.ExpirationDate = math.Round(chromiumToUnix(expiresUTC)*1000) / 1000
		}
		cookies = append(cookies, c)
	}
	return cookies
}

func writeCookiesToFile(cookiePath string, cookies []Cookie) {
	os.MkdirAll(filepath.Dir(cookiePath), 0755)

	cookieDB, err := sql.Open("sqlite", cookiePath)
	if err != nil {
		log.Printf("[Cookie] failed to open %s: %v", cookiePath, err)
		return
	}
	defer cookieDB.Close()

	cookieDB.Exec(`CREATE TABLE IF NOT EXISTS cookies (
		creation_utc INTEGER NOT NULL,
		host_key TEXT NOT NULL,
		top_frame_site_key TEXT NOT NULL,
		name TEXT NOT NULL,
		value TEXT NOT NULL,
		encrypted_value BLOB NOT NULL,
		path TEXT NOT NULL,
		expires_utc INTEGER NOT NULL,
		is_secure INTEGER NOT NULL,
		is_httponly INTEGER NOT NULL,
		last_access_utc INTEGER NOT NULL,
		has_expires INTEGER NOT NULL,
		is_persistent INTEGER NOT NULL,
		priority INTEGER NOT NULL,
		samesite INTEGER NOT NULL,
		source_scheme INTEGER NOT NULL,
		source_port INTEGER NOT NULL,
		last_update_utc INTEGER NOT NULL,
		source_type INTEGER NOT NULL,
		has_cross_site_ancestor INTEGER NOT NULL
	)`)
	cookieDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS cookies_unique_index
		ON cookies(host_key, top_frame_site_key, has_cross_site_ancestor,
		           name, path, source_scheme, source_port)`)

	tx, err := cookieDB.Begin()
	if err != nil {
		log.Printf("[Cookie] transaction error: %v", err)
		return
	}

	stmt, _ := tx.Prepare(`INSERT OR REPLACE INTO cookies
		(creation_utc, host_key, top_frame_site_key, name, value, encrypted_value,
		 path, expires_utc, is_secure, is_httponly, last_access_utc, has_expires,
		 is_persistent, priority, samesite, source_scheme, source_port,
		 last_update_utc, source_type, has_cross_site_ancestor)
		VALUES (?, ?, '', ?, ?, '', ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, -1, ?, 0, 0)`)
	if stmt == nil {
		tx.Rollback()
		return
	}
	defer stmt.Close()

	now := chromiumNow()
	for _, c := range cookies {
		hasExpires := 1
		isPersistent := 1
		var expiresUTC int64
		if c.Session || c.ExpirationDate == 0 {
			hasExpires = 0
			isPersistent = 0
			expiresUTC = 0
		} else {
			expiresUTC = unixToChromium(c.ExpirationDate)
		}
		sourceScheme := 0
		if c.Secure {
			sourceScheme = 2
		}

		stmt.Exec(now, c.Domain, c.Name, c.Value,
			c.Path, expiresUTC, boolToInt(c.Secure), boolToInt(c.HTTPOnly), now,
			hasExpires, isPersistent, sameSiteToInt(c.SameSite), sourceScheme, now)
	}
	tx.Commit()
	log.Printf("[Cookie] wrote %d cookies to %s", len(cookies), cookiePath)
}
