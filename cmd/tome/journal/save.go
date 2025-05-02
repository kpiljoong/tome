package journal

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/core"
)

var SaveCmd = &cobra.Command{
	Use:   "save [namespace] [file]",
	Short: "Save a file into the given namespace",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		file := args[1]

		entry, err := core.Save(namespace, file)
		if err != nil {
			fmt.Printf("Error saving file: %v\n", err)
			return
		}
		fmt.Printf("Saved file %s with hash %v\n", file, entry)

		fmt.Printf(" Saved %s into namesapace %s\n", file, namespace)
	},
}
