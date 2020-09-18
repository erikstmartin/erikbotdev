package hue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amimof/huego"
	"github.com/erikstmartin/erikbotdev/bot"
)

var randColor *rand.Rand
var sleepDuration = 200 * time.Millisecond

var colorMap = map[string]uint16{
	"orange": 3000,
	"yellow": 7000,
	"red":    65000,
	"blue":   45000,
	"green":  30000,
	"purple": 53000,
	"pink":   60000,
}

type Config struct {
	User   string `json:"user"`
	Bridge string `json:"bridge"`
}

func (c *Config) GetUser() string {
	if strings.HasPrefix(c.User, "$") {
		return os.Getenv(strings.TrimPrefix(c.User, "$"))
	}
	return c.User
}

var config Config
var bridge *huego.Bridge

func init() {
	bot.RegisterModule(bot.Module{
		Name: "hue",
		Actions: map[string]bot.ActionFunc{
			"RoomHue":        roomHueAction,
			"RoomAlert":      roomAlertAction,
			"ZoneHue":        zoneHueAction,
			"ZoneAlert":      zoneAlertAction,
			"ZoneBrightness": zoneBrightnessAction,
			"RoomBrightness": roomBrightnessAction,
		},
		Init: func(c json.RawMessage) error {
			s := rand.NewSource(time.Now().UnixNano())
			randColor = rand.New(s)

			err := json.Unmarshal(c, &config)
			if err != nil {
				return err
			}

			bridge, err = getBridge()
			return err
		},
	})
}

func getBridge() (*huego.Bridge, error) {
	if bridge != nil {
		return bridge, nil
	}

	if config.Bridge != "" {
		return huego.New(config.Bridge, config.GetUser()), nil
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

	bridge = bridges[0].Login(config.GetUser())
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
	var groupRegex = regexp.MustCompile("^[a-zA-Z0-9]+$")
	if !groupRegex.Match([]byte(groupName)) {
		return fmt.Errorf("Invalid Group Name")
	}

	g, err := getGroup(groupName, groupType)
	if err != nil {
		return err
	}

	err = g.Hue(hue)
	time.Sleep(sleepDuration)
	return err
}

func roomHueAction(a bot.Action, cmd bot.Params) error {
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

func zoneHueAction(a bot.Action, cmd bot.Params) error {
	if _, ok := a.Args["zone"]; !ok {
		return fmt.Errorf("Argument 'zone' is required.")
	}

	if _, ok := a.Args["hue"]; !ok {
		return fmt.Errorf("Argument 'hue' is required.")
	}

	color, err := ParseColor(a.Args["hue"])
	if err != nil {
		return err
	}

	return ZoneHue(a.Args["zone"], color)
}

func roomAlertAction(a bot.Action, cmd bot.Params) error {
	if _, ok := a.Args["room"]; !ok {
		return fmt.Errorf("Argument 'room' is required.")
	}
	if _, ok := a.Args["type"]; !ok {
		return fmt.Errorf("Argument 'type' is required.")
	}

	return RoomAlert(a.Args["room"], a.Args["type"])
}

func zoneAlertAction(a bot.Action, cmd bot.Params) error {
	if _, ok := a.Args["zone"]; !ok {
		return fmt.Errorf("Argument 'zone' is required.")
	}
	if _, ok := a.Args["type"]; !ok {
		return fmt.Errorf("Argument 'type' is required.")
	}
	if _, ok := a.Args["hue"]; ok {
		color, err := ParseColor(a.Args["hue"])
		if err != nil {
			return err
		}
		if err := ZoneHue(a.Args["zone"], color); err != nil {
			return err
		}
	}

	return ZoneAlert(a.Args["zone"], a.Args["type"])
}

func RoomHue(roomName string, hue uint16) error {
	return groupHue(roomName, "Room", hue)
}

func ZoneHue(zoneName string, hue uint16) error {
	return groupHue(zoneName, "Zone", hue)
}

func groupAlert(groupName string, groupType string, alertType string) error {
	if alertType != "none" && alertType != "select" && alertType != "lselect" {
		return fmt.Errorf("alert type must be one of 'none', 'select', 'lselect'")
	}

	resp, err := bridge.GetGroupsContext(context.Background())
	if err != nil {
		return err
	}

	for _, g := range resp {
		if g.Type == groupType && g.Name == groupName {
			err := g.Alert(alertType)
			time.Sleep(sleepDuration)
			return err
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

func roomBrightnessAction(a bot.Action, cmd bot.Params) error {
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

func zoneBrightnessAction(a bot.Action, cmd bot.Params) error {
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
	if color == "rand" || color == "random" {
		return uint16(randColor.Intn(65000)), nil
	}

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
