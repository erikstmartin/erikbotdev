package bot

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

var builtinCommands map[string]CommandFunc = map[string]CommandFunc{
	"help":   helpCmd,
	"me":     userInfoCmd,
	"props":  givePointsCmd,
	"sounds": soundListCmd,
}

func helpCmd(cmd UserCommand) error {
	// TODO: If any arguments are supplied, return description
	cmds := make([]string, 0)
	for _, c := range config.Commands {
		if c.Enabled {
			cmds = append(cmds, c.Name)
		}
	}
	args := map[string]string{
		"channel": cmd.Channel,
		"message": strings.Join(cmds, ", "),
	}
	return ExecuteAction("Twitch", "Say", args, cmd)
}

func userInfoCmd(cmd UserCommand) error {
	// TODO: Pull this out into bot actions that are automatically registered?
	u, err := GetUser(cmd.UserID)
	if err != nil {
		return err
	}

	args := map[string]string{
		"channel": cmd.Channel,
		"message": fmt.Sprintf("%s: %d points", u.DisplayName, u.Points),
	}
	return ExecuteAction("Twitch", "Say", args, cmd)
}

func givePointsCmd(cmd UserCommand) error {
	if len(cmd.Args) != 2 {
		return nil
	}

	user, err := GetUser(cmd.UserID)
	if err != nil {
		return err
	}

	points, err := strconv.ParseUint(cmd.Args[1], 10, 64)
	if err != nil {
		return err
	}

	twitchUser, err := GetTwitchUserByName(cmd.Args[0])
	if err != nil {
		return nil
	}

	// Allow owner to give unlimited points
	if strings.ToLower(user.DisplayName) == strings.ToLower(cmd.Channel) {
		destUser, err := GetUser(twitchUser.ID)
		if err != nil {
			return err
		}
		destUser.GivePoints(points)
	} else {
		user.TransferPoints(points, twitchUser.ID)
	}

	return nil
}

func soundListCmd(cmd UserCommand) error {
	// TODO: This should really be a configurable directory with a default location ($HOME??)
	files, err := ioutil.ReadDir("./media")
	if err != nil {
		return err
	}

	sounds := make([]string, 0, len(files))
	for _, f := range files {
		if !f.IsDir() {
			name := filepath.Base(f.Name())
			sounds = append(sounds, strings.TrimSuffix(name, filepath.Ext(name)))
		}
	}

	args := map[string]string{
		"channel": cmd.Channel,
		"message": strings.Join(sounds, ", "),
	}
	return ExecuteAction("Twitch", "Say", args, cmd)
}
