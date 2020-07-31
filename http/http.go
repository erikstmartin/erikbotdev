package http

import (
	"net/http"

	"github.com/erikstmartin/erikbotdev/bot"
)

var hub *Hub

func Start(addr string, webPath string) error {
	hub = newHub()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// TODO: Fix path to web
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/public"))))
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Remove hardcoded path
		http.ServeFile(w, r, "./web/public/index.html")
	})

	return http.ListenAndServe(addr, nil)
}

func BroadcastMessage(msg Message) error {
	return hub.BroadcastMessage(msg)
}

func BroadcastChatMessage(user *bot.User, msg string) error {
	m := &ChatMessage{
		User: user,
		Text: msg,
	}

	return hub.BroadcastMessage(m)
}
