package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/erikstmartin/erikbotdev/bot"
	"github.com/erikstmartin/erikbotdev/cmd"
	_ "github.com/erikstmartin/erikbotdev/modules/bot"
)

var configFileName string

func init() {
	configFileName = os.Getenv("ERIKBOTDEV_CONFIG_FILE_NAME")
	if configFileName == "" {
		configFileName = "erikbotdev.json"
	}
}

func main() {
	file, err := os.Open(findConfigFile())
	if err != nil {
		fmt.Println(err)
		return
	}
	err = bot.LoadConfig(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = bot.Init(); err != nil {
		fmt.Println(err)
		return
	}

	cmd.Execute()
}

func findConfigFile() string {
	home, err := os.UserHomeDir()
	if err == nil {
		path := filepath.Join(home, configFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check relative to binary
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if _, err := os.Stat(filepath.Join(path, configFileName)); err == nil {
		return filepath.Join(path, configFileName)
	}

	return filepath.Join(".", configFileName)
}
