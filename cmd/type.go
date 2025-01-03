package cmd

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var verbose bool

func init() {
	typeCmd := &cobra.Command{
		Use:   "type [flags] path",
		Short: "Set the type of a file or directory",
		Long:  `Set the type of a file or directory in the semlink xattr data. The type can be either 'source' or 'receiver'.`,
		Args:  cobra.ExactArgs(2),
		Run:   runType,
	}

	typeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.AddCommand(typeCmd)
}

func runType(cmd *cobra.Command, args []string) {
	path := args[1]
	typeArg := args[0]

	if typeArg != "source" && typeArg != "receiver" {
		log.Fatalf("Invalid type specified: %s. Must be 'source' or 'receiver'.", typeArg)
	}

	// Read existing xattr
	value := make([]byte, 1024)
	vLen, err := unix.Getxattr(path, semlinkXattrKey, value)

	var existingTags []string
	if err == nil {
		existingTags = strings.Split(string(value[:vLen]), ",")
	} else if err != unix.ENODATA {
		log.Fatalf("Failed to read existing xattr: %v", err)
	}

	// Remove any existing type tags
	filteredTags := []string{}
	for _, tag := range existingTags {
		if !strings.HasPrefix(tag, "type=") {
			filteredTags = append(filteredTags, tag)
		}
	}

	// Add the new type tag
	typeTag := fmt.Sprintf("type=%s", typeArg)
	filteredTags = append(filteredTags, typeTag)

	// Create the new xattr value
	newTagString := strings.Join(filteredTags, ",")

	// Set the updated xattr value
	err = unix.Setxattr(path, semlinkXattrKey, []byte(newTagString), 0)
	if err != nil {
		log.Fatalf("Failed to set xattr: %v", err)
	}

	if verbose {
		fmt.Printf("Successfully updated type for %s\n", path)
		fmt.Printf("New xattr data: %s\n", newTagString)
	}
}
