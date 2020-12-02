package twitchirc

import "github.com/pkg/errors"

// Handler is a generic interface which provides the parsed IRC message from the reader.
// It allows to customize the behavior on how a IRC message is handled.
//
// One way to use this interface is to provide multiple callbacks which then get
// called by the implemented HandleIRC method.
type Handler interface {
	HandleIRC(*Connection, *Message) error
}

// ChannelHandler is a default implementation of Handler which holds all callback functions for chat events.
//
// It provides multiple callbacks for various chat events which occur in a chat room.
// This includes, but is not limited to, normales chat messages, timeouts and so on.
type ChannelHandler struct {
	OnPrivateMessage   func(*Connection, *PrivateMessage)
	OnClearchatMessage func(*Connection, *ClearChatMessage)
}

// HandleIRC parses the message to a specialized struct and calls the corresponding
// callback function.
func (ch *ChannelHandler) HandleIRC(conn *Connection, msg *Message) error {
	switch msg.Command {
	case "PRIVMSG":
		if ch.OnPrivateMessage != nil {
			privMSG, err := parsePrivateMessage(msg)

			if err != nil {
				return errors.Wrapf(err, "chatHandler.HandleIRC: could not parse privmsg: %#v", msg)
			}

			ch.OnPrivateMessage(conn, privMSG)
		}

	case "CLEARCHAT":
		if ch.OnClearchatMessage != nil {
			clearchatMSG, err := parseClearChat(msg)

			if err != nil {
				return errors.Wrapf(err, "chatHandler.HandleIRC: could not parse clearchat: %#v", msg)
			}

			ch.OnClearchatMessage(conn, clearchatMSG)
		}
	}

	return nil
}

// IRCHandler is a default implementation of Handler which holds all callback functions for gerneral IRC Events.
//
// Only callback functions which are not limited to specific rooms are provided here.
type IRCHandler struct {
	OnPing    func(*Connection)
	OnWhisper func(*Connection, *WhisperMessage)
}

// HandleIRC parses the message to a specialized struct and calls the corresponding
// callback function.
func (h *IRCHandler) HandleIRC(conn *Connection, msg *Message) error {
	switch msg.Command {

	case "PING":
		if h.OnPing != nil {
			h.OnPing(conn)
		}

	case "WHISPER":
		if h.OnWhisper != nil {
			whisperMSG, err := parseWhisper(msg)

			if err != nil {
				return errors.Wrapf(err, "IRCHandler.HandleIRC: could not parse whisper: %#v", msg)
			}

			h.OnWhisper(conn, whisperMSG)
		}
	}

	return nil
}
