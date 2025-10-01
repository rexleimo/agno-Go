package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	ft := New()
	if ft == nil {
		t.Fatal("New() returned nil")
	}

	funcs := ft.Functions()
	expectedFuncs := []string{"read_file", "write_file", "list_files", "delete_file", "file_exists"}

	for _, name := range expectedFuncs {
		if _, exists := funcs[name]; !exists {
			t.Errorf("Expected function %s not found", name)
		}
	}
}

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	// Create test file
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ft := New()
	result, err := ft.readFile(context.Background(), map[string]interface{}{
		"path": testFile,
	})

	if err != nil {
		t.Errorf("readFile() error = %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["content"] != testContent {
		t.Errorf("readFile() content = %v, want %v", resultMap["content"], testContent)
	}
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir", "test.txt")
	testContent := "Test content"

	ft := New()
	result, err := ft.writeFile(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": testContent,
	})

	if err != nil {
		t.Errorf("writeFile() error = %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Error("writeFile() success = false")
	}

	// Verify file was written
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to verify written file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("File content = %v, want %v", string(content), testContent)
	}
}

func TestListFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, name := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	ft := New()
	result, err := ft.listFiles(context.Background(), map[string]interface{}{
		"path": tmpDir,
	})

	if err != nil {
		t.Errorf("listFiles() error = %v", err)
	}

	resultMap := result.(map[string]interface{})
	files := resultMap["files"].([]map[string]interface{})

	if len(files) != len(testFiles) {
		t.Errorf("listFiles() count = %v, want %v", len(files), len(testFiles))
	}
}

func TestDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ft := New()
	result, err := ft.deleteFile(context.Background(), map[string]interface{}{
		"path": testFile,
	})

	if err != nil {
		t.Errorf("deleteFile() error = %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Error("deleteFile() success = false")
	}

	// Verify file was deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File was not deleted")
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	ft := New()

	// Test non-existent file
	result, err := ft.fileExists(context.Background(), map[string]interface{}{
		"path": testFile,
	})
	if err != nil {
		t.Errorf("fileExists() error = %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["exists"].(bool) {
		t.Error("fileExists() = true for non-existent file")
	}

	// Create file and test again
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err = ft.fileExists(context.Background(), map[string]interface{}{
		"path": testFile,
	})
	if err != nil {
		t.Errorf("fileExists() error = %v", err)
	}

	resultMap = result.(map[string]interface{})
	if !resultMap["exists"].(bool) {
		t.Error("fileExists() = false for existing file")
	}
}

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()
	ft := NewWithBaseDir(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path inside base dir",
			path:    filepath.Join(tmpDir, "test.txt"),
			wantErr: false,
		},
		{
			name:    "valid relative path",
			path:    filepath.Join(tmpDir, "subdir", "test.txt"),
			wantErr: false,
		},
		{
			name:    "invalid path outside base dir",
			path:    "/tmp/test.txt",
			wantErr: true,
		},
		{
			name:    "invalid path with parent directory",
			path:    filepath.Join(tmpDir, "..", "test.txt"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ft.validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
