package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetHomeDir returns the user's home directory
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// GetCacheDir returns the cache directory for the current OS
func GetCacheDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("APPDATA"), nil
	case "darwin":
		home, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Caches"), nil
	default: // linux and others
		if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
			return xdgCache, nil
		}
		home, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".cache"), nil
	}
}

// GetConfigDir returns the config directory for the current OS
func GetConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("APPDATA"), nil
	case "darwin":
		home, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	default: // linux and others
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return xdgConfig, nil
		}
		home, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config"), nil
	}
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDir checks if a path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if a path is a file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// GetDirSize calculates the total size of a directory
func GetDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// GetFileCount counts the number of files in a directory
func GetFileCount(path string) (int, error) {
	var count int
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

// NormalizePath normalizes a file path for the current OS
func NormalizePath(path string) string {
	return filepath.Clean(path)
}

// JoinPath joins path elements and normalizes the result
func JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}

// ExpandPath expands ~ to home directory
func ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	
	if path == "~" {
		return home, nil
	}
	
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}
	
	return path, nil
}

// MakeDir creates a directory if it doesn't exist
func MakeDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// RemoveDir removes a directory and all its contents
func RemoveDir(path string) error {
	return os.RemoveAll(path)
}
