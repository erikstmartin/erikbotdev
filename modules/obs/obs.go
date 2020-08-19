package obs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	obsws "github.com/christopher-dG/go-obs-websocket"
	"github.com/erikstmartin/erikbotdev/bot"
)

type Config struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

var config Config

func init() {
	bot.RegisterModule(bot.Module{
		Name: "obs",
		Actions: map[string]bot.ActionFunc{
			"SourceFilterEnabled": enableSourceFilterAction,
			"ChangeScene":         changeSceneAction,
			"StopStream":          stopStreamAction,
		},
		Init: func(c json.RawMessage) error {
			if err := json.Unmarshal(c, &config); err != nil {
				return err
			}
			// TODO: Get host:port from config and connect here
			port, err := strconv.ParseInt(config.Port, 10, 32)
			if err != nil {
				fmt.Printf("Failed to parse OBS_PORT: %s\n", err)
			}
			err = Connect(config.Host, int(port))
			if err != nil {
				return err
			}

			client.AddEventHandler("SwitchScenes", func(e obsws.Event) {
				// Make sure to assert the actual event type.
				bot.Status.Scene = e.(obsws.SwitchScenesEvent).SceneName
			})

			client.AddEventHandler("StreamStatus", func(e obsws.Event) {
				// Make sure to assert the actual event type.
				bot.Status.Streaming = e.(obsws.StreamStatusEvent).Streaming
			})

			// Ensure we set the current status on the bot
			statusReq := obsws.NewGetStreamingStatusRequest()
			status, err := statusReq.SendReceive(client)
			if err != nil {
				return err
			}
			bot.Status.Streaming = status.Streaming

			sceneReq := obsws.NewGetCurrentSceneRequest()
			scene, err := sceneReq.SendReceive(client)
			if err != nil {
				return err
			}
			bot.Status.Scene = scene.Name

			return nil
		},
	})
}

var client obsws.Client

func Connect(host string, port int) error {
	client = obsws.Client{Host: host, Port: port}
	if err := client.Connect(); err != nil {
		return err
	}
	obsws.SetReceiveTimeout(time.Second * 2)
	return nil
}

func Disconnect() error {
	return client.Disconnect()
}

func Streaming() (bool, error) {
	statusReq := obsws.NewGetStreamingStatusRequest()
	status, err := statusReq.SendReceive(client)
	if err != nil {
		return false, err
	}
	return status.Streaming, nil
}

func StopStream() error {
	statusReq := obsws.NewGetStreamingStatusRequest()
	status, err := statusReq.SendReceive(client)
	if err != nil {
		return err
	}

	if !status.Streaming {
		return nil
	}

	req := obsws.NewStartStopStreamingRequest()
	_, err = req.SendReceive(client)
	return err
}

func stopStreamAction(a bot.Action, cmd bot.Params) error {
	return StopStream()
}

func EnableSourceFilter(sourceName string, filterName string, enabled bool) error {
	req := obsws.NewSetSourceFilterVisibilityRequest(sourceName, filterName, enabled)
	if _, err := req.SendReceive(client); err != nil {
		return err
	}

	return nil
}

func enableSourceFilterAction(a bot.Action, cmd bot.Params) error {
	if _, ok := a.Args["source"]; !ok {
		return fmt.Errorf("Argument 'source' is required.")
	}

	if _, ok := a.Args["filterName"]; !ok {
		return fmt.Errorf("Argument 'filterName' is required.")
	}

	if _, ok := a.Args["enabled"]; !ok {
		return fmt.Errorf("Argument 'enabled' is required.")
	}

	enabled := true
	if strings.ToLower(a.Args["enabled"]) == "false" {
		enabled = false
	}
	return EnableSourceFilter(a.Args["source"], a.Args["filterName"], enabled)
}

func ChangeScene(scene string) error {
	req := obsws.NewSetCurrentSceneRequest(scene)
	if _, err := req.SendReceive(client); err != nil {
		return err
	}
	return nil
}

func changeSceneAction(a bot.Action, cmd bot.Params) error {
	if _, ok := a.Args["scene"]; !ok {
		return fmt.Errorf("Argument 'scene' is required.")
	}

	return ChangeScene(a.Args["scene"])
}
