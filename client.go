// Package twitchirc allows you to interact with the twitch irc server.
package twitchirc

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	chatTLS     = "irc.chat.twitch.tv:6697"
	chatNoneTLS = "irc.chat.twitch.tv:6667"

	// anonymousNick is a anonymous nickname for twitch irc.
	anonymousNick = "justinfan123123"
	// anonymousPass is a anonymous oauth for twitch irc.
	anonymousPass = "oauth:123123123"
)

// Config holds the configuration for the client
type Config struct {
	UseTLS            bool
	AutoPing          bool
	CaptureTags       bool
	CaptureCommands   bool
	CaptureMembership bool
}

// Client holds a client which allows creating connections to the twitch irc servers
type Client struct {
	nick   string
	pass   string
	config *Config
}

// NewClient returns a new client with the provided config
//
// nick is the username
// pass is the auth pass which starts with oauth:
func NewClient(nick string, pass string, conf *Config) *Client {
	return &Client{
		nick:   nick,
		pass:   pass,
		config: conf,
	}
}

// NewAnonymousClient returns a new anonymous client with the provided config
//
// With this Client you can create a read only connection.
// You can't use this to write in rooms.
func NewAnonymousClient(conf *Config) *Client {
	return &Client{
		nick:   anonymousNick,
		pass:   anonymousPass,
		config: conf,
	}
}

// Connect dials the twitch IRC servers and creates a new connection
// which allows interacting with the IRC servers.
//
// The handler interface is used to handle all general IRC events like pings and whispers.
// The handler does not get called for events which occur in a specific channel.
// If the handler is nil an empty twitchirc.IRCHandler will be used
//
// Use the connections JoinChannels method to join a channel with a provided handler struct
// if you want to handle room specific events.
//
// This method creates a net.Conn which could leak if the returned connection
// does not get a chance to close the connection.
func (c *Client) Connect(ircHandler Handler) (*Connection, error) {

	if ircHandler == nil {
		ircHandler = &IRCHandler{}
	}

	dialer := &net.Dialer{
		KeepAlive: time.Second * 10,
	}

	var (
		conn net.Conn
		err  error
	)

	if c.config.UseTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", chatTLS, &tls.Config{})
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		conn, err = dialer.DialContext(ctx, "tcp", chatNoneTLS)
	}

	if err != nil {
		return nil, errors.Wrap(err, "client.Connect: could not dial twitch server")
	}

	r := bufio.NewScanner(conn)
	w := bufio.NewWriter(conn)

	if err = c.sendAuth(w); err != nil {
		return nil, errors.Wrap(err, "client.Connect: could not send authentication")
	}

	if err = c.sendCaptures(w); err != nil {
		return nil, errors.Wrap(err, "connection.Connect: could not send irc captures")
	}

	connection := &Connection{
		channelHandler: make(map[string]Handler),
		ircHandler:     ircHandler,
		handlerLock:    &sync.RWMutex{},
		config:         c.config,
		conn:           conn,
		r:              r,
		w:              w,
	}

	return connection, nil
}

// sendAuth sends the authentication messages into w
func (c *Client) sendAuth(w io.Writer) error {
	auth := fmt.Sprintf("PASS %s\r\nNICK %s\r\n", c.pass, c.nick)
	_, err := w.Write([]byte(auth))

	if err != nil {
		return errors.Wrap(err, "client.sendAuth: could not send auth")
	}

	return nil
}

// sendCaptures sends the capture messages into w
func (c *Client) sendCaptures(w io.Writer) error {
	captures := []string{}
	if c.config.CaptureTags {
		captures = append(captures, "twitch.tv/tags")
	}
	if c.config.CaptureCommands {
		captures = append(captures, "twitch.tv/commands")
	}
	if c.config.CaptureMembership {
		captures = append(captures, "twitch.tv/membership")
	}

	if len(captures) > 0 {
		captureString := strings.Join(captures, " ")
		captureString = fmt.Sprintf("CAP REQ :%s\r\n", captureString)

		_, err := w.Write([]byte(captureString))

		if err != nil {
			return errors.Wrap(err, "client.sendCaptures: could not send captures")
		}
	}

	return nil
}
