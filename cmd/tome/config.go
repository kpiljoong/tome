package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kpiljoong/tome/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var setDefaultRemote string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or set config",
	Run: func(cmd *cobra.Command, args []string) {
		if setDefaultRemote != "" {
			write := config.Config{DefaultRemote: setDefaultRemote}
			data, _ := yaml.Marshal(&write)
			err := os.WriteFile(config.Path(), data, 0o644)
			if err != nil {
				log.Fatalf("Failed to write config: %v", err)
			}
			fmt.Printf("Default remote set: %s\n", setDefaultRemote)
			return
		}
	},
}

func init() {
	configCmd.Flags().StringVar(&setDefaultRemote, "default-remote", "", "Set default remote target (e.g., s3://my-bucket/path)")
}
