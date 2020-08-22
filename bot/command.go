package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicklaw5/helix"
)

type ActionFunc func(Action, Params) error
type CommandFunc func(Params) error

type ModuleInitFunc func(config json.RawMessage) error

var modules []Module
var registeredActions map[string]ActionFunc

var config Config
var Status status
var helixClient *helix.Client

type Config struct {
	Commands       map[string]*Command        `json:"commands"`
	Triggers       map[string]Trigger         `json:"triggers"`
	EnabledModules []string                   `json:"enabledModules"`
	DatabasePath   string                     `json:"databasePath"`
	WebPath        string                     `json:"webPath"`
	MediaPath      string                     `json:"mediaPath"`
	ModuleConfig   map[string]json.RawMessage `json:"moduleConfig"`
}

func WebPath() string {
	if config.WebPath == "" {
		path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return filepath.Join(path, "web")
	}

	return config.WebPath
}

func MediaPath() string {
	if config.MediaPath == "" {
		path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return filepath.Join(path, "media")
	}

	return config.MediaPath
}

func DatabasePath() string {
	if config.DatabasePath == "" {
		path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return filepath.Join(path, "bot.db")
	}

	return config.DatabasePath
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

type Params struct {
	Channel     string
	UserID      string
	UserName    string
	Command     string
	CommandArgs []string
	Payload     map[string]string
}

type Module struct {
	Name    string
	Actions map[string]ActionFunc
	Init    ModuleInitFunc
}

type Trigger struct {
	Actions []Action `json:"actions"`
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

func ExecuteAction(module string, name string, args map[string]string, cmd Params) error {
	action := fmt.Sprintf("%s::%s", module, name)
	if f, ok := registeredActions[action]; ok {
		return f(Action{Name: action, Args: args}, cmd)
	}
	return nil
}

func ExecuteCommand(cmd Params) error {
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
					if len(cmd.CommandArgs) >= i+1 {
						a.Args[argName] = cmd.CommandArgs[i]
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

func ExecuteTrigger(name string, cmd Params) error {
	if t, ok := config.Triggers[name]; ok {

		for _, a := range t.Actions {
			parts := strings.Split(a.Name, "::")
			if len(parts) >= 2 {
				ExecuteAction(parts[0], parts[1], a.Args, cmd)
			}
		}
	}

	return nil
}

func LoadConfig(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(&config); err != nil {
		return err
	}

	for key := range config.Commands {
		cmd := config.Commands[key]
		cmd.Name = key
	}

	return nil
}

func Init() error {
	for _, m := range modules {
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

	token, err := helixClient.GetAppAccessToken()
	if err != nil {
		return err
	}

	helixClient.SetUserAccessToken(token.Data.AccessToken)
	return nil
}
