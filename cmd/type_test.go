package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestIsValidType(t *testing.T) {
	tests := []struct {
		name     string
		typeArg  Type
		expected bool
	}{
		{"Valid Receiver", RECEIVER, true},
		{"Valid Virtual", VIRTUAL, true},
		{"Valid Source", SOURCE, true},
		{"Invalid Type", Type("invalid"), false},
		{"Empty Type", Type(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidType(tt.typeArg)
			if result != tt.expected {
				t.Errorf("isValidType(%v) = %v, want %v", tt.typeArg, result, tt.expected)
			}
		})
	}
}

func TestIsDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Valid Directory", tempDir, true},
		{"Valid File", tempFile.Name(), false},
		{"Non-existent Path", "/nonexistent/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDirectory(tt.path)
			if result != tt.expected {
				t.Errorf("isDirectory(%v) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestSetType(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		path        string
		typeArg     Type
		expectError bool
	}{
		{"Valid Source Type", tempDir, SOURCE, false},
		{"Valid Receiver Type", tempDir, RECEIVER, false},
		{"Invalid Type", tempDir, Type("invalid"), true},
		{"Non-existent Path", "/nonexistent/path", SOURCE, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setType(tt.path, tt.typeArg)
			if (err != nil) != tt.expectError {
				t.Errorf("setType(%v, %v) error = %v, wantErr %v", tt.path, tt.typeArg, err, tt.expectError)
			}

			if !tt.expectError {
				value, err := getXattr(tt.path, semlinkTypeXattrKey)
				if err != nil {
					t.Errorf("Failed to get xattr: %v", err)
				}
				if string(value) != string(tt.typeArg) {
					t.Errorf("xattr value = %v, want %v", string(value), tt.typeArg)
				}
			}
		})
	}
}

func TestListValidTypes(t *testing.T) {
	t.Run("List Valid Types", func(t *testing.T) {
		oldStdout := os.Stdout

		r, w, _ := os.Pipe()
		os.Stdout = w

		listValidTypes()

		w.Close()

		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		capturedOutput := buf.String()

		if capturedOutput == "" {
			t.Error("Expected some output, got none")
		}

	})
}

// func TestTypeCommand(t *testing.T) {
// 	// Create a temporary directory for testing
// 	tempDir, err := os.MkdirTemp("", "testdir")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp dir: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)
//
// 	tests := []struct {
// 		name        string
// 		args        []string
// 		expectError bool
// 	}{
// 		{"Valid Source Type", []string{"source", tempDir}, false},
// 		{"Valid Receiver Type", []string{"receiver", tempDir}, false},
// 		{"Invalid Type", []string{"invalid", tempDir}, true},
// 		{"Missing Path", []string{"source"}, true},
// 		{"Extra Arguments", []string{"source", tempDir, "extra"}, true},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cmd := typeCmd
// 			cmd.SetArgs(tt.args)
// 			err := cmd.Execute()
// 			if (err != nil) != tt.expectError {
// 				t.Errorf("type command with args %v error = %v, wantErr %v", tt.args, err, tt.expectError)
// 			}
// 		})
// 	}
// }
//
// func TestTypeEdgeCases(t *testing.T) {
// 	// Test with a file instead of a directory
// 	tempFile, err := os.CreateTemp("", "testfile")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp file: %v", err)
// 	}
// 	defer os.Remove(tempFile.Name())
// 	defer tempFile.Close()
//
// 	t.Run("Set Type on File", func(t *testing.T) {
// 		err := setType(tempFile.Name(), SOURCE)
// 		if err == nil {
// 			t.Error("Expected error when setting type on a file")
// 		}
// 	})
//
// 	// Test with a non-existent path
// 	t.Run("Set Type on Non-existent Path", func(t *testing.T) {
// 		err := setType("/nonexistent/path", SOURCE)
// 		if err == nil {
// 			t.Error("Expected error when setting type on non-existent path")
// 		}
// 	})
//
// 	// Test with empty type
// 	t.Run("Set Empty Type", func(t *testing.T) {
// 		tempDir, err := os.MkdirTemp("", "testdir")
// 		if err != nil {
// 			t.Fatalf("Failed to create temp dir: %v", err)
// 		}
// 		defer os.RemoveAll(tempDir)
//
// 		err = setType(tempDir, Type(""))
// 		if err == nil {
// 			t.Error("Expected error when setting empty type")
// 		}
// 	})
// }
