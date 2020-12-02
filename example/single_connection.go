package main

import (
	"context"
	"fmt"

	"github.com/julez-dev/twitchirc"
)

func main() {
	conf := &twitchirc.Config{
		AutoPing:    true,
		UseTLS:      true,
		CaptureTags: true, // allows you to access the users DisplayName
	}

	// Create a client without user credentials
	// This means all connections will be read only
	client := twitchirc.NewAnonymousClient(conf)

	conn, _ := client.Connect(&twitchirc.IRCHandler{
		OnPing: func(_ *twitchirc.Connection) {
			fmt.Println("got a ping from twitch!")
		},
	})

	defer conn.Close()

	conn.JoinOne("lirik", &twitchirc.ChannelHandler{
		OnPrivateMessage: func(_ *twitchirc.Connection, m *twitchirc.PrivateMessage) {
			fmt.Printf("%s said '%s'\n", m.User.DisplayName, m.Text)
		},
	})

	conn.Run(context.Background())
}
