//go:build linux && !android

package main

import (
	"log"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type xContext struct {
	mu   sync.Mutex
	conn *xgb.Conn
	root xproto.Window

	// Cached atoms
	atomClientList  xproto.Atom
	atomPID         xproto.Atom
	atomActiveWin   xproto.Atom
	atomCloseWin    xproto.Atom
}

var xCtx xContext

func (x *xContext) ensureConn() error {
	if x.conn != nil {
		return nil
	}
	conn, err := xgb.NewConn()
	if err != nil {
		return err
	}
	x.conn = conn
	x.root = xproto.Setup(conn).DefaultScreen(conn).Root
	x.atomClientList = internAtom(conn, "_NET_CLIENT_LIST")
	x.atomPID = internAtom(conn, "_NET_WM_PID")
	x.atomActiveWin = internAtom(conn, "_NET_ACTIVE_WINDOW")
	x.atomCloseWin = internAtom(conn, "_NET_CLOSE_WINDOW")
	return nil
}

func (x *xContext) reset() {
	if x.conn != nil {
		x.conn.Close()
		x.conn = nil
	}
}

func internAtom(conn *xgb.Conn, name string) xproto.Atom {
	reply, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply()
	if err != nil {
		return 0
	}
	return reply.Atom
}

func windowsByPID(conn *xgb.Conn, root xproto.Window, clientListAtom, pidAtom xproto.Atom, pid uint32) []xproto.Window {
	if clientListAtom == 0 || pidAtom == 0 {
		return nil
	}

	listReply, err := xproto.GetProperty(conn, false, root, clientListAtom,
		xproto.GetPropertyTypeAny, 0, 1024).Reply()
	if err != nil || listReply.Format != 32 {
		return nil
	}

	var matches []xproto.Window
	for i := 0; i+4 <= len(listReply.Value); i += 4 {
		winID := xproto.Window(xgb.Get32(listReply.Value[i:]))
		pidReply, err := xproto.GetProperty(conn, false, winID, pidAtom,
			xproto.GetPropertyTypeAny, 0, 1).Reply()
		if err != nil || pidReply.Format != 32 || len(pidReply.Value) < 4 {
			continue
		}
		if xgb.Get32(pidReply.Value) == pid {
			matches = append(matches, winID)
		}
	}
	return matches
}

func sendClientMsg(conn *xgb.Conn, root, win xproto.Window, atom xproto.Atom, data [5]uint32) {
	event := xproto.ClientMessageEvent{
		Format: 32,
		Window: win,
		Type:   atom,
		Data:   xproto.ClientMessageDataUnionData32New(data[:]),
	}
	xproto.SendEvent(conn, false, root,
		xproto.EventMaskSubstructureRedirect|xproto.EventMaskSubstructureNotify,
		string(event.Bytes()))
}

func bringWindowToFront(pid uint32) {
	xCtx.mu.Lock()
	defer xCtx.mu.Unlock()

	if err := xCtx.ensureConn(); err != nil {
		log.Printf("[Window] X server connection failed: %v", err)
		return
	}

	wins := windowsByPID(xCtx.conn, xCtx.root, xCtx.atomClientList, xCtx.atomPID, pid)
	if len(wins) == 0 {
		return
	}

	if xCtx.atomActiveWin == 0 {
		return
	}
	sendClientMsg(xCtx.conn, xCtx.root, wins[0], xCtx.atomActiveWin, [5]uint32{2, 0, 0, 0, 0})
	xCtx.conn.Sync()
}

func closeWindowsByPID(pid uint32) {
	xCtx.mu.Lock()
	defer xCtx.mu.Unlock()

	if err := xCtx.ensureConn(); err != nil {
		log.Printf("[Window] X server connection failed: %v", err)
		return
	}

	wins := windowsByPID(xCtx.conn, xCtx.root, xCtx.atomClientList, xCtx.atomPID, pid)
	if len(wins) == 0 {
		return
	}

	if xCtx.atomCloseWin == 0 {
		return
	}
	for _, win := range wins {
		sendClientMsg(xCtx.conn, xCtx.root, win, xCtx.atomCloseWin, [5]uint32{0, 2, 0, 0, 0})
	}
	xCtx.conn.Sync()
}
