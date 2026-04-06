//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32              = syscall.MustLoadDLL("user32.dll")
	procEnumWindows     = user32.MustFindProc("EnumWindows")
	procGetWindowPID    = user32.MustFindProc("GetWindowThreadProcessId")
	procIsWindowVisible = user32.MustFindProc("IsWindowVisible")
	procGetWindow       = user32.MustFindProc("GetWindow")
	procSetForeground   = user32.MustFindProc("SetForegroundWindow")
	procShowWindow      = user32.MustFindProc("ShowWindow")
	procPostMessage     = user32.MustFindProc("PostMessageW")
	procIsIconic        = user32.MustFindProc("IsIconic")
)

const (
	gwOwner   = uintptr(4)
	swRestore = uintptr(9)
	wmClose   = uintptr(0x0010)
)

func isMainWindow(hwnd syscall.Handle) bool {
	isVisible, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	isOwned, _, _ := procGetWindow.Call(uintptr(hwnd), gwOwner)
	return isVisible == 1 && isOwned == 0
}

func getWindowPID(hwnd syscall.Handle) uint32 {
	var pid uint32
	procGetWindowPID.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
	return pid
}

func bringWindowToFront(pid uint32) {
	cb := syscall.NewCallback(func(hwnd syscall.Handle, _ uintptr) uintptr {
		if getWindowPID(hwnd) == pid && isMainWindow(hwnd) {
			isMinimized, _, _ := procIsIconic.Call(uintptr(hwnd))
			if isMinimized != 0 {
				procShowWindow.Call(uintptr(hwnd), swRestore)
			}
			procSetForeground.Call(uintptr(hwnd))
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
}

func closeWindowsByPID(pid uint32) {
	cb := syscall.NewCallback(func(hwnd syscall.Handle, _ uintptr) uintptr {
		if getWindowPID(hwnd) == pid {
			procPostMessage.Call(uintptr(hwnd), wmClose, 0, 0)
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
}
