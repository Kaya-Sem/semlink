package cmd

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/unix"

	"github.com/Kaya-Sem/semlink/cmd/repository"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(scrubCmd)
}

var (
	allFlag bool
)

var scrubCmd = &cobra.Command{
	Use:   "scrub path",
	Short: "Remove semlink tags from a directory",
	Long:  `Remove the user.semlink tags from a directory`,
	Args:  cobra.ExactArgs(1),
	Run:   runScrub,
}

func init() {
	scrubCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Remove all semlink xattr data, including type information and database")
}

//  TODO: for each folder you want to scrub, if its virtual, unmount everything in it

func runScrub(cmd *cobra.Command, args []string) {
	ensureIsPrivileged()

	path := args[0]

	// Remove the user.semlink xattr
	err := unix.Removexattr(path, semlinkTagXattrKey)
	if err != nil && err != unix.ENODATA {
		log.Fatalf("Failed to remove xattr: %v", err)
	} else if err == unix.ENODATA {
		fmt.Printf("No semlink data found for %s\n", path)
	} else {
		fmt.Printf("Successfully removed semlink data for %s\n", path)
	}

	// If --all flag is set, remove the semlink.type xattr as well
	if allFlag {

		// Get inode for the file info
		var stat unix.Stat_t
		if err := unix.Stat(path, &stat); err != nil {
			log.Fatalf("Failed to stat file: %v", err)
		}

		inode := stat.Ino

		repo, err := repository.NewSqliteRepo()
		if err != nil {
			log.Fatalf("%v", err)
		}

		// Remove from registry
		if err := repo.RemoveFolder(repository.FolderInfo{Inode: inode}); err != nil {
			log.Fatalf("Failed to remove entry from database: %v", err)
			os.Exit(1)
		}

		// TODO: remove tags from folder_tags

		fmt.Printf("Successfully removed database entry for %s\n", path)

		err = unix.Removexattr(path, semlinkTypeXattrKey)
		if err != nil && err != unix.ENODATA {
			log.Fatalf("Failed to remove type xattr: %v", err)
		} else if err == unix.ENODATA {
			fmt.Printf("No type data found for %s\n", path)
		} else {
			fmt.Printf("Successfully removed type data for %s\n", path)

		}
	}

	triggerUpdate()
}
