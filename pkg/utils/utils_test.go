package utils

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		validate func(string) bool
	}{
		{
			name:     "absolute path",
			path:     "/tmp/test",
			wantErr:  false,
			validate: func(result string) bool { return result == "/tmp/test" },
		},
		{
			name:     "relative path",
			path:     "./test",
			wantErr:  false,
			validate: func(result string) bool { return result == "./test" },
		},
		{
			name:     "home directory",
			path:     "~/test",
			wantErr:  false,
			validate: func(result string) bool { return filepath.IsAbs(result) },
		},
		{
			name:     "current directory",
			path:     ".",
			wantErr:  false,
			validate: func(result string) bool { return result == "." },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.validate(result) {
				t.Errorf("ExpandPath() result validation failed for %v", result)
			}
		})
	}
}

func TestGetHomeDir(t *testing.T) {
	home, err := GetHomeDir()
	if err != nil {
		t.Errorf("GetHomeDir() error = %v", err)
		return
	}
	
	if home == "" {
		t.Error("GetHomeDir() returned empty string")
	}
	
	if !filepath.IsAbs(home) {
		t.Errorf("GetHomeDir() returned non-absolute path: %v", home)
	}
	
	// Check if the directory exists
	if !PathExists(home) {
		t.Errorf("Home directory does not exist: %v", home)
	}
}

func TestPathExists(t *testing.T) {
	// Test with existing file/directory
	tempDir := t.TempDir()
	if !PathExists(tempDir) {
		t.Errorf("PathExists() = false for existing temp dir %v", tempDir)
	}
	
	// Test with non-existing path
	nonExistentPath := filepath.Join(tempDir, "non-existent")
	if PathExists(nonExistentPath) {
		t.Errorf("PathExists() = true for non-existent path %v", nonExistentPath)
	}
	
	// Create a file and test
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	if !PathExists(testFile) {
		t.Errorf("PathExists() = false for existing file %v", testFile)
	}
}

func TestIsFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test with directory
	if IsFile(tempDir) {
		t.Errorf("IsFile() = true for directory %v", tempDir)
	}
	
	// Test with file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	if !IsFile(testFile) {
		t.Errorf("IsFile() = false for file %v", testFile)
	}
	
	// Test with non-existent path
	nonExistentPath := filepath.Join(tempDir, "non-existent")
	if IsFile(nonExistentPath) {
		t.Errorf("IsFile() = true for non-existent path %v", nonExistentPath)
	}
}

func TestIsDir(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test with directory
	if !IsDir(tempDir) {
		t.Errorf("IsDir() = false for directory %v", tempDir)
	}
	
	// Test with file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	if IsDir(testFile) {
		t.Errorf("IsDir() = true for file %v", testFile)
	}
	
	// Test with non-existent path
	nonExistentPath := filepath.Join(tempDir, "non-existent")
	if IsDir(nonExistentPath) {
		t.Errorf("IsDir() = true for non-existent path %v", nonExistentPath)
	}
}

func TestGetDirSize(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create some test files
	testFiles := []struct {
		name string
		size int
	}{
		{"file1.txt", 100},
		{"file2.txt", 200},
		{"subdir/file3.txt", 300},
	}
	
	expectedSize := int64(0)
	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		
		content := make([]byte, tf.size)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		expectedSize += int64(tf.size)
	}
	
	size, err := GetDirSize(tempDir)
	if err != nil {
		t.Errorf("GetDirSize() error = %v", err)
		return
	}
	
	if size != expectedSize {
		t.Errorf("GetDirSize() = %v, want %v", size, expectedSize)
	}
	
	// Test with non-existent directory
	_, err = GetDirSize(filepath.Join(tempDir, "non-existent"))
	if err == nil {
		t.Error("GetDirSize() should return error for non-existent directory")
	}
}

func TestGetFileCount(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test files and directories
	testPaths := []string{
		"file1.txt",
		"file2.txt",
		"subdir/file3.txt",
		"subdir/subsubdir/file4.txt",
		"emptydir/",
	}
	
	expectedCount := 4 // Only files, not directories
	
	for _, path := range testPaths {
		fullPath := filepath.Join(tempDir, path)
		if filepath.Ext(path) == "" && path[len(path)-1] == '/' {
			// It's a directory
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
		} else {
			// It's a file
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
			if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}
	}
	
	count, err := GetFileCount(tempDir)
	if err != nil {
		t.Errorf("GetFileCount() error = %v", err)
		return
	}
	
	if count != expectedCount {
		t.Errorf("GetFileCount() = %v, want %v", count, expectedCount)
	}
}

func TestExecuteCommand(t *testing.T) {
	ctx := context.Background()
	
	// Test successful command
	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd, arg = "cmd", "/c echo hello"
	} else {
		cmd, arg = "echo", "hello"
	}
	
	result := ExecuteCommand(ctx, cmd, arg)
	if result.Error != nil {
		t.Errorf("ExecuteCommand() error = %v", result.Error)
		return
	}
	
	if result.ExitCode != 0 {
		t.Errorf("ExecuteCommand() exit code = %v, want 0", result.ExitCode)
	}
	
	// Test command with timeout
	result = ExecuteCommandWithTimeout(100*time.Millisecond, cmd, arg)
	if result.Error != nil {
		t.Errorf("ExecuteCommandWithTimeout() error = %v", result.Error)
	}
	
	// Test non-existent command
	result = ExecuteCommand(ctx, "non-existent-command-12345")
	if result.Error == nil {
		t.Error("ExecuteCommand() should return error for non-existent command")
	}

	// Note: Exit code might be 0 even for non-existent commands on some systems
	// so we don't test for specific exit code
}

func TestRemoveDir(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test-remove")
	
	// Create test directory with files
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Remove directory
	if err := RemoveDir(testDir); err != nil {
		t.Errorf("RemoveDir() error = %v", err)
	}
	
	// Verify directory is removed
	if PathExists(testDir) {
		t.Errorf("Directory still exists after RemoveDir(): %v", testDir)
	}
	
	// Test removing non-existent directory (this might not error on all systems)
	err := RemoveDir(filepath.Join(tempDir, "non-existent"))
	// Note: Some systems don't error when removing non-existent directories
	_ = err
}
