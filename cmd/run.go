package cmd

import (
	"log"
	"os"
	"os/signal"

	"github.com/erikstmartin/erikbotdev/bot"
	"github.com/erikstmartin/erikbotdev/http"
	"github.com/erikstmartin/erikbotdev/modules/obs"
	"github.com/erikstmartin/erikbotdev/modules/twitch"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run chatbot server",
	Long:  `TODO: fix me`,
	Run: func(cmd *cobra.Command, args []string) {
		go http.Start(":8080", "./web")

		// TODO: Start working on database
		err := bot.InitDatabase("bot.db", 0600)
		if err != nil {
			log.Fatal("Failed to initialize database: ", err)
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			<-sig

			obs.EnableSourceFilter("Heil PR40", "Deep", false)
			obs.EnableSourceFilter("Heil PR40", "HighPitch", false)
			obs.EnableSourceFilter("Heil PR40", "False", false)
			obs.Disconnect()
			os.Exit(0)
		}()
		if err := twitch.Run(); err != nil {
			panic(err)
		}
	},
}
