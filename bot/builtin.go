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

func TwitchSay(cmd UserCommand, msg string) error {
	args := map[string]string{
		"channel": cmd.Channel,
		"message": msg,
	}
	return ExecuteAction("twitch", "Say", args, cmd)
}

func helpCmd(cmd UserCommand) error {
	// TODO: If any arguments are supplied, return description
	if len(cmd.Args) > 0 {
		cname := cmd.Args[0]
		return TwitchSay(cmd, fmt.Sprintf("%s: %s", cname, config.Commands[cname].Description))
	}

	cmds := make([]string, 0)
	for _, c := range config.Commands {
		if c.Enabled {
			cmds = append(cmds, c.Name)
		}
	}

	return TwitchSay(cmd, strings.Join(cmds, ", "))
}

func userInfoCmd(cmd UserCommand) error {
	u, err := GetUser(cmd.UserID)
	if err != nil {
		return err
	}

	return TwitchSay(cmd, fmt.Sprintf("%s: %d points", u.DisplayName, u.Points))
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

	recipient := strings.TrimPrefix(cmd.Args[0], "@")
	twitchUser, err := GetTwitchUserByName(recipient)
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

	return TwitchSay(cmd, strings.Join(sounds, ", "))
}
