package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func BenchmarkExpandPath(b *testing.B) {
	testPaths := []string{
		"/tmp/test",
		"./test",
		"~/test",
		".",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			_, _ = ExpandPath(path)
		}
	}
}

func BenchmarkPathExists(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	paths := []string{
		tempDir,
		testFile,
		filepath.Join(tempDir, "non-existent"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = PathExists(path)
		}
	}
}

func BenchmarkIsFile(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	paths := []string{
		tempDir,
		testFile,
		filepath.Join(tempDir, "non-existent"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = IsFile(path)
		}
	}
}

func BenchmarkIsDir(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	paths := []string{
		tempDir,
		testFile,
		filepath.Join(tempDir, "non-existent"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = IsDir(path)
		}
	}
}

func BenchmarkGetDirSize(b *testing.B) {
	tempDir := b.TempDir()
	
	// Create test files
	for i := 0; i < 10; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("test%d.txt", i))
		content := make([]byte, 1024) // 1KB files
		os.WriteFile(testFile, content, 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetDirSize(tempDir)
	}
}

func BenchmarkGetFileCount(b *testing.B) {
	tempDir := b.TempDir()
	
	// Create test files
	for i := 0; i < 10; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("test%d.txt", i))
		os.WriteFile(testFile, []byte("test"), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetFileCount(tempDir)
	}
}

func BenchmarkExecuteCommand(b *testing.B) {
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if runtime.GOOS == "windows" {
			_ = ExecuteCommand(ctx, "cmd", "/c", "echo", "test")
		} else {
			_ = ExecuteCommand(ctx, "echo", "test")
		}
	}
}

// Benchmark concurrent operations
func BenchmarkConcurrentPathExists(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = PathExists(testFile)
		}
	})
}

func BenchmarkConcurrentExpandPath(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = ExpandPath("~/test")
		}
	})
}

// Memory allocation benchmarks
func BenchmarkExpandPathAllocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExpandPath("~/test/path")
	}
}

func BenchmarkGetDirSizeAllocs(b *testing.B) {
	tempDir := b.TempDir()
	
	// Create test files
	for i := 0; i < 5; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("test%d.txt", i))
		os.WriteFile(testFile, []byte("test"), 0644)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetDirSize(tempDir)
	}
}
