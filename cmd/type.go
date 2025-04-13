package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Kaya-Sem/oopsie"
	"github.com/spf13/cobra"
)

const (
	RECEIVER = "receiver"
	VIRTUAL  = "virtual"
	SOURCE   = "source"
)

var validTypes = []string{RECEIVER, VIRTUAL, SOURCE}

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

	listCmd := &cobra.Command{
		Use:   "list type",
		Short: "List all directories with given type",
		Long:  `List the directories with given type in the semlink xattr data.`,
		Args:  cobra.MaximumNArgs(1),
		Run:   runTypeList,
	}

	typeCmd.AddCommand(listCmd)

	rootCmd.AddCommand(typeCmd)
}

func isValidType(typeArg string) bool {
	for _, t := range validTypes {
		if t == typeArg {
			return true
		}
	}
	return false
}

// TODO: ensure path is a folder
func setType(path, typeArg string) error {
	ensureIsPrivileged()

	if !isValidType(typeArg) {
		fmt.Print(oopsie.CreateOopsie().Title("Invalid type").IndicatorColors(oopsie.BLACK, oopsie.RED).Error(fmt.Errorf("Invalid type specified: %s. Must be 'source' or 'receiver'", typeArg)).Render())
		os.Exit(1)
	}

	setXattr(path, semlinkTypeXattrKey, typeArg)

	return nil
}

func runTypeSet(cmd *cobra.Command, args []string) {
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

func runTypeList(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		listValidTypes()
	} else {
		listType(args[0])
	}
}

func listValidTypes() {
	fmt.Println("Available types:")
	for _, t := range validTypes {
		fmt.Printf(" - %s\n", t)
	}
}

func listType(t string) {
	folderInfoList, err := GetTypedFolders(t)

	if err != nil {
		fmt.Print(oopsie.CreateOopsie().Title("Encountered an issue").Error(err).Render())
	}

	if len(folderInfoList) == 0 {

		fmt.Print(oopsie.CreateOopsie().Title(fmt.Sprintf("No folders with type %s found", t)).IndicatorMessage("INFO").Error(fmt.Errorf("")).IndicatorColors(oopsie.GREEN, oopsie.BRIGHT_BLACK).Render())
	}

	for _, info := range folderInfoList {
		fmt.Printf("[%d] at %s", info.Inode, info.FullPath)
	}
}
