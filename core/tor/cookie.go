package tor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// cookieRelPaths are known relative paths from a Tor install root to the cookie file.
var cookieRelPaths = []string{
	"control_auth_cookie",
	filepath.Join("Data", "Tor", "control_auth_cookie"),
	filepath.Join("data", "Tor", "control_auth_cookie"),
	filepath.Join("Browser", "TorBrowser", "Data", "Tor", "control_auth_cookie"),
	filepath.Join("TorBrowser", "Data", "Tor", "control_auth_cookie"),
}

func defaultCookiePath() string {
	switch runtime.GOOS {
	case "linux":
		paths := []string{
			"/run/tor/control.authcookie",
			filepath.Join(os.Getenv("HOME"), ".local/share/tor/control.authcookie"),
		}
		for _, p := range paths {
			if fileExists(p) {
				return p
			}
		}
		return paths[0]

	case "darwin":
		paths := []string{
			filepath.Join(os.Getenv("HOME"), "Library/Application Support/TorBrowser-Data/Tor/control_auth_cookie"),
			filepath.Join(os.Getenv("HOME"), "Library/Application Support/tor/control_auth_cookie"),
		}
		for _, p := range paths {
			if fileExists(p) {
				return p
			}
		}
		return paths[0]

	case "windows":
		if p := findWindowsCookiePath(); p != "" {
			return p
		}
		return filepath.Join(os.Getenv("APPDATA"), "tor", "control_auth_cookie")

	default:
		return "/var/lib/tor/control_auth_cookie"
	}
}

func findWindowsCookiePath() string {
	// ── مرحله ۱: مسیرهای استاندارد (سریع‌ترین) ──────────────────────
	appdata := os.Getenv("APPDATA")
	localappdata := os.Getenv("LOCALAPPDATA")
	userprofile := os.Getenv("USERPROFILE")
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")

	standardBases := []string{
		filepath.Join(appdata, "tor"),
		filepath.Join(appdata, "Tor"),
		filepath.Join(appdata, "Tor Browser"),
		filepath.Join(localappdata, "Tor Browser"),
		filepath.Join(programFiles, "Tor Browser"),
		filepath.Join(programFilesX86, "Tor Browser"),
		filepath.Join(userprofile, "Tor Browser"),
		filepath.Join(userprofile, "Desktop", "Tor Browser"),
		filepath.Join(userprofile, "Downloads", "Tor Browser"),
	}

	for _, base := range standardBases {
		if p := checkCookieInDir(base); p != "" {
			return p
		}
	}

	// ── مرحله ۲: tor.exe در PATH ──────────────────────────────────────
	if torExe, err := exec.LookPath("tor.exe"); err == nil {
		if p := cookieFromTorExe(torExe); p != "" {
			return p
		}
	}

	// ── مرحله ۳: Registry ─────────────────────────────────────────────
	if p := findTorInRegistry(); p != "" {
		return p
	}

	// ── مرحله ۴: جستجو در پوشه‌های سطح اول همه درایوها ──────────────
	// فقط depth=2 و فقط پوشه‌هایی که اسمشون به "tor" ربط داره
	for _, drive := range getWindowsDrives() {
		if p := searchTopLevelDirs(drive); p != "" {
			return p
		}
	}

	return ""
}

// checkCookieInDir checks all known relative cookie paths inside a base dir.
func checkCookieInDir(base string) string {
	for _, rel := range cookieRelPaths {
		p := filepath.Join(base, rel)
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// cookieFromTorExe infers cookie path from tor.exe location.
func cookieFromTorExe(torExe string) string {
	dir := filepath.Dir(torExe)
	relatives := []string{
		dir,
		filepath.Join(dir, ".."),
		filepath.Join(dir, "..", ".."),
		filepath.Join(dir, "..", "..", ".."),
	}
	for _, base := range relatives {
		if p := checkCookieInDir(filepath.Clean(base)); p != "" {
			return p
		}
	}
	return ""
}

// findTorInRegistry queries Windows registry for Tor Browser install path.
func findTorInRegistry() string {
	keys := []string{
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\Tor Browser`,
		`HKLM\Software\Microsoft\Windows\CurrentVersion\Uninstall\Tor Browser`,
		`HKLM\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\Tor Browser`,
	}
	for _, key := range keys {
		out, err := exec.Command("reg", "query", key, "/v", "InstallLocation").Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			if !strings.Contains(line, "InstallLocation") {
				continue
			}
			parts := strings.SplitN(line, "REG_SZ", 2)
			if len(parts) != 2 {
				continue
			}
			installDir := strings.TrimSpace(parts[1])
			if p := checkCookieInDir(installDir); p != "" {
				return p
			}
		}
	}
	return ""
}

// getWindowsDrives returns available drive letters on Windows.
func getWindowsDrives() []string {
	out, err := exec.Command("fsutil", "fsinfo", "drives").Output()
	if err != nil {
		return []string{`C:\`}
	}
	// خروجی: "Drives: C:\ D:\ E:\"
	var drives []string
	for _, part := range strings.Fields(string(out)) {
		if len(part) == 3 && part[1] == ':' && part[2] == '\\' {
			drives = append(drives, part)
		}
	}
	if len(drives) == 0 {
		return []string{`C:\`}
	}
	return drives
}

// searchTopLevelDirs searches only depth-1 and depth-2 directories whose
// names contain "tor" or "browser" (case-insensitive) on a given drive.
func searchTopLevelDirs(drive string) string {
	entries, err := os.ReadDir(drive)
	if err != nil {
		return ""
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		base := filepath.Join(drive, e.Name())

		// depth-1: اگه اسم پوشه مستقیماً به Tor ربط داره
		if isTorRelated(e.Name()) {
			if p := checkCookieInDir(base); p != "" {
				return p
			}
		}

		// depth-2: داخل هر پوشه سطح اول (مثل C:\Tools\) دنبال Tor میگردیم
		subs, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, sub := range subs {
			if !sub.IsDir() || !isTorRelated(sub.Name()) {
				continue
			}
			subBase := filepath.Join(base, sub.Name())
			if p := checkCookieInDir(subBase); p != "" {
				return p
			}
		}
	}
	return ""
}

// isTorRelated returns true if a directory name is likely Tor-related.
func isTorRelated(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "tor")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// readCookieFile reads and validates a Tor auth cookie file (must be 32 bytes).
func readCookieFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)
	if clean != path {
		return nil, fmt.Errorf("tor: suspicious cookie path: %q", path)
	}
	data, err := os.ReadFile(clean)
	if err != nil {
		return nil, err
	}
	if len(data) != 32 {
		return nil, fmt.Errorf("tor: cookie file is %d bytes, expected 32", len(data))
	}
	return data, nil
}
