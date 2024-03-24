/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configures the target server for our SOS app",
	Long:  `It sets up the server with various binaries and scripts for our OpenPubkey 
	authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("---Setup Initiated----")
		serverFlag, _ := cmd.Flags().GetString("server")
		keyFlag, _ := cmd.Flags().GetString("key")
		emailFlag, _ := cmd.Flags().GetString("email")
		emailArgs := emailFlag
		userArgs := serverFlag

		principal := strings.Split(userArgs, "@")[0]
		fmt.Println(principal)
		privateKeyAuth := false
		privateKeyPath := ""

		if keyFlag != "" {
			fmt.Println("Private Key Mode selected.")
			privateKeyAuth = true
			privateKeyPath = keyFlag
		}

		var scpCommandScript []string
		var scpCommandVerifier []string
		var sshCommandChmod []string
		var sshCommandConfigure []string

		if privateKeyAuth && privateKeyPath != "" {
			scpCommandScript = []string{
				"scp",
				"-i",
				privateKeyPath,
				"scripts/configure-sos-server.sh",
				userArgs + ":/root/configure-sos-server.sh",
			}
			scpCommandVerifier = []string{
				"scp",
				"-i",
				privateKeyPath,
				"verifier/verifier",
				userArgs + ":/root/verifier",
			}
			sshCommandChmod = []string{
				"ssh",
				"-i",
				privateKeyPath,
				userArgs,
				"chmod",
				"+x",
				"/root/configure-sos-server.sh",
			}
			sshCommandConfigure = []string{
				"ssh",
				"-i",
				privateKeyPath,
				userArgs,
				"/root/configure-sos-server.sh",
				emailArgs,
				principal,
			}
		} else {
			scpCommandScript = []string{
				"scp",
				"scripts/configure-sos-server.sh",
				userArgs + ":/root/configure-sos-server.sh",
			}
			scpCommandVerifier = []string{
				"scp",
				"verifier/verifier",
				userArgs + ":/root/verifier",
			}
			sshCommandChmod = []string{
				"ssh",
				userArgs,
				"chmod",
				"+x",
				"/root/configure-sos-server.sh",
			}
			sshCommandConfigure = []string{
				"ssh",
				userArgs,
				"/root/configure-sos-server.sh",
				emailArgs,
				principal,
			}
		}

		scpCmdScript := exec.Command(
			scpCommandScript[0], scpCommandScript[1:]...,
		)

		fmt.Println("Copying configuration script to server...")

		err := scpCmdScript.Run()
		if err != nil {
			log.Fatal(err)
		}

		scpCmdVerifier := exec.Command(
			scpCommandVerifier[0], scpCommandVerifier[1:]...,
		)

		fmt.Println("Copying verifier to server...")

		err = scpCmdVerifier.Run()
		if err != nil {
			log.Fatal(err)
		}

		sshCmdChmod := exec.Command(
			sshCommandChmod[0], sshCommandChmod[1:]...,
		)

		fmt.Println("Making configuration script executable...")

		err = sshCmdChmod.Run()
		if err != nil {
			log.Fatal(err)
		}

		sshCmdConfigure := exec.Command(
			sshCommandConfigure[0], sshCommandConfigure[1:]...,
		)

		fmt.Println("Configuring SOS server...")

		err = sshCmdConfigure.Run()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Configured SOS server for:", emailArgs)
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.Flags().String("server", "", "The server you want to configure in")
	configureCmd.Flags().String("email", "", "The email you want to configure with")
	configureCmd.Flags().StringP("key", "i", "", "To choose if you want to use private key for setup")
}
