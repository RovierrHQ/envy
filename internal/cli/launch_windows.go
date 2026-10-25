//go:build windows

package cli

import (
	"os"
	"syscall"
	"unsafe"
)

func shouldShowLaunchHelp() bool {
	if len(os.Args) != 1 {
		return false
	}
	// Explorer double-click attaches a console with only this process.
	// A real terminal session has the shell attached too (count > 1).
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleProcessList := kernel32.NewProc("GetConsoleProcessList")
	var buf [8]uint32
	n, _, _ := getConsoleProcessList.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return n == 1
}

func showLaunchHelp() {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")
	title, _ := syscall.UTF16PtrFromString("envy")
	text, _ := syscall.UTF16PtrFromString(launchHelp)
	const mbOK = 0x0
	const mbIconInformation = 0x40
	messageBox.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), mbOK|mbIconInformation)
}
