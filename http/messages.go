package http

import "github.com/erikstmartin/erikbotdev/bot"

type Message interface {
}

type WebsocketMessage struct {
	Type    string  `json:"type"`
	Message Message `json:"message"`
}

type ChatMessage struct {
	User *bot.User `json:"user"`
	Text string    `json:"text"`
}

type RaidMessage struct {
	UserName     string `json:"userName"`
	PartySize    uint16 `json:"partySize"`
	ProfileImage string `json:"profileImage"`
	Message      string `json:"message"`
}
