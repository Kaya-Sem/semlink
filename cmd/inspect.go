package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

func init() {
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect folder semlink data",
	Long:  `List all xattr data starting with 'user.semlink' for the given directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		displaySemlinkXAttrs(filePath)
	},
}

func displaySemlinkXAttrs(path string) {
	ensureHasType(path)

	var stat unix.Stat_t
	if err := unix.Stat(path, &stat); err != nil {
		fmt.Printf("Failed to stat file: %v\n", err)
		return
	}

	inode := stat.Ino
	fmt.Printf("Inode: %d\n", inode)

	tags := getSemlinkTags(path)

	fmt.Println("Parsed tags:")
	if len(tags) == 0 {
		fmt.Println("  No tags found")
	}

	for _, tag := range tags {
		fmt.Printf("%s\n", tag)
	}

	fmt.Printf("type: %s\n", getSemlinkType(path))
}
