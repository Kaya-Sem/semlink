package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/Kaya-Sem/oopsie"
	"golang.org/x/sys/unix"
)

type FolderInfo struct {
	Inode    uint64 `json:"inode"`
	FullPath string `json:"full_path"`
}

//  TODO: add a command to trigger an update manually -> users can run it at startup to mount everything

func triggerUpdate() {

	fmt.Printf("\nUpdate triggered!\n")

	/*  TODO: before mounting, attempt a repair (system Inode scan) */

	// resolve_orphans()
	// unmount_all()
	mountDirectories()
}

func mountDirectories() {
	registry, err := loadRegistry()
	if err != nil {
		log.Fatalf("Failed to load registry: %v", err)
	}

	sourceMap := make(map[string][]string)
	receiverMap := make(map[string][]string)

	for _, folder := range registry.TaggedFiles {
		folderType := getSemlinkType(folder.FullPath)
		tags := getSemlinkTags(folder.FullPath)

		if folderType == "receiver" {
			for _, tag := range tags {
				if _, ok := receiverMap[tag]; !ok { // tag present?
					// If not, initialize it with a new slice
					receiverMap[tag] = []string{}
				}
				// Add the folder path to the slice
				receiverMap[tag] = append(receiverMap[tag], folder.FullPath)
			}
		} else if folderType == "source" {
			for _, tag := range tags {
				if _, ok := sourceMap[tag]; !ok { // tag present?
					// If not, initialize it with a new slice
					sourceMap[tag] = []string{}
				}
				// Add the folder path to the slice
				sourceMap[tag] = append(sourceMap[tag], folder.FullPath)
			}
		} else {
			log.Fatalf("Unexpected type: %s for path %s", folderType, folder.FullPath)
		}
	}

	// Example: Print the maps to verify the results
	fmt.Println("Receiver Map:", receiverMap)
	fmt.Println("Source Map:", sourceMap)

	for tag, fileInfoSlice := range sourceMap {
		targets := receiverMap[tag]

		for _, sourcePath := range fileInfoSlice {
			for _, targetPath := range targets {
				err := linkFolder(sourcePath, targetPath)

				if err != nil {
					fmt.Printf("Error: %v", err)

				}

			}
		}

	}

}

func logFatalWithCaller(msg string, err error) {
	// Get caller details
	pc, file, line, ok := runtime.Caller(1) // Caller(1) means one level up in the call stack
	if !ok {
		log.Fatalf("Failed to retrieve caller info: %v", err)
	}

	// Extract function name from Program Counter (PC)
	funcName := runtime.FuncForPC(pc).Name()
	funcName = trimFunctionName(funcName) // Trim package/module prefixes for readability

	// Log with caller details
	log.Fatalf("[ERROR] %s:%d [%s] %s: %v", file, line, funcName, msg, err)
}

// Helper function to simplify function names (removes package path)
func trimFunctionName(funcName string) string {
	if idx := strings.LastIndex(funcName, "/"); idx != -1 {
		return funcName[idx+1:]
	}
	return funcName
}

// linkFolder creates a subdirectory in the target path and binds the source folder to it.
func linkFolder(source string, target string) error {
	// Extract the last piece of the target path (the last folder name)
	subDir := path.Join(target, path.Base(source))

	// Create the subdirectory in the target folder
	err := os.MkdirAll(subDir, 0755) // Create target subdirectory if it doesn't exist
	if err != nil {
		return fmt.Errorf("failed to create subdirectory %s: %w", subDir, err)
	}

	err = setType(subDir, VIRTUAL)

	if err != nil {
		return fmt.Errorf("failed to set type %s on %s", VIRTUAL, subDir)
	}

	// Bind mount the source folder to the subdirectory in the target folder
	err = unix.Mount(source, subDir, "", unix.MS_BIND, "")
	if err != nil {
		return fmt.Errorf("failed to bind mount %s to %s: %w", source, subDir, err)
	}

	return nil
}

func isPrivileged() bool {
	return os.Geteuid() == 0
}

func ensureIsPrivileged() {
	if !isPrivileged() {
		fmt.Print(oopsie.CreateOopsie().Title("Invalid Permissions").Error(fmt.Errorf("semlink needs privileges to function. Please run with sudo, doas or as root.")).Render())
		os.Exit(1)
	}
}
