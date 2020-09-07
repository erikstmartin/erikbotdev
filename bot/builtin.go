package bot

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

var builtinCommands map[string]CommandFunc = map[string]CommandFunc{
	"help":     helpCmd,
	"commands": helpCmd,
	"me":       userInfoCmd,
	"props":    givePointsCmd,
	"sounds":   soundListCmd,
	"so":       shoutoutCmd,
}

func TwitchSay(cmd Params, msg string) error {
	args := map[string]string{
		"channel": cmd.Channel,
		"message": msg,
	}
	return ExecuteAction("twitch", "Say", args, cmd)
}

func helpCmd(cmd Params) error {
	if len(cmd.CommandArgs) > 0 {
		cname := cmd.CommandArgs[0]
		if c, ok := config.Commands[cname]; ok {
			return TwitchSay(cmd, fmt.Sprintf("%s: %s", cname, c.Description))
		}

		return nil
	}

	cmds := make([]string, 0)
	for _, c := range config.Commands {
		if c.Enabled {
			cmds = append(cmds, c.Name)
		}
	}

	return TwitchSay(cmd, strings.Join(cmds, ", "))
}

func userInfoCmd(cmd Params) error {
	u, err := GetUser(cmd.UserID)
	if err != nil {
		return err
	}

	return TwitchSay(cmd, fmt.Sprintf("%s: %d points", u.DisplayName, u.Points))
}

func givePointsCmd(cmd Params) error {
	if len(cmd.CommandArgs) != 2 {
		return nil
	}

	user, err := GetUser(cmd.UserID)
	if err != nil {
		return err
	}

	points, err := strconv.ParseUint(cmd.CommandArgs[1], 10, 64)
	if err != nil {
		return err
	}

	recipient := strings.TrimPrefix(cmd.CommandArgs[0], "@")
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

func soundListCmd(cmd Params) error {
	files, err := ioutil.ReadDir(MediaPath())
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

// TODO; Hit Twitch API and ensure user exists
func shoutoutCmd(cmd Params) error {
	if len(cmd.CommandArgs) > 0 {
		user := cmd.CommandArgs[0]
		return TwitchSay(cmd, fmt.Sprintf("Shoutout %s! Check out their channel, shower them with follows and subs: https://twitch.tv/%s", user, user))
	}

	return fmt.Errorf("username is required")
}
