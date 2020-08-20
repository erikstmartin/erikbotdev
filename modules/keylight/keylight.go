package keylight

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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
			"Blink":    blinkAction,
			"Settings": settingsAction,
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
		settingsAction(a, cmd)

		time.Sleep(duration)

		a.Args["on"] = "true"
		settingsAction(a, cmd)

		time.Sleep(duration)
	}

	return nil
}

func settingsAction(a bot.Action, cmd bot.Params) error {
	var brightness int
	var temperature int

	if b, ok := a.Args["brightness"]; ok {
		bright, err := strconv.ParseInt(b, 10, 64)
		if err != nil {
			return fmt.Errorf("Error parsing brightness: %s", err)
		}
		brightness = int(bright)
	}

	if t, ok := a.Args["temperature"]; ok {
		temp, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return fmt.Errorf("Error parsing temp: %s", err)
		}
		temperature = convertToElgato(int(temp))
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

			if _, ok := a.Args["brightness"]; ok {
				opt.Lights[i].Brightness = brightness
			}

			if _, ok := a.Args["temperature"]; ok {
				opt.Lights[i].Temperature = temperature
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

		if resp.StatusCode != 200 {
			json, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("Error setting light settings: %s", json)
		}

	}
	return nil
}

func convertToElgato(kelvin int) int {
	elgato := float64(kelvin-9900) / 20.35
	return int(math.Abs(math.Trunc(elgato)))
}
