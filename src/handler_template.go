package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Template struct {
	ID          string            `json:"_id"`
	Name        string            `json:"name"`
	Sort        int               `json:"sort"`
	Proxy       string            `json:"proxy"`
	Fingerprint FingerprintConfig `json:"fingerprint"`
	Args        string            `json:"args"`
	Notes       string            `json:"notes"`
	CreatedAt   int64             `json:"createdAt"`
	UpdatedAt   int64             `json:"updatedAt"`
}

func getTemplates(w http.ResponseWriter, r *http.Request) {
	page, pageSize := 1, 10
	fmt.Sscanf(r.URL.Query().Get("page"), "%d", &page)
	fmt.Sscanf(r.URL.Query().Get("pageSize"), "%d", &pageSize)
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}

	if r.URL.Query().Get("all") == "1" {
		rows, err := db.Query("SELECT id, name, sort, proxy, fingerprint, args, notes, created_at, updated_at FROM templates ORDER BY sort DESC, created_at DESC")
		if err != nil {
			writeJSON(w, Response[any]{Code: 500, Message: err.Error()})
			return
		}
		defer rows.Close()

		var list []Template
		for rows.Next() {
			var t Template
			var rawID, rawProxy int64
			if err := rows.Scan(&rawID, &t.Name, &t.Sort, &rawProxy, &t.Fingerprint, &t.Args, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err == nil {
				t.ID = encodeID(rawID)
				if rawProxy > 0 {
					t.Proxy = encodeID(rawProxy)
				}
				list = append(list, t)
			}
		}
		writeJSON(w, Response[any]{Code: 200, Message: "success", Data: list})
		return
	}

	query := "SELECT id, name, sort, proxy, fingerprint, args, notes, created_at, updated_at FROM templates WHERE 1=1"
	var args []interface{}

	if keyword != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM ("+query+")", args...).Scan(&total); err != nil {
		writeJSON(w, Response[any]{Code: 500, Message: err.Error()})
		return
	}

	query += " ORDER BY sort DESC, created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := db.Query(query, args...)
	if err != nil {
		writeJSON(w, Response[any]{Code: 500, Message: err.Error()})
		return
	}
	defer rows.Close()

	var list []Template
	for rows.Next() {
		var t Template
		var rawID, rawProxy int64
		if err := rows.Scan(&rawID, &t.Name, &t.Sort, &rawProxy, &t.Fingerprint, &t.Args, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err == nil {
			t.ID = encodeID(rawID)
			if rawProxy > 0 {
				t.Proxy = encodeID(rawProxy)
			}
			list = append(list, t)
		}
	}
	writeJSON(w, Response[any]{Code: 200, Message: "success", Data: map[string]interface{}{
		"list":  list,
		"total": total,
	}})
}

func getTemplate(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	rawID := decodeID(idStr)
	if rawID <= 0 {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid request body"})
		return
	}
	var t Template
	var rawProxy int64
	err := db.QueryRow("SELECT id, name, sort, proxy, fingerprint, args, notes, created_at, updated_at FROM templates WHERE id=?", rawID).
		Scan(&rawID, &t.Name, &t.Sort, &rawProxy, &t.Fingerprint, &t.Args, &t.Notes, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		writeJSON(w, Response[any]{Code: 404, Message: "template not found"})
		return
	}
	t.ID = encodeID(rawID)
	if rawProxy > 0 {
		t.Proxy = encodeID(rawProxy)
	}
	writeJSON(w, Response[any]{Code: 200, Message: "success", Data: t})
}

func addTemplate(w http.ResponseWriter, r *http.Request) {
	var t Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid request body"})
		return
	}
	t.CreatedAt = time.Now().UnixMilli()
	t.UpdatedAt = t.CreatedAt
	rawProxy := decodeID(t.Proxy)

	res, err := db.Exec("INSERT INTO templates (name, sort, proxy, fingerprint, args, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		t.Name, t.Sort, rawProxy, t.Fingerprint, t.Args, t.Notes, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "UNIQUE") {
			msg = "模板名称已存在！"
		}
		writeJSON(w, Response[any]{Code: 500, Message: msg})
		return
	}
	rowID, _ := res.LastInsertId()
	t.ID = encodeID(rowID)
	writeJSON(w, Response[any]{Code: 200, Message: "success", Data: t})
}

