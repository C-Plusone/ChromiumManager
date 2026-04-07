package main

import (
	"log"
	"math/rand/v2"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
)

type runningProfile struct {
	cmd        *exec.Cmd
	done       chan struct{}
	profileID  string
	cookiePath string
}

var (
	runningMu       sync.Mutex
	runningProfiles = map[string]*runningProfile{}
)

// getRunningIDs returns a snapshot of currently running profile IDs.
// Caller must NOT hold runningMu.
func getRunningIDs() []string {
	runningMu.Lock()
	ids := make([]string, 0, len(runningProfiles))
	for id := range runningProfiles {
		ids = append(ids, id)
	}
	runningMu.Unlock()
	slices.Sort(ids)
	return ids
}

func findBrowserPath() string {
	local := filepath.Join("bin", "chromium", "chrome-wrapper")
	if runtime.GOOS == "windows" {
		local = filepath.Join("bin", "chromium", "chrome.exe")
	}
	for _, name := range []string{"chrome", local} {
		if p, err := exec.LookPath(name); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return ""
}

// proxyInfo holds the fields we need from a proxy record.
type proxyInfo struct {
	URL      string
	Timezone string
	Lang     string
	Location string
}

// getProxyInfo fetches proxy info (url, timezone, lang) by proxy rowid.
// Returns zero-value struct when proxyRawID <= 0.
func getProxyInfo(proxyRawID int64) proxyInfo {
	if proxyRawID <= 0 {
		return proxyInfo{}
	}
	var info proxyInfo
	if err := db.QueryRow("SELECT url, timezone, lang, location FROM proxies WHERE id=?", proxyRawID).
		Scan(&info.URL, &info.Timezone, &info.Lang, &info.Location); err != nil {
		log.Printf("[DB] getProxyInfo(%d) error: %v", proxyRawID, err)
	}
	return info
}

// enrichFingerprint fills in Seed on new profiles.
func enrichFingerprint(p *Profile, isNew bool) {
	if isNew {
		p.Fingerprint.Seed = rand.Int32()
	}
}

// splitArgs splits a command-line string into arguments.
func splitArgs(s string) []string {
	parts := strings.Split(s, "--")
	var args []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			args = append(args, "--"+p)
		}
	}
	return args
}
