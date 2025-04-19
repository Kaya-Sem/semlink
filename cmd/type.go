package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Kaya-Sem/oopsie"
	"github.com/spf13/cobra"
	"slices"
)

type Type string

const (
	RECEIVER Type = "receiver"
	VIRTUAL  Type = "virtual"
	SOURCE   Type = "source"
)

var validTypes = []Type{RECEIVER, VIRTUAL, SOURCE}
var validUserFacingTypes = []Type{RECEIVER, SOURCE} // marking a dir as virtual might mark its death because force removal on nuke

var verbose bool

func init() {
	typeCmd := &cobra.Command{
		Use:   "type",
		Short: "Manage directory types",
		Long:  `Manage the type of a directory in the semlink xattr data.`,
	}

	setCmd := &cobra.Command{
		Use:   "set [flags] type path",
		Short: "Set the type of a  directory",
		Long:  `Set the type of a directory in the semlink xattr data. The type can be either 'source' or 'receiver'.`,
		Args:  cobra.ExactArgs(2),
		Run:   runTypeSet,
	}

	setCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	typeCmd.AddCommand(setCmd)

	// listCmd := &cobra.Command{
	// 	Use:   "list type",
	// 	Short: "List all directories with given type",
	// 	Long:  `List the directories with given type in the semlink xattr data.`,
	// 	Args:  cobra.MaximumNArgs(1),
	// 	Run:   runTypeList,
	// }
	//
	// typeCmd.AddCommand(listCmd)

	rootCmd.AddCommand(typeCmd)
}

func isValidType(typeArg Type) bool {
	return slices.Contains(validTypes, typeArg)
}

// TODO: ensure path is a folder
func setType(path string, typeArg Type) error {
	ensureIsPrivileged()
	if !isDirectory(path) {

		fmt.Print(oopsie.CreateOopsie().Title("Not a folder").IndicatorColors(oopsie.BLACK, oopsie.RED).Error(fmt.Errorf("Only folders are supported in semlink")).Render())
		os.Exit(1)
	}

	if !isValidType(typeArg) {
		validStrs := make([]string, len(validUserFacingTypes))
		for i, t := range validUserFacingTypes {
			validStrs[i] = string(t)
		}
		fmt.Print(oopsie.CreateOopsie().
			Title("Invalid type").
			IndicatorColors(oopsie.BLACK, oopsie.RED).
			Error(fmt.Errorf("Invalid type specified: %s. Must be one of: %v", typeArg, validStrs)).
			Render())
		os.Exit(1)
	}

	setXattr(path, semlinkTypeXattrKey, string(typeArg))

	return nil
}

func runTypeSet(cmd *cobra.Command, args []string) {
	typeArg := args[0]
	path := args[1]

	err := setType(path, Type(typeArg))
	if err != nil {
		log.Fatalf("%v", err)
	}

	if verbose {
		fmt.Printf("Successfully updated type for %s\n", path)
		fmt.Printf("New xattr data: type=%s\n", typeArg)
	}

	triggerUpdate()
}

// func runTypeList(cmd *cobra.Command, args []string) {
// 	if len(args) == 0 {
// 		listValidTypes()
// 	} else {
// 		listType(args[0])
// 	}
// }

func listValidTypes() {
	fmt.Println("Available types:")
	for _, t := range validTypes {
		fmt.Printf(" - %s\n", t)
	}
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false // path doesn't exist or isn't accessible
	}
	return info.IsDir()
}

// func listType(t string) {
// 	folderInfoList, err := GetTypedFolders(t)
//
// 	if err != nil {
// 		fmt.Print(oopsie.CreateOopsie().Title("Encountered an issue").Error(err).Render())
// 	}
//
// 	if len(folderInfoList) == 0 {
//
// 		fmt.Print(oopsie.CreateOopsie().Title(fmt.Sprintf("No folders with type %s found", t)).IndicatorMessage("INFO").Error(fmt.Errorf("")).IndicatorColors(oopsie.GREEN, oopsie.BRIGHT_BLACK).Render())
// 	}
//
// 	for _, info := range folderInfoList {
// 		fmt.Printf("[%d] at %s", info.Inode, info.FullPath)
// 	}
// }