func updateTemplate(w http.ResponseWriter, r *http.Request) {
	var t Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid request body"})
		return
	}
	t.UpdatedAt = time.Now().UnixMilli()
	rawID := decodeID(t.ID)
	rawProxy := decodeID(t.Proxy)

	_, err := db.Exec("UPDATE templates SET name=?, sort=?, proxy=?, fingerprint=?, args=?, notes=?, updated_at=? WHERE id=?",
		t.Name, t.Sort, rawProxy, t.Fingerprint, t.Args, t.Notes, t.UpdatedAt, rawID)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "UNIQUE") {
			msg = "模板名称已存在！"
		}
		writeJSON(w, Response[any]{Code: 500, Message: msg})
		return
	}
	writeJSON(w, Response[any]{Code: 200, Message: "success", Data: t})
}

func deleteTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid request body"})
		return
	}
	rawID := decodeID(req.ID)
	if _, err := db.Exec("DELETE FROM templates WHERE id=?", rawID); err != nil {
		writeJSON(w, Response[any]{Code: 500, Message: err.Error()})
		return
	}
	writeJSON(w, Response[any]{Code: 200, Message: "success"})
}

func createFromTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TemplateID string `json:"templateId"`
		Name       string `json:"name"`
		GroupID    string `json:"groupId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid request body"})
		return
	}

	rawTemplateID := decodeID(req.TemplateID)
	if rawTemplateID <= 0 {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid template id"})
		return
	}

	var t Template
	var rawProxy int64
	err := db.QueryRow("SELECT id, name, sort, proxy, fingerprint, args, notes FROM templates WHERE id=?", rawTemplateID).
		Scan(&rawTemplateID, &t.Name, &t.Sort, &rawProxy, &t.Fingerprint, &t.Args, &t.Notes)
	if err != nil {
		writeJSON(w, Response[any]{Code: 404, Message: "template not found"})
		return
	}

	p := Profile{
		Name:        req.Name,
		GroupID:     req.GroupID,
		Sort:        t.Sort,
		Proxy:       t.Proxy,
		Fingerprint: t.Fingerprint,
		Args:        t.Args,
		Notes:       t.Notes,
		CreatedAt:   time.Now().UnixMilli(),
	}
	p.UpdatedAt = p.CreatedAt
	enrichFingerprint(&p, true)

	rawGroupID := decodeID(p.GroupID)

	res, err := db.Exec("INSERT INTO profiles (name, group_id, sort, proxy, args, fingerprint, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, rawGroupID, p.Sort, rawProxy, p.Args, p.Fingerprint, p.Notes, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "UNIQUE") {
			msg = "配置名称已存在！"
		}
		writeJSON(w, Response[any]{Code: 500, Message: msg})
		return
	}
	rowID, _ := res.LastInsertId()
	p.ID = encodeID(rowID)
	if rawGroupID > 0 {
		p.GroupID = encodeID(rawGroupID)
	}
	if rawProxy > 0 {
		p.Proxy = encodeID(rawProxy)
	}
	writeJSON(w, Response[any]{Code: 200, Message: "success", Data: p})
}

func saveAsTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProfileID    string `json:"profileId"`
		TemplateName string `json:"templateName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid request body"})
		return
	}

	rawProfileID := decodeID(req.ProfileID)
	if rawProfileID <= 0 {
		writeJSON(w, Response[any]{Code: 400, Message: "invalid profile id"})
		return
	}

	var rawProxy int64
	var fp FingerprintConfig
	var args, notes string
	err := db.QueryRow("SELECT proxy, fingerprint, args, notes FROM profiles WHERE id=?", rawProfileID).
		Scan(&rawProxy, &fp, &args, &notes)
	if err != nil {
		writeJSON(w, Response[any]{Code: 404, Message: "profile not found"})
		return
	}

	fp.Seed = 0
	now := time.Now().UnixMilli()
	res, err := db.Exec("INSERT INTO templates (name, sort, proxy, fingerprint, args, notes, created_at, updated_at) VALUES (?, 0, ?, ?, ?, ?, ?, ?)",
		req.TemplateName, rawProxy, fp, args, notes, now, now)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "UNIQUE") {
			msg = "模板名称已存在！"
		}
		writeJSON(w, Response[any]{Code: 500, Message: msg})
		return
	}
	rowID, _ := res.LastInsertId()
	t := Template{
		ID:          encodeID(rowID),
		Name:        req.TemplateName,
		Proxy:       "",
		Fingerprint: fp,
		Args:        args,
		Notes:       notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if rawProxy > 0 {
		t.Proxy = encodeID(rawProxy)
	}
	writeJSON(w, Response[any]{Code: 200, Message: "success", Data: t})
}
