package cmd

import (
	"fmt"
	"log"
)

func triggerUpdate() {

	fmt.Printf("\nUpdate triggered!\n")

	mountDirectories()
}

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
