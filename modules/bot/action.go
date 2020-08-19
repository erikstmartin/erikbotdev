package bot

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/erikstmartin/erikbotdev/bot"
	"github.com/erikstmartin/erikbotdev/http"
)

func init() {
	bot.RegisterModule(bot.Module{
		Name: "bot",
		Actions: map[string]bot.ActionFunc{
			"Sleep":     sleepAction,
			"PlaySound": playSoundAction,
			"ShellExec": shellExecAction,
		},
	})
}

func sleepAction(a bot.Action, cmd bot.Params) error {
	var d string
	var ok bool

	if d, ok = a.Args["duration"]; !ok {
		return fmt.Errorf("Argument 'duration' is required.")
	}

	duration, err := time.ParseDuration(d)
	if err != nil {
		return err
	}

	time.Sleep(duration)
	return nil
}

type PlaySoundMessage struct {
	Sound string `json:"sound"`
}

func playSoundAction(a bot.Action, cmd bot.Params) error {
	var s string
	var ok bool
	if s, ok = a.Args["sound"]; !ok {
		return fmt.Errorf("Argument 'sound' is required.")
	}
	// TODO: Check media directory to ensure sound exists
	// Also ensure path traversal is accounted for
	http.BroadcastMessage(&PlaySoundMessage{
		Sound: s,
	})
	return nil
}

func shellExecAction(a bot.Action, cmd bot.Params) error {
	var s string
	var ok bool

	if s, ok = a.Args["command"]; !ok {
		return fmt.Errorf("Argument 'command' is required.")
	}

	var args = make([]string, 0)
	if passArgs, ok := a.Args["passArgs"]; ok && strings.ToLower(passArgs) == "true" {
		args = cmd.CommandArgs
	}

	shellCmd := exec.Command(s, args...)
	out, err := shellCmd.CombinedOutput()
	if err != nil {
		return err
	}

	if output, ok := a.Args["output"]; ok && strings.ToLower(output) == "true" {
		return bot.TwitchSay(cmd, string(out))
	}
	return nil
}
