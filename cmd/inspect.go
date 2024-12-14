package cmd

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

const semlinkXattrKey = "user.semlink.tags"

func init() {
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect file semlink data",
	Long:  `List all xattr data starting with 'user.semlink' for the given file or directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		displaySemlinkXAttrs(filePath)
	},
}

func parseTags(tagString string) []string {
	if tagString == "" {
		return []string{}
	}
	// Split the string by comma and trim spaces
	tags := strings.Split(tagString, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}

func displaySemlinkXAttrs(path string) {
	// Allocate a buffer to retrieve the value
	value := make([]byte, 1024)
	vLen, err := unix.Getxattr(path, semlinkXattrKey, value)
	if err != nil {
		if err == unix.ENODATA {
			fmt.Printf("No tags found for: %s\n", path)
			return
		}
		log.Fatalf("Failed to get xattr value: %v", err)
	}

	rawValue := string(value[:vLen])
	fmt.Printf("Raw xattr value: %s\n", rawValue)

	fmt.Println("Parsed tags:")
	tags := parseTags(rawValue)
	if len(tags) == 0 {
		fmt.Println("  No tags found")
		return
	}

	for _, tag := range tags {
		fmt.Printf("%s\n", tag)
	}
}
