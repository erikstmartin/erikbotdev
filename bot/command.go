package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type ActionFunc func(Action) error
type ModuleConfig map[string]string
type ModuleInitFunc func(config ModuleConfig) error

var modules []Module
var registeredActions map[string]ActionFunc
var config Config
var Status status

type Config struct {
	TwitchUser string              `json:"twitchUser"`
	Commands   map[string]*Command `json:"commands"`
}

type status struct {
	Streaming bool
	Scene     string
}

type Action struct {
	Name       string            `json:"name"`
	Args       map[string]string `json:"args"`
	UserArgMap []string          `json:"userArgMap"`
}

type Command struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	Offline     bool     `json:"offline"`
	Actions     []Action `json:"actions"`
}

type UserCommand struct {
	Channel  string
	UserID   string
	UserName string
	Command  string
	Args     []string
}

type Module struct {
	Name    string
	Actions map[string]ActionFunc
	Init    ModuleInitFunc
}

func RegisterModule(m Module) error {
	if modules == nil {
		modules = make([]Module, 0)
	}
	modules = append(modules, m)

	for name, f := range m.Actions {
		if err := registerAction(m.Name, name, f); err != nil {
			return err
		}
	}
	return nil
}

func registerAction(module string, name string, f func(Action) error) error {
	n := fmt.Sprintf("%s::%s", module, name)

	if registeredActions == nil {
		registeredActions = make(map[string]ActionFunc)
	}

	if _, ok := registeredActions[n]; ok {
		return fmt.Errorf("Action %s exists already", n)
	}

	registeredActions[n] = f

	return nil
}

func ExecuteAction(module string, name string, args map[string]string) error {
	action := fmt.Sprintf("%s::%s", module, name)
	if f, ok := registeredActions[action]; ok {
		return f(Action{Name: action, Args: args})
	}
	return nil
}

func ExecuteCommand(cmd UserCommand) error {
	// Is it the help command?
	if strings.ToLower(cmd.Command) == "help" {
		// TODO: If any arguments are supplied, return description
		cmds := make([]string, 0)
		for _, c := range config.Commands {
			if c.Enabled {
				cmds = append(cmds, c.Name)
			}
		}
		args := map[string]string{
			"channel": "erikdotdev",
			"message": strings.Join(cmds, ","),
		}
		return ExecuteAction("Twitch", "Say", args)
	}

	if c, ok := config.Commands[cmd.Command]; ok && c.Enabled {
		if !Status.Streaming && !c.Offline {
			return nil
		}

		fmt.Println("Command executed", cmd.UserName, cmd.Command)
		for _, a := range c.Actions {
			if f, ok := registeredActions[a.Name]; ok {
				for i, argName := range a.UserArgMap {
					if len(cmd.Args) >= i+1 {
						a.Args[argName] = cmd.Args[i]
					}
				}

				if err := f(a); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return fmt.Errorf("Command not found %s", cmd.Command)
}

func LoadConfig(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(&config); err != nil {
		return err
	}

	for key, _ := range config.Commands {
		cmd := config.Commands[key]
		cmd.Name = key
	}

	return nil
}

func Init() error {
	for _, m := range modules {
		// TODO: Pass in any config info from config.json
		if m.Init != nil {
			if err := m.Init(nil); err != nil {
				return err
			}
		}
	}
	return nil
}
