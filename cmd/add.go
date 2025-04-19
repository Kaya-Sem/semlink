package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/Kaya-Sem/oopsie"
	"github.com/Kaya-Sem/semlink/cmd/repository"
	"github.com/spf13/cobra"
)

var tags []string

func init() {
	addCmd := &cobra.Command{
		Use:   "add [flags] path",
		Short: "Add tags to a directory",
		Long:  `Add one or more tags to a directory's semlink xattr data.`,
		Args:  cobra.ExactArgs(1),
		Run:   runAdd,
	}

	addCmd.Flags().StringSliceVarP(&tags, "tag", "t", []string{}, "Tags to add (can be specified multiple times)")
	addCmd.MarkFlagRequired("tag")

	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) {

	ensureIsPrivileged()

	rawPath := args[0]
	path, err := filepath.Abs(rawPath)
	if err != nil {
		log.Fatalf("Failed to resolve absolute path: %v", err)
	}

	ensureHasType(path)

	existingTags, err := getSemlinkTags(path)
	if err != nil {
		log.Fatalf("%v", err)
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

	setXattr(path, semlinkTagXattrKey, newTagString)

	// Get inode for the file info
	var stat unix.Stat_t
	if err := unix.Stat(path, &stat); err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}

	inode := stat.Ino

	repo, err := repository.NewSqliteRepo()
	if err != nil {
		fmt.Print(oopsie.CreateOopsie().Title("Database error").Error(err).IndicatorMessage("SQL").Render())
		os.Exit(1)
	}

	folder := repository.FolderInfo{Inode: inode, FullPath: path}

	err = repo.AddFolder(folder)
	if err != nil {
		fmt.Print(oopsie.CreateOopsie().Title("Database error").Error(err).IndicatorMessage("SQL").Render())
	}

	err = repo.AddTagsToFolder(folder, allTags)
	if err != nil {
		log.Fatalf("Could not add tags to folder %s in the database: %v", folder.FullPath, err)
	}

	if verbose {
		fmt.Printf("Successfully updated tags for %s\n", path)
		fmt.Printf("New tags: %s\n", newTagString)
	}

	triggerUpdate()
}

func ensureHasType(path string) {
	folderType, err := getSemlinkType(path)
	if err != nil {
		log.Fatalf("Coud not get type for %s", path)
	}

	if !isValidType(Type(folderType)) {
		log.Printf("Invalid type found, replaced with %s", defaultType)
		setType(path, defaultType)
	}
}
