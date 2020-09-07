package twitch

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/erikstmartin/erikbotdev/bot"
	"github.com/erikstmartin/erikbotdev/http"
	"github.com/gempir/go-twitch-irc/v2"
	"github.com/nicklaw5/helix"
)

type Config struct {
	MainChannel  string   `json:"mainChannel"`
	ClientID     string   `json:"clientID"`
	ClientSecret string   `json:"clientSecret"`
	OauthToken   string   `json:"oauthToken"`
	Channels     []string `json:"channels"`
}

func (c *Config) GetClientID() string {
	if strings.HasPrefix(c.ClientID, "$") {
		return os.Getenv(strings.TrimPrefix(c.ClientID, "$"))
	}
	return c.ClientID
}

func (c *Config) GetClientSecret() string {
	if strings.HasPrefix(c.ClientSecret, "$") {
		return os.Getenv(strings.TrimPrefix(c.ClientSecret, "$"))
	}
	return c.ClientSecret
}

func (c *Config) GetOauthToken() string {
	if strings.HasPrefix(c.OauthToken, "$") {
		return os.Getenv(strings.TrimPrefix(c.OauthToken, "$"))
	}
	return c.OauthToken
}

var client *twitch.Client
var config Config

func init() {
	bot.RegisterModule(bot.Module{
		Name: "twitch",
		Actions: map[string]bot.ActionFunc{
			"Say":    sayAction,
			"Uptime": uptimeAction,
		},
		Init: func(c json.RawMessage) error {
			return json.Unmarshal(c, &config)
		},
	})
}

func uptimeAction(a bot.Action, cmd bot.Params) error {
	var channel = cmd.Channel

	if _, ok := a.Args["channel"]; ok {
		channel = a.Args["channel"]
	}

	streamResp, err := bot.GetHelixClient().GetStreams(&helix.StreamsParams{
		UserLogins: []string{config.MainChannel},
	})
	if err != nil {
		return err
	}
	streams := streamResp.Data.Streams
	if len(streams) != 1 {
		return fmt.Errorf(
			"Expected 1 active stream for %s, got %d",
			config.MainChannel,
			len(streams),
		)
	}

	startedAt := streams[0].StartedAt.Truncate(time.Minute)
	uptime := time.Now().Truncate(time.Minute).Sub(startedAt)
	client.Say(
		channel,
		fmt.Sprintf(
			"I've been streaming for %d hours, %d minutes",
			int(uptime.Hours()),
			int(uptime.Minutes()),
		),
	)
	return nil
}

func sayAction(a bot.Action, cmd bot.Params) error {
	var channel = cmd.Channel

	if _, ok := a.Args["channel"]; ok {
		channel = a.Args["channel"]
	}

	if _, ok := a.Args["message"]; !ok {
		return fmt.Errorf("Argument 'message' is required.")
	}
	client.Say(channel, a.Args["message"])
	return nil
}

func Run() error {
	client = twitch.NewClient(config.MainChannel, config.GetOauthToken())

	client.OnConnect(func() {
		fmt.Println("Connected!")
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		var u *bot.User
		var err error

		u, err = bot.GetUser(message.User.ID)
		if err != nil {
			fmt.Println("Error retrieving user: ", err)
			return
		}

		u.DisplayName = message.User.DisplayName
		u.Color = message.User.Color
		u.Badges = message.User.Badges

		if u.New {
			u.ID = message.User.ID
			u.Points = 25000
		}

		err = u.Save()
		if err != nil {
			fmt.Println("Error saving user: ", err)
			return
		}

		if !strings.HasPrefix(message.Message, "!") || len(message.Message) <= 1 {
			u.GivePoints(1000)

			if message.Channel == config.MainChannel {
				bot.ExecuteTrigger("twitch::Chat", bot.Params{
					UserID:   u.ID,
					UserName: u.DisplayName,
					Channel:  message.Channel,
				})
				http.BroadcastChatMessage(u, message.Message)
			}
			return
		}

		if message.Channel == config.MainChannel {
			parts := strings.Fields(message.Message[1:])
			cmdName := strings.ToLower(parts[0])
			cmd := bot.Params{
				Channel:     message.Channel,
				UserID:      message.User.ID,
				UserName:    message.User.DisplayName,
				Command:     cmdName,
				CommandArgs: parts[1:],
			}
			err = bot.ExecuteCommand(cmd)
			if err != nil {
				fmt.Println("Error executing command: ", err)
			}
		}
	})

	//TODO: Respond to Twitch events
	//https://dev.twitch.tv/docs/irc/tags#usernotice-twitch-tags

	client.OnUserNoticeMessage(func(message twitch.UserNoticeMessage) {
		// TODO: Leave this here, till we've implement all notice messages
		b, _ := json.Marshal(message)
		fmt.Println("UserNoticeMessage", string(b))

		// TODO: Document all possible triggers
		bot.ExecuteTrigger(fmt.Sprintf("twitch::%s", message.MsgID), bot.Params{
			UserID:   message.User.ID,
			UserName: message.User.DisplayName,
			Channel:  message.Channel,
			Payload:  message.Tags,
		})

		switch message.MsgID {
		case "sub":
		case "resub":
		case "subgift":
		case "anonsubgift":
		case "submysterygift":
		case "giftpaidupgrade":
		case "rewardgift":
		case "anongiftpaidupgrade":
		case "raid":
			/* TODO: Features
			- Notify websocket to to display message
			- Give raider points?
			- Give raiding party members points?
			- Say something in chat?
			*/
		case "unraid":
			// TODO: You're dead to me!
		case "ritual":
		case "bitsbadgetier":
			// TODO: Where are the bits?
			// TODO: Where are the follows?
			// TODO: Can we get insight into channel point redemptions?
		}
	})

	client.Join(config.Channels...)

	return client.Connect()
}
