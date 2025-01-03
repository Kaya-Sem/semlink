package cmd

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var tags []string

func init() {
	addCmd := &cobra.Command{
		Use:   "add [flags] path",
		Short: "Add tags to a file or directory",
		Long:  `Add one or more tags to a file or directory's semlink xattr data.`,
		Args:  cobra.ExactArgs(1),
		Run:   runAdd,
	}

	addCmd.Flags().StringSliceVarP(&tags, "tag", "t", []string{}, "Tags to add (can be specified multiple times)")
	addCmd.MarkFlagRequired("tag")

	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) {
	path := args[0]

	// Load registry
	registry, err := loadRegistry()
	if err != nil {
		log.Fatalf("Failed to load registry: %v", err)
	}

	// Read existing tags
	value := make([]byte, 1024)
	vLen, err := unix.Getxattr(path, semlinkXattrKey, value)

	var existingTags []string
	if err == nil {
		existingTags = parseTags(string(value[:vLen]))
	} else if err != unix.ENODATA {
		log.Fatalf("Failed to read existing tags: %v", err)
	}

	// Combine existing and new tags, removing duplicates
	tagMap := make(map[string]bool)
	for _, tag := range existingTags {
		tagMap[tag] = true
	}
	for _, tag := range tags {
		tagMap[tag] = true
	}

	// Convert back to slice
	var allTags []string
	for tag := range tagMap {
		if tag != "" {
			allTags = append(allTags, tag)
		}
	}

	// Create the new tag string
	newTagString := strings.Join(allTags, ",")

	// Set the new xattr value
	err = unix.Setxattr(path, semlinkXattrKey, []byte(newTagString), 0)
	if err != nil {
		log.Fatalf("Failed to set xattr: %v", err)
	}

	// Get inode for the file info
	var stat unix.Stat_t
	if err := unix.Stat(path, &stat); err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}

	inode := stat.Ino

	// Update registry
	if err := registry.updateFile(inode, path); err != nil {
		log.Fatalf("Failed to update registry: %v", err)
	}

	fmt.Printf("Successfully updated tags for %s\n", path)
	fmt.Printf("New tags: %s\n", newTagString)
}
