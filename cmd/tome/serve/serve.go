package serve

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/server"
)

var port int

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the Tome API",
	Long:  `Serve the Tome API on the given port.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.Start(port); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	},
}

func init() {
	ServeCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to serve the Tome API on")
}
