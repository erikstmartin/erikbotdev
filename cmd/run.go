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
		err := bot.InitDatabase(bot.DatabasePath(), 0600)
		if err != nil {
			if err.Error() == "timeout" {
				log.Fatal("Timeout opening database. Check to ensure another process does not have the database file open")
			}
			log.Fatal("Failed to initialize database: ", err)
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		go func() {
			<-sig

			bot.ExecuteTrigger("bot::Shutdown", bot.Params{
				Command: "shutdown",
			})

			if bot.IsModuleEnabled("OBS") {
				obs.Disconnect()
			}
			os.Exit(0)
		}()

		// TODO: Handle scenario where startup trigger contains a twitch action
		bot.ExecuteTrigger("bot::Startup", bot.Params{
			Command: "startup",
		})

		if err := twitch.Run(); err != nil {
			panic(err)
		}
	},
}
