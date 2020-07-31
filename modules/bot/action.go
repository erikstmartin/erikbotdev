package bot

import (
	"fmt"
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
		},
	})
}

// TODO: Add more actions
// Execute cmd
// Say (respond back to chat)
// Play audio

func sleepAction(a bot.Action) error {
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

func playSoundAction(a bot.Action) error {
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
