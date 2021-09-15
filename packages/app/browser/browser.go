package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Browser - Container for Chrome/Edge locator
type Browser struct {
	ExecPath string
}

// GetExecPath - Get real execPath for Chrome/Edge, if not specified.
func (b Browser) GetExecPath() string {
	path := b.ExecPath

	switch path {
	case "":
		path = b.locateChrome()
		if path == "" {
			path = b.locateEdge()
		}
	case "chrome", "google-chrome", "chromium":
		path = b.locateChrome()
	case "edge", "msedge", "microsoft-edge":
		path = b.locateEdge()
	}

	return path
}

func (Browser) locateChrome() string {
	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/usr/bin/ungoogled-chromium",
			"/usr/bin/ungoogled-chromium-browser",
		}
	case "windows":
		paths = []string{
			os.Getenv("LocalAppData") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("ProgramFiles") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Google/Chrome/Application/chrome.exe",
			os.Getenv("LocalAppData") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles") + "/Chromium/Application/chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "/Chromium/Application/chrome.exe",
		}
	default:
		paths = []string{
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
			"/usr/bin/ungoogled-chromium",
			"/usr/bin/ungoogled-chromium-browser",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		return path
	}

	if runtime.GOOS == "linux" {
		flatpakSet := make(map[string]bool)

		b, e := exec.Command("flatpak", "list", "--columns=application").Output()
		if e == nil && len(b) > 0 {
			for _, p := range strings.Split(string(b), "\n") {
				flatpakSet[p] = true
			}
		}

		if len(flatpakSet) > 0 {
			appIDs := []string{
				"org.chromium.Chromium",
				"com.github.Eloston.UngoogledChromium",
			}
			paths = make([]string, 0)

			for _, id := range appIDs {
				if flatpakSet[id] {
					paths = append(
						paths,
						fmt.Sprintf("/var/lib/flatpak/exports/bin/%s", id),
						os.Getenv("HOME")+fmt.Sprintf("/.local/share/flatpak/exports/bin/%s", id),
					)
				}
			}

			if len(paths) > 0 {
				for _, path := range paths {
					if _, err := os.Stat(path); os.IsNotExist(err) {
						continue
					}

					return path
				}
			}
		}
	}

	return ""
}

func (Browser) locateEdge() string {
	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/usr/bin/microsoft-edge",
			"/usr/bin/microsoft-edge-beta",
			"/usr/bin/microsoft-edge-dev",
		}
	case "windows":
		paths = []string{
			os.Getenv("ProgramFiles") + "/Microsoft/Edge/Application/msedge.exe",
			os.Getenv("ProgramFiles(x86)") + "/Microsoft/Edge/Application/msedge.exe",
		}
	default:
		paths = []string{
			"/usr/bin/microsoft-edge",
			"/usr/bin/microsoft-edge-beta",
			"/usr/bin/microsoft-edge-dev",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		return path
	}
	return ""
}
