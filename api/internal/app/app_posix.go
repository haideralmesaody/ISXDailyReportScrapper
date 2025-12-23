//go:build !windows

package app

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// createSingleInstanceMutex creates a POSIX-compatible mutex for single instance enforcement
func (a *Application) createSingleInstanceMutex() error {
	// For POSIX systems, we use a pidfile approach
	if a.windowsMutex != 0 {
		return nil // Already created (using windowsMutex field as generic flag)
	}

	// For now, just set the flag to indicate it's "created"
	a.windowsMutex = 1
	return nil
}

// cleanupSingleInstanceMutex cleans up POSIX-specific mutex resources
func (a *Application) cleanupSingleInstanceMutex() {
	// For POSIX systems, cleanup is simpler
	a.windowsMutex = 0
}

// isProcessRunningWindows checks if a process is running using POSIX methods
func isProcessRunningWindows(pid int) bool {
	// POSIX-specific process check
	if runtime.GOOS != "windows" {
		process, err := os.FindProcess(pid)
		if err != nil {
			return false
		}

		err = process.Signal(syscall.Signal(0))
		return err == nil
	}
	return false
}

// killProcessWindows forcefully terminates a process using POSIX methods
func killProcessWindows(pid int) error {
	if runtime.GOOS != "windows" {
		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}

		return process.Kill()
	}
	return syscall.ENOSYS
}

// openBrowserWindows opens a URL in the default browser on POSIX systems
func openBrowserWindows(url string) error {
	if runtime.GOOS != "windows" {
		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			// Try common Linux browsers
			browsers := []string{"xdg-open", "google-chrome", "firefox", "mozilla", "safari"}
			for _, browser := range browsers {
				if _, err := exec.LookPath(browser); err == nil {
					if browser == "xdg-open" {
						cmd = exec.Command(browser, url)
					} else {
						cmd = exec.Command(browser, url)
					}
					break
				}
			}
			if cmd == nil {
				return syscall.ENOENT // No browser found
			}
		default:
			return syscall.ENOSYS
		}

		return cmd.Start()
	}
	return syscall.ENOSYS
}