package twitchirc

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Connection represents a connection to the twitch IRC server.
//
// It holds the active connection created by the Client.Connect() method.
// Additionally its holds all chat handlers.
type Connection struct {
	config *Config

	handlerLock    *sync.RWMutex
	channelHandler map[string]Handler
	ircHandler     Handler

	conn net.Conn
	w    *bufio.Writer
	r    *bufio.Scanner
}

// Run parses the messages from the connection.
//
// Run blocks the current goroutine until ctx is canceled
// or the bufio.Scanner.Scan method returns false.
//
// If this method is cancled it will automatically close the connection.
//
// The messages are read from a buffered reader so the buffer will be overwritten by the connection
// if a provided handler will take too long to proccess.
//
// The reader will wait until the last message was parsed.
func (c *Connection) Run(ctx context.Context) error {
	errCh := make(chan error)

	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		c.Close()
		close(errCh)
	}()

	go func() {
		defer cancel()

		for c.r.Scan() {
			if err := c.handleLine(c.r.Text()); err != nil {
				errCh <- errors.Wrap(err, "connection.Run: could not handle message")
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

// handleLine sends a parsed message to the ircHandler or the chatHandler for the channel.
func (c *Connection) handleLine(line string) error {
	msg, err := parseMessage(line)

	if err != nil {
		return errors.Wrap(err, "connection.handleLine: could not parse message")
	}

	if c.config.AutoPing && msg.Command == "PING" {
		if err := c.sendPong(); err != nil {
			return err
		}
	}

	stream := strings.TrimPrefix(msg.Params[0], "#")

	// The stream is tmi.twitch.tv if its about the IRC connection itself.
	// So we will let the ircHandler worry about that and return early.
	if stream == "tmi.twitch.tv" || msg.Command == "WHISPER" {
		if err = c.ircHandler.HandleIRC(c, msg); err != nil {
			return errors.Wrap(err, "connection.handleLine: could not handle message with the provided irc handler")
		}
	}

	c.handlerLock.RLock()
	defer c.handlerLock.RUnlock()

	if chatHandler, ok := c.channelHandler[stream]; ok {
		if err = chatHandler.HandleIRC(c, msg); err != nil {
			return errors.Wrap(err, "connection.handleLine: could not handle message with the provided chat handler")
		}
	}

	return nil
}

// Join joins the provided channels and attaches the provided handler to the channel.
// If the handler is nil an empty twitchirc.ChannelHandler will be used
//
// If the channel already had a handler it will not be overwritten.
func (c *Connection) Join(channels []string, handler Handler) error {
	if handler == nil {
		handler = &ChannelHandler{}
	}

	c.handlerLock.Lock()
	defer c.handlerLock.Unlock()

	for _, ch := range channels {
		ch = strings.ToLower(ch)

		if _, ok := c.channelHandler[ch]; !ok {
			c.channelHandler[ch] = handler
			if err := c.Write(fmt.Sprintf("JOIN #%s", ch)); err != nil {
				return errors.Wrapf(err, "connection.Join: could not join channel %s", channels)
			}
		}

	}

	return nil
}

// JoinOne is the same as Join but with one channel only.
func (c *Connection) JoinOne(channel string, handler Handler) error {
	return c.Join([]string{channel}, handler)
}

// Depart leaves a channel and removes the handler.
func (c *Connection) Depart(channel string) error {
	channel = strings.ToLower(channel)

	if err := c.Write(fmt.Sprintf("PART #%s", channel)); err != nil {
		return errors.Wrapf(err, "connection.Depart could not depart %s", channel)
	}

	c.handlerLock.Lock()
	delete(c.channelHandler, channel)
	c.handlerLock.Unlock()

	return nil
}

// DepartAll calls Depart for all channels in the channel handler
func (c *Connection) DepartAll() error {
	for v := range c.channelHandler {
		if err := c.Depart(v); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the connection.
//
// This means the underlying net.Conn gets closed.
// So you need to create a new connection if you want to reconnect.
func (c *Connection) Close() error {
	return c.conn.Close()
}

// Say is a wrapper over Write() which allows saying PRIVMSG in the provided channel.
func (c *Connection) Say(channel, text string) error {
	channel = strings.ToLower(channel)
	return c.Write(fmt.Sprintf("PRIVMSG #%s :%s", channel, text))
}

// Write writes message into the connection.
func (c *Connection) Write(message string) error {
	_, err := c.write(message)

	if err != nil {
		return err
	}

	return nil
}

// write writes message into the connection and flushes the buffer.
func (c *Connection) write(message string) (int, error) {
	n, err := c.w.WriteString(fmt.Sprintf("%s\r\n", message))

	if err != nil {
		return 0, errors.Wrapf(err, "connection.write: could not write message %s", message)
	}

	err = c.w.Flush()

	if err != nil {
		return 0, errors.Wrap(err, "connection.write: could not flush buffer")
	}

	return n, nil
}

// sendPong sends a Pong response
func (c *Connection) sendPong() error {
	return c.Write("PONG :tmi.twitch.tv")
}
