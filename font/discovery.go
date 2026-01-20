package font

import (
	"os"
	"path/filepath"
	"runtime"
)

// SystemFontDirs returns the system font directories for the current platform.
func SystemFontDirs() []string {
	switch runtime.GOOS {
	case "darwin":
		return darwinFontDirs()
	case "linux":
		return linuxFontDirs()
	case "windows":
		return windowsFontDirs()
	default:
		return nil
	}
}

// darwinFontDirs returns macOS font directories.
func darwinFontDirs() []string {
	dirs := []string{
		"/System/Library/Fonts",
		"/Library/Fonts",
	}

	// Add user font directory
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Library", "Fonts"))
	}

	return filterExistingDirs(dirs)
}

// linuxFontDirs returns Linux font directories.
func linuxFontDirs() []string {
	dirs := []string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
	}

	// Add user font directories
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs,
			filepath.Join(home, ".fonts"),
			filepath.Join(home, ".local", "share", "fonts"),
		)
	}

	// Add XDG data dirs
	if xdgDataDirs := os.Getenv("XDG_DATA_DIRS"); xdgDataDirs != "" {
		for _, dir := range filepath.SplitList(xdgDataDirs) {
			dirs = append(dirs, filepath.Join(dir, "fonts"))
		}
	}

	return filterExistingDirs(dirs)
}

// windowsFontDirs returns Windows font directories.
func windowsFontDirs() []string {
	dirs := []string{}

	// System fonts directory
	if winDir := os.Getenv("WINDIR"); winDir != "" {
		dirs = append(dirs, filepath.Join(winDir, "Fonts"))
	} else {
		dirs = append(dirs, `C:\Windows\Fonts`)
	}

	// User fonts (Windows 10+)
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		dirs = append(dirs, filepath.Join(localAppData, "Microsoft", "Windows", "Fonts"))
	}

	return filterExistingDirs(dirs)
}

// filterExistingDirs returns only directories that exist.
func filterExistingDirs(dirs []string) []string {
	existing := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			existing = append(existing, dir)
		}
	}
	return existing
}

// DiscoverFonts discovers all fonts in the given directories.
// It walks directories recursively and loads all font files found.
func DiscoverFonts(dirs []string) ([]*Font, error) {
	var fonts []*Font
	seen := make(map[string]bool) // Track seen paths to avoid duplicates

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Skip inaccessible directories
				return nil
			}

			if info.IsDir() {
				return nil
			}

			// Skip if already processed
			if seen[path] {
				return nil
			}
			seen[path] = true

			if !IsFontFile(path) {
				return nil
			}

			loaded, err := LoadFromFile(path)
			if err != nil {
				// Skip fonts that fail to load
				return nil
			}

			fonts = append(fonts, loaded...)
			return nil
		})
		if err != nil {
			// Continue with other directories
			continue
		}
	}

	return fonts, nil
}

// DiscoverSystemFonts discovers all fonts in system font directories.
func DiscoverSystemFonts() ([]*Font, error) {
	return DiscoverFonts(SystemFontDirs())
}
