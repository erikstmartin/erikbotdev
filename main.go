package main

import (
	"fmt"
	"os"

	"github.com/erikstmartin/erikbotdev/bot"
	"github.com/erikstmartin/erikbotdev/cmd"
	_ "github.com/erikstmartin/erikbotdev/modules/bot"
)

func main() {
	// TODO: don't look in CWD. Look relative to home or bot executable, or /etc
	file, err := os.Open("./config.json")
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
