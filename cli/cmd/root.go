/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sos",
	Short: "Secure OpenPubkey Shell(SOS)",
	Long:  `This project uses the OpenPubKey project for passwordless SSH using email authentication`,
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
