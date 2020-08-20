package keylight

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/erikstmartin/erikbotdev/bot"
)

type Config struct {
	Lights []string `json:"lights"`
}

type Light struct {
	On          int `json:"on"`
	Brightness  int `json:"brightness"`
	Temperature int `json:"temperature"`
}

type LightOptions struct {
	Count  int     `json:"numberOfLights"`
	Lights []Light `json:"lights"`
}

var config Config

func init() {
	bot.RegisterModule(bot.Module{
		Name: "keylight",
		Actions: map[string]bot.ActionFunc{
			"Blink": blinkAction,
			"Power": powerAction,
		},
		Init: func(c json.RawMessage) error {
			return json.Unmarshal(c, &config)
		},
	})
}

func blinkAction(a bot.Action, cmd bot.Params) error {
	var count int64 = 1
	var duration = 250 * time.Millisecond
	var err error

	if _, ok := a.Args["count"]; ok {
		// TODO: parse count
		count, err = strconv.ParseInt(a.Args["count"], 10, 64)
	}

	if _, ok := a.Args["duration"]; ok {
		duration, err = time.ParseDuration(a.Args["duration"])
		if err != nil {
			return err
		}
	}

	for i := 0; int64(i) < count; i++ {
		a.Args["on"] = "false"
		powerAction(a, cmd)

		time.Sleep(duration)

		a.Args["on"] = "true"
		powerAction(a, cmd)

		time.Sleep(duration)
	}

	return nil
}

func powerAction(a bot.Action, cmd bot.Params) error {
	if _, ok := a.Args["on"]; !ok {
		return fmt.Errorf("Argument 'on' is required.")
	}

	for _, addr := range config.Lights {
		resp, err := http.Get(fmt.Sprintf("http://%s/elgato/lights", addr))
		if err != nil {
			return fmt.Errorf("Error fetching light info '%s': %s", addr, err)
		}

		var opt LightOptions
		if err = json.NewDecoder(resp.Body).Decode(&opt); err != nil {
			return fmt.Errorf("Error unmarshalling light info '%s': %s", addr, err)
		}

		for i, _ := range opt.Lights {
			if a.Args["on"] == "true" {
				opt.Lights[i].On = 1
			}

			if a.Args["on"] == "false" {
				opt.Lights[i].On = 0
			}
		}

		buf := bytes.NewBuffer([]byte{})
		err = json.NewEncoder(buf).Encode(opt)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("http://%s/elgato/lights", addr)
		req, err := http.NewRequestWithContext(context.Background(), "PUT", url, buf)
		client := &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// TODO: Read response body and check HTTP status code
	}
	return nil
}
