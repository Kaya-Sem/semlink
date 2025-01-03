package cmd

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(scrubCmd)
}

var scrubCmd = &cobra.Command{
	Use:   "scrub path",
	Short: "Remove semlink xattr data and its registry entry",
	Long:  `Remove the user.semlink xattr from a file or directory and delete its entry from the registry.`,
	Args:  cobra.ExactArgs(1),
	Run:   runScrub,
}

func runScrub(cmd *cobra.Command, args []string) {
	path := args[0]

	// Load registry
	registry, err := loadRegistry()
	if err != nil {
		log.Fatalf("Failed to load registry: %v", err)
	}

	// Remove the user.semlink xattr
	err = unix.Removexattr(path, semlinkXattrKey)
	if err != nil && err != unix.ENODATA {
		log.Fatalf("Failed to remove xattr: %v", err)
	} else if err == unix.ENODATA {
		fmt.Printf("No semlink data found for %s\n", path)
	} else {
		fmt.Printf("Successfully removed semlink data for %s\n", path)

		// Get inode for the file info
		var stat unix.Stat_t
		if err := unix.Stat(path, &stat); err != nil {
			log.Fatalf("Failed to stat file: %v", err)
		}

		inode := stat.Ino

		// Remove from registry
		if err := registry.removeFile(inode); err != nil {
			log.Fatalf("Failed to remove entry from registry: %v", err)
		}

		fmt.Printf("Successfully removed registry entry for %s\n", path)
	}
}
