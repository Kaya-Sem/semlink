package cmd

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

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

func displaySemlinkXAttrs(path string) {
	// Allocate a buffer to retrieve xattr keys
	buf := make([]byte, 1024)
	n, err := unix.Listxattr(path, buf)
	if err != nil {
		log.Fatalf("Failed to list xattr keys: %v", err)
	}
	keys := parseXattrKeys(buf[:n])

	// Filter and print keys that start with 'user.semlink'
	fmt.Printf("Inspecting file: %s\n", path)
	fmt.Println("Matching xattr keys:")
	found := false
	for _, key := range keys {
		if strings.HasPrefix(key, "user.semlink") {
			found = true
			// Allocate a buffer to retrieve the value
			value := make([]byte, 1024)
			vLen, err := unix.Getxattr(path, key, value)
			if err != nil {
				log.Printf("Failed to get xattr value for %s: %v\n", key, err)
				continue
			}
			fmt.Printf("  %s: %s\n", key, string(value[:vLen]))
		}
	}

	if !found {
		fmt.Println("  No xattr keys found starting with 'user.semlink'.")
	}
}

func parseXattrKeys(data []byte) []string {
	keys := []string{}
	start := 0
	for i, b := range data {
		if b == 0 {
			keys = append(keys, string(data[start:i]))
			start = i + 1
		}
	}
	return keys
}
