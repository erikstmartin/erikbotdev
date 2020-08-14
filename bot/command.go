package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/nicklaw5/helix"
)

type ActionFunc func(Action, UserCommand) error
type CommandFunc func(UserCommand) error

type ModuleInitFunc func(config json.RawMessage) error

var modules []Module
var registeredActions map[string]ActionFunc

var config Config
var Status status
var helixClient *helix.Client

type Config struct {
	Commands       map[string]*Command        `json:"commands"`
	EnabledModules []string                   `json:"enabledModules`
	DatabasePath   string                     `json:"databasePath"`
	ModuleConfig   map[string]json.RawMessage `json:"moduleConfig"`
}

func DatabasePath() string {
	if config.DatabasePath != "" {
		return config.DatabasePath
	}

	return "bot.db"
}
func IsModuleEnabled(m string) bool {
	for _, mod := range config.EnabledModules {
		if mod == m {
			return true
		}
	}
	return false
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
	Points      uint64   `json:"points"`
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

func registerAction(module string, name string, f ActionFunc) error {
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

func ExecuteAction(module string, name string, args map[string]string, cmd UserCommand) error {
	action := fmt.Sprintf("%s::%s", module, name)
	if f, ok := registeredActions[action]; ok {
		return f(Action{Name: action, Args: args}, cmd)
	}
	return nil
}

func ExecuteCommand(cmd UserCommand) error {
	// First look in builtin commands
	if c, ok := builtinCommands[cmd.Command]; ok {
		return c(cmd)
	}

	// Next check user created commands
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

				if err := f(a, cmd); err != nil {
					return err
				}
			}
		}

		u, err := GetUser(cmd.UserID)
		if err == nil && !u.New {
			u.TakePoints(c.Points)
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
		if IsModuleEnabled(m.Name) && m.Init != nil {
			if err := m.Init(config.ModuleConfig[m.Name]); err != nil {
				return err
			}
		}
	}
	var err error
	helixClient, err = helix.NewClient(&helix.Options{
		ClientID:     os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
	})
	if err != nil {
		return err
	}

	// TODO: Better error handling to ensure valid token
	token, err := helixClient.GetAppAccessToken()
	if err != nil {
		return err
	}

	helixClient.SetUserAccessToken(token.Data.AccessToken)
	return nil
}
