package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Zhima-Mochi/autovpn/pritunl"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "autovpn",
	Short: "Automatically connect to a Pritunl VPN server",
	Run: func(cmd *cobra.Command, args []string) {
		// Create a new Pritunl instance
		pm := pritunl.GetPritunlManager()

		// Fetch profiles and connections
		profiles, err := pm.Profiles()
		if err != nil {
			color.Red("Failed to fetch profiles: %v", err)
			return
		}
		conns, err := pm.Connections()
		if err != nil {
			color.Red("Failed to fetch connections: %v", err)
			return
		}

		// List profiles and connections
		if err := pm.List(profiles, conns); err != nil {
			color.Yellow("Failed to list profiles: %v", err)
			return
		}

		var id string
		prompt := &survey.Input{
			Message: "Enter ID or Server:",
			Default: "",
		}
		if err := survey.AskOne(prompt, &id); err != nil {
			color.Red("Error reading input: %v", err)
			return
		}

		if id == "" {
			return
		}

		// Check if profile exists
		var targetProfile pritunl.Profile
		isActionDisconnect := false
		for i, profile := range profiles {
			if strconv.Itoa(i+1) == id || strings.ToUpper(id) == profile.Server {
				targetProfile = profile
				isActionDisconnect = conns[profile.ID].Status == "connected"
				break
			}
		}

		if targetProfile == (pritunl.Profile{}) {
			color.Red("Profile does not exist!")
			return
		}

		// Perform disconnect action if profile is connected
		if isActionDisconnect {
			color.White("Disconnecting from %s...", targetProfile.Server)
			if err := pm.Disconnect(targetProfile.ID); err != nil {
				color.Red("Failed to disconnect: %v", err)
			}
			return
		}

		// Disconnect all active connections before connecting to the target profile
		for _, profile := range profiles {
			if _, ok := conns[profile.ID]; ok {
				color.White("Disconnecting from %s...", profile.Server)
				if err := pm.Disconnect(profile.ID); err != nil {
					color.Red("Failed to disconnect: %v", err)
				}
				time.Sleep(time.Second)
			}
		}

		// Connect to the target profile
		color.Yellow("Connecting to %s...", targetProfile.Server)
		if err := pm.Connect(targetProfile.ID); err != nil {
			color.Red("Failed to connect: %v", err)
			return
		}

		// Check the connection status
		timeout := time.NewTimer(30 * time.Second)

	Loop:
		for {
			select {
			case <-timeout.C:
				color.Red("Connection to %s timed out!", targetProfile.Server)
				break Loop
			default:
				connections, err := pm.Connections()
				status := connections[targetProfile.ID].Status
				if err != nil {
					color.Red("Failed to fetch connection status: %v", err)
					break Loop
				}
				switch status {
				case "connected":
					color.Green("Successfully connected to %s!", targetProfile.Server)
					break Loop
				case "connecting", "":
					// Do nothing
				default:
					color.Red("Failed to connect to %s: %s", targetProfile.Server, status)
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	},
}

func init() {
	// Disable default help and completion commands
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Error executing command: %v", err)
	}
}
