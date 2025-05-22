package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"log"
)

func init() {
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect [path...]",
	Short: "Inspect folder semlink data",
	Long:  `List all semlink xattr data for the specified directories. Supports multiple paths.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, path := range args {
			if len(args) > 1 {
				fmt.Printf("\n=== %s ===\n", path)
			}
			displaySemlinkXAttrs(path)
		}
	},
}

func displaySemlinkXAttrs(path string) {
	var stat unix.Stat_t
	if err := unix.Stat(path, &stat); err != nil {
		fmt.Printf("Failed to stat file: %v\n", err)
		return
	}

	inode := stat.Ino

	folderType, err := getSemlinkType(path)
	if err != nil {
		log.Printf("Error getting semlink type for %s: %v", path, err)
		return
	}

	if folderType == "" {
		folderType = "no type found. Consider setting a type or scrubbing the folder tags"
	}

	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Inode: %d\n", inode)
	fmt.Printf("Type: %s\n", folderType)

	tags, err := getSemlinkTags(path)
	if err != nil {
		log.Printf("Error getting semlink tags for %s: %v", path, err)
		return
	}

	if len(tags) == 0 {
		fmt.Println("No tags found")
		return
	}

	fmt.Println("Parsed tags:")
	for _, tag := range tags {
		fmt.Printf("  %s\n", tag)
	}
}
