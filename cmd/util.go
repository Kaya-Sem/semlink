package cmd

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

func triggerUpdate() {

	fmt.Printf("\nUpdate triggered!\n")

	/*  TODO: before mounting, attempt a repair (system Inode scan) */

	// resolve_orphans()
	mountDirectories()
}

func mountDirectories() {
	registry, err := loadRegistry()
	if err != nil {
		log.Fatalf("Failed to load registry: %v", err)
	}

	sourceMap := make(map[string][]string)
	receiverMap := make(map[string][]string)

	/* TODO: before mounting, attempt a repair (system Inode scan) */

	for _, folder := range registry.TaggedFiles {
		folderType := getSemlinkType(folder.FullPath)
		tags := getSemlinkTags(folder.FullPath)

		if folderType == "receiver" {
			for _, tag := range tags {
				// Check if the tag is already a key in the receiverMap
				if _, ok := receiverMap[tag]; !ok {
					// If not, initialize it with a new slice
					receiverMap[tag] = []string{}
				}
				// Add the folder path to the slice
				receiverMap[tag] = append(receiverMap[tag], folder.FullPath)
			}
		} else if folderType == "source" {
			for _, tag := range tags {
				// Check if the tag is already a key in the sourceMap
				if _, ok := sourceMap[tag]; !ok {
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
