package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSE client registry
var (
	sseMu      sync.Mutex
	sseClients = map[chan string]struct{}{}
)

// broadcastRunning sends the current list of running profile IDs to all SSE clients.
func broadcastRunning() {
	data, _ := json.Marshal(getRunningIDs())
	msg := string(data)

	sseMu.Lock()
	for ch := range sseClients {
		select {
		case ch <- msg:
		default:
		}
	}
	sseMu.Unlock()
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 4)
	sseMu.Lock()
	sseClients[ch] = struct{}{}
	sseMu.Unlock()
	defer func() {
		sseMu.Lock()
		delete(sseClients, ch)
		sseMu.Unlock()
	}()

	// Send initial state
	initData, _ := json.Marshal(getRunningIDs())
	fmt.Fprintf(w, "data: %s\n\n", initData)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	ctx := r.Context()
	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-ctx.Done():
			return
		}
	}
}
