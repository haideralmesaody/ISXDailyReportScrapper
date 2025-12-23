//go:build windows

package app

import (
	"os/exec"
	"runtime"
	"syscall"

	"golang.org/x/sys/windows"
)

// createSingleInstanceMutex creates a Windows-specific mutex for single instance enforcement
func (a *Application) createSingleInstanceMutex() error {
	if a.windowsMutex != 0 {
		return nil // Already created
	}

	// Create Windows mutex for single instance
	mutexName, err := windows.UTF16PtrFromString("Global\\ISXPulse_SingleInstance_Mutex")
	if err != nil {
		return err
	}

	handle, err := windows.CreateMutex(nil, true, mutexName)
	if err != nil {
		return err
	}

	a.windowsMutex = uintptr(handle)

	// Check if another instance is already running
	if err == windows.ERROR_ALREADY_EXISTS {
		windows.CloseHandle(handle)
		a.windowsMutex = 0
		return syscall.EINVAL
	}

	return nil
}

// cleanupSingleInstanceMutex cleans up Windows-specific mutex resources
func (a *Application) cleanupSingleInstanceMutex() {
	if a.windowsMutex != 0 && runtime.GOOS == "windows" {
		windows.CloseHandle(windows.Handle(a.windowsMutex))
		a.windowsMutex = 0
	}
}

// isProcessRunningWindows checks if a process is running using Windows-specific APIs
func isProcessRunningWindows(pid int) bool {
	// Try Windows-specific method first
	if runtime.GOOS == "windows" {
		handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
		if err != nil {
			return false
		}
		defer windows.CloseHandle(handle)

		var exitCode uint32
		err = windows.GetExitCodeProcess(handle, &exitCode)
		return err == nil && exitCode == 259 // STILL_ACTIVE (259)
	}
	return false
}

// killProcessWindows forcefully terminates a process using Windows-specific APIs
func killProcessWindows(pid int) error {
	if runtime.GOOS == "windows" {
		// Try Windows-specific termination
		handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
		if err == nil {
			defer windows.CloseHandle(handle)
			err = windows.TerminateProcess(handle, 1)
			if err == nil {
				return nil
			}
		}
	}
	return syscall.EINVAL
}

// openBrowserWindows opens a URL in the default browser on Windows
func openBrowserWindows(url string) error {
	if runtime.GOOS == "windows" {
		// Use simpler approach with exec.Command
		cmd := exec.Command("cmd", "/c", "start", url)
		return cmd.Start()
	}
	return syscall.ENOSYS
}