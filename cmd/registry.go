package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type FileInfo struct {
	Inode    uint64 `json:"inode"`
	FullPath string `json:"full_path"`
}

type Registry struct {
	TaggedFiles []FileInfo `json:"tagged_files"`
}

// TODO: trigger update function.

func mountDirectories() {
	registry, err := loadRegistry()

	if err != nil {
		log.Fatalf("Failed to load registry: %v", err)
	}

	// sourceMap := make(map[string][]string)
	// receiverMap := make(map[string][]string)

	for _, folder := range registry.TaggedFiles {
		fmt.Println(folder.FullPath)
	}

}

func getRegistryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	registryDir := filepath.Join(home, registryDir)
	return filepath.Join(registryDir, registryFile), nil
}

func loadRegistry() (*Registry, error) {
	registryPath, err := getRegistryPath()
	if err != nil {
		return nil, err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(registryPath), registryPermissions); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(registryPath)
	if os.IsNotExist(err) {
		// Return empty registry if file doesn't exist
		return &Registry{TaggedFiles: []FileInfo{}}, nil
	} else if err != nil {
		return nil, err
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	return &registry, nil
}

func (r *Registry) save() error {
	registryPath, err := getRegistryPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(registryPath, data, 0644)
}

func (r *Registry) updateFile(inode uint64, path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check if the file already exists in the registry
	for i, file := range r.TaggedFiles {
		if file.Inode == inode {
			// Update existing entry
			r.TaggedFiles[i].FullPath = absPath
			return r.save()
		}
	}

	// Add a new entry if it doesn't exist
	r.TaggedFiles = append(r.TaggedFiles, FileInfo{
		Inode:    inode,
		FullPath: absPath,
	})

	return r.save()
}

// New list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tagged files",
	Long:  `List all files that have semlink tags, along with their tags.`,
	Run: func(cmd *cobra.Command, args []string) {
		registry, err := loadRegistry()
		if err != nil {
			log.Fatalf("Failed to load registry: %v", err)
		}

		if len(registry.TaggedFiles) == 0 {
			fmt.Println("No tagged files found.")
			return
		}

		for _, file := range registry.TaggedFiles {
			fmt.Printf("- Inode: %d\n  Full Path: %s\n", file.Inode, file.FullPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

// removeFile removes the entry with the given inode from the registry
func (r *Registry) removeFile(inode uint64) error {
	// Find the index of the file with the given inode
	var indexToRemove = -1
	for i, file := range r.TaggedFiles {
		if file.Inode == inode {
			indexToRemove = i
			break
		}
	}

	// If the file is not found, return an error
	if indexToRemove == -1 {
		return fmt.Errorf("file with inode %d not found in registry", inode)
	}

	// Remove the file entry from the TaggedFiles slice
	r.TaggedFiles = append(r.TaggedFiles[:indexToRemove], r.TaggedFiles[indexToRemove+1:]...)

	// Save the updated registry
	return r.save()
}
