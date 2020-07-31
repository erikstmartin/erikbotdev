package cmd

import (
	"fmt"

	"github.com/erikstmartin/erikbotdev/modules/hue"
	"github.com/spf13/cobra"
)

var hueCmd = &cobra.Command{
	Use:   "hue",
	Short: "commands for configuring hue lights",
	Long:  `TODO: fix me`,
}

var hueCreateUserCmd = &cobra.Command{
	Use:   "user-create",
	Short: "commands for configuring hue lights",
	Long:  `TODO: fix me`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: This is where our server code will go
		if len(args) == 0 {
			fmt.Println("You must supply a username")
			return
		}
		user, err := hue.CreateUser(args[0])
		if err != nil {
			fmt.Println("Error creating user", err)
			return
		}
		fmt.Printf("User created. Please save this to an environment variable HUE_USER='%s'\n", user)
	},
}

var hueBridgeListCmd = &cobra.Command{
	Use:   "bridge-list",
	Short: "List Hue bridges",
	Long:  `TODO: fix me`,
	Run: func(cmd *cobra.Command, args []string) {
		bridges, err := hue.ListBridges()
		if err != nil {
			fmt.Println("Error retrieving list of bridges", err)
		}

		for _, b := range bridges {
			fmt.Println(b)
		}
	},
}

var hueLightListCmd = &cobra.Command{
	Use:   "light-list",
	Short: "List Hue lights",
	Long:  `TODO: fix me`,
	Run: func(cmd *cobra.Command, args []string) {
		lights, err := hue.ListLights()
		if err != nil {
			fmt.Println("Error retrieving list of lights", err)
		}

		for _, l := range lights {
			fmt.Println(l)
		}
	},
}

var hueRoomListCmd = &cobra.Command{
	Use:   "room-list",
	Short: "List Hue rooms",
	Long:  `TODO: fix me`,
	Run: func(cmd *cobra.Command, args []string) {
		rooms, err := hue.ListRooms()
		if err != nil {
			fmt.Println("Error retrieving list of rooms", err)
		}

		for _, r := range rooms {
			fmt.Println(r)
		}
	},
}

var hueRoomHueCmd = &cobra.Command{
	Use:   "room-hue",
	Short: "Change hue of room",
	Long:  `TODO: fix me`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid arguments")
			return
		}

		color, err := hue.ParseColor(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		err = hue.RoomHue("Office", color)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var hueZoneListCmd = &cobra.Command{
	Use:   "zone-list",
	Short: "List Hue zones",
	Long:  `TODO: fix me`,
	Run: func(cmd *cobra.Command, args []string) {
		zones, err := hue.ListZones()
		if err != nil {
			fmt.Println("Error retrieving list of rooms", err)
		}

		for _, z := range zones {
			fmt.Println(z)
		}
	},
}

var hueRoomAlertCmd = &cobra.Command{
	Use:   "room-alert",
	Short: "Flash lights in room",
	Long:  `TODO: fix me`,
	Run: func(cmd *cobra.Command, args []string) {
		err := hue.RoomAlert("Office", "lselect")
		if err != nil {
			fmt.Println("Error alerting room", err)
		}
	},
}

func initHueCmd() {
	rootCmd.AddCommand(hueCmd)
	hueCmd.AddCommand(hueCreateUserCmd)
	hueCmd.AddCommand(hueBridgeListCmd)
	hueCmd.AddCommand(hueLightListCmd)
	hueCmd.AddCommand(hueRoomListCmd)
	hueCmd.AddCommand(hueZoneListCmd)
	hueCmd.AddCommand(hueRoomAlertCmd)
	hueCmd.AddCommand(hueRoomHueCmd)
}
