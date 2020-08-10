package hue

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/amimof/huego"
	"github.com/erikstmartin/erikbotdev/bot"
)

var colorMap = map[string]uint16{
	"orange": 3000,
	"red":    65000,
	"blue":   45000,
	"green":  30000,
	"purple": 53000,
	"pink":   60000,
}

var bridge *huego.Bridge

func init() {
	bot.RegisterModule(bot.Module{
		Name: "Hue",
		Actions: map[string]bot.ActionFunc{
			"RoomHue":        roomHueAction,
			"RoomAlert":      roomAlertAction,
			"ZoneBrightness": zoneBrightnessAction,
			"RoomBrightness": roomBrightnessAction,
		},
		Init: func(c bot.ModuleConfig) error {
			var err error
			bridge, err = getBridge()
			return err
		},
	})
}

// TODO: Use WithTimeout on contexts
// TODO: Get Hue bridge host and user from config, not environment variables
func getBridge() (*huego.Bridge, error) {
	if bridge != nil {
		return bridge, nil
	}

	if h := os.Getenv("HUE_BRIDGE"); h != "" {
		return huego.New(h, os.Getenv("HUE_USER")), nil
	}

	bridges, err := huego.DiscoverAll()
	if err != nil {
		return nil, err
	}

	if len(bridges) == 0 {
		return nil, fmt.Errorf("More than one Hue bridge %v discovered. Please specify the bridge using the HUE_BRIDGE environment variable.", bridges)
	}

	if len(bridges) > 1 {
		return nil, fmt.Errorf("No Hue bridges found")
	}

	bridge = bridges[0].Login(os.Getenv("HUE_USER"))
	return bridge, nil
}

func CreateUser(user string) (string, error) {
	return bridge.CreateUser(user) // Link button needs to be pressed
}

func ListBridges() ([]string, error) {
	bridges := make([]string, 0)

	resp, err := huego.DiscoverAll()
	if err != nil {
		return bridges, err
	}

	for _, b := range resp {
		bridges = append(bridges, b.Host)
	}
	return bridges, nil
}

func listGroups(groupType string) ([]string, error) {
	groups := make([]string, 0)

	resp, err := bridge.GetGroupsContext(context.Background())
	if err != nil {
		return groups, err
	}

	for _, g := range resp {
		if g.Type == groupType {
			groups = append(groups, g.Name)
		}
	}

	return groups, err
}

func ListRooms() ([]string, error) {
	return listGroups("Room")
}

func ListZones() ([]string, error) {
	return listGroups("Zone")
}

func getGroup(groupName string, groupType string) (huego.Group, error) {
	var g huego.Group
	resp, err := bridge.GetGroupsContext(context.Background())
	if err != nil {
		return g, err
	}

	for _, g := range resp {
		if g.Type == groupType && g.Name == groupName {
			return g, nil
		}
	}
	return g, fmt.Errorf("Group not found: %s", groupName)
}

func groupHue(groupName string, groupType string, hue uint16) error {
	g, err := getGroup(groupName, groupType)
	if err != nil {
		return err
	}
	return g.Hue(hue)
}

func roomHueAction(a bot.Action, cmd bot.UserCommand) error {
	if _, ok := a.Args["room"]; !ok {
		return fmt.Errorf("Argument 'room' is required.")
	}

	if _, ok := a.Args["hue"]; !ok {
		return fmt.Errorf("Argument 'hue' is required.")
	}

	color, err := ParseColor(a.Args["hue"])
	if err != nil {
		return err
	}

	return RoomHue(a.Args["room"], color)
}

func roomAlertAction(a bot.Action, cmd bot.UserCommand) error {
	if _, ok := a.Args["room"]; !ok {
		return fmt.Errorf("Argument 'room' is required.")
	}
	// TODO: ensure it's one of 'none', 'select', 'lselect'
	if _, ok := a.Args["type"]; !ok {
		return fmt.Errorf("Argument 'type' is required.")
	}

	return RoomAlert(a.Args["room"], a.Args["type"])
}

func RoomHue(roomName string, hue uint16) error {
	return groupHue(roomName, "Room", hue)
}

func ZoneHue(zoneName string, hue uint16) error {
	return groupHue(zoneName, "Zone", hue)
}

func groupAlert(groupName string, groupType string, alertType string) error {
	resp, err := bridge.GetGroupsContext(context.Background())
	if err != nil {
		return err
	}

	for _, g := range resp {
		if g.Type == groupType && g.Name == groupName {
			return g.Alert(alertType)
		}
	}
	return fmt.Errorf("Group not found: %s", groupName)
}

func RoomAlert(roomName string, alertType string) error {
	return groupAlert(roomName, "Room", alertType)
}

func ZoneAlert(zoneName string, alertType string) error {
	return groupAlert(zoneName, "Zone", alertType)
}

func ListLights() ([]string, error) {
	lights := make([]string, 0)

	resp, err := bridge.GetLightsContext(context.Background())
	if err != nil {
		return lights, err
	}

	for _, l := range resp {
		lights = append(lights, l.Name)
	}

	return lights, err
}

func roomBrightnessAction(a bot.Action, cmd bot.UserCommand) error {
	if _, ok := a.Args["brightness"]; !ok {
		return fmt.Errorf("Argument 'brightness' is required.")
	}

	if _, ok := a.Args["room"]; !ok {
		return fmt.Errorf("Argument 'room' is required.")
	}

	b, err := strconv.ParseUint(a.Args["brightness"], 10, 8)
	if err != nil {
		return err
	}

	brightness := uint8(b)
	return GroupBrightness(a.Args["room"], "Room", brightness)
}

func zoneBrightnessAction(a bot.Action, cmd bot.UserCommand) error {
	if _, ok := a.Args["brightness"]; !ok {
		return fmt.Errorf("Argument 'brightness' is required.")
	}

	if _, ok := a.Args["zone"]; !ok {
		return fmt.Errorf("Argument 'zone' is required.")
	}

	b, err := strconv.ParseUint(a.Args["brightness"], 10, 8)
	if err != nil {
		return err
	}

	brightness := uint8(b)
	return GroupBrightness(a.Args["room"], "Zone", brightness)
}

func GroupBrightness(groupName string, groupType string, b uint8) error {
	g, err := getGroup(groupName, "Room")
	if err != nil {
		return err
	}

	return g.BriContext(context.Background(), b)
}

func ZoneBrightness(groupName string, b uint8) error {
	g, err := getGroup(groupName, "Room")
	if err != nil {
		return err
	}

	return g.BriContext(context.Background(), b)
}

func ParseColor(color string) (uint16, error) {
	var (
		colorCode uint16
		ok        bool
	)

	// First map string 'red' to uint16 color, if it doesn't exist. Try to parse uint16
	if colorCode, ok = colorMap[color]; !ok {
		c, err := strconv.ParseUint(color, 10, 16)
		if err != nil {
			return 0, err
		}
		colorCode = uint16(c)
	}

	return colorCode, nil
}
