package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

func init() {
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect folder semlink data",
	Long:  `List all semlink xattr data for the specified directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		displaySemlinkXAttrs(filePath)
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
		log.Fatalf("Error: %v", err)
	}

	if folderType == "" {
		folderType = "no type found. Consider setting a type or scrubbing the folder tags"
	}

	fmt.Printf("Inode: %d\n", inode)
	fmt.Printf("Type: %s\n", folderType)

	tags, err := getSemlinkTags(path)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(tags) == 0 {
		fmt.Println("No tags found")
		return
	}

	fmt.Println("Parsed tags:")

	for _, tag := range tags {
		fmt.Printf("%s\n", tag)
	}

}
