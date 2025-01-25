package cmd

import (
	"fmt"
	"log"

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

func isValidType(typeArg string) bool {
	return (typeArg == RECEIVER) || (typeArg == VIRTUAL) || (typeArg == SOURCE)
}

func setType(path, typeArg string) error {
	ensureIsPrivileged()

	//  TODO: create a validator function for this.
	if isValidType(typeArg) {
		return fmt.Errorf("invalid type specified: %s. Must be 'source' or 'receiver'", typeArg)
	}

	setXattr(path, semlinkTypeXattrKey, typeArg)

	return nil
}

func runType(cmd *cobra.Command, args []string) {
	typeArg := args[0]
	path := args[1]

	err := setType(path, typeArg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if verbose {
		fmt.Printf("Successfully updated type for %s\n", path)
		fmt.Printf("New xattr data: type=%s\n", typeArg)
	}

	triggerUpdate()
}
