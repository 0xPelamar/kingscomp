/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve command",

	Run: serve,
}

func serve(cmd *cobra.Command, args []string) {
	fmt.Println("serve called")
}

func init() {
	rootCmd.AddCommand(serveCmd)

}
