package twitchirc

import (
	"strconv"
	"strings"
	"time"
)

// User represents a twitch irc user.
type User struct {
	ID          string
	DisplayName string
	Name        string
	Color       string
	Badges      map[string]int
}

// Emote represents a twitch emote.
type Emote struct {
	Name  string
	ID    string
	Count int
}

// parseEmotes creates a slice of emotes from emoteString.
// It uses the message to slice the emote name out of the message.
func parseEmotes(emoteString, message string) []*Emote {
	emotes := []*Emote{}

	if emoteString == "" {
		return emotes
	}

	runes := []rune(message)

	for _, v := range strings.Split(emoteString, "/") {
		split := strings.SplitN(v, ":", 2)
		pairs := strings.SplitN(split[1], ",", 2)
		pair := strings.SplitN(pairs[0], "-", 2)

		firstIndex, _ := strconv.Atoi(pair[0])
		lastIndex, _ := strconv.Atoi(pair[1])

		if lastIndex+1 > len(runes) {
			lastIndex--
		}

		emote := &Emote{
			Name:  string(runes[firstIndex : lastIndex+1]),
			ID:    split[0],
			Count: strings.Count(split[1], ",") + 1,
		}

		emotes = append(emotes, emote)
	}

	return emotes
}

// parseTime converts the twitch tag time to golang's time.Time
func parseTime(timeTag string) time.Time {
	if timeTag == "" {
		return time.Time{}
	}

	time64, _ := strconv.ParseInt(timeTag, 10, 64)
	return time.Unix(0, int64(time64*1e6))
}

// parseBadges parses badges from badgesTag.
func parseBadges(badgesTag string) map[string]int {
	badges := make(map[string]int)

	for _, badge := range strings.Split(badgesTag, ",") {
		pair := strings.SplitN(badge, "/", 2)
		badges[pair[0]], _ = strconv.Atoi(pair[1])
	}

	return badges
}

// Ping related stuff

// PingMessage represents a parsed PING.
type PingMessage struct {
	Raw *Message
}

func parsePing(message *Message) *PingMessage {
	return &PingMessage{
		Raw: message,
	}
}

// PRIVMSG related stuff

// PrivateMessage represents a parsed PRIVMSG.
type PrivateMessage struct {
	ID      string
	RoomID  string
	User    *User
	Emotes  []*Emote
	Text    string
	Channel string
	Time    time.Time

	Raw *Message
}

func parsePrivateMessage(message *Message) (*PrivateMessage, error) {
	privateMessage := &PrivateMessage{
		User: &User{},
		Raw:  message,
	}

	if badges, ok := message.GetTag("badges"); ok {
		if badges != "" {
			privateMessage.User.Badges = parseBadges(badges)
		}
	}

	if emotes, ok := message.GetTag("emotes"); ok {
		if emotes != "" {
			privateMessage.Emotes = parseEmotes(emotes, message.Params[1])
		}
	}

	if displayName, ok := message.GetTag("display-name"); ok {
		privateMessage.User.DisplayName = displayName
	}

	if userID, ok := message.GetTag("user-id"); ok {
		privateMessage.User.ID = userID
	}

	if messageID, ok := message.GetTag("id"); ok {
		privateMessage.ID = messageID
	}

	if roomID, ok := message.GetTag("room-id"); ok {
		privateMessage.RoomID = roomID
	}

	if color, ok := message.GetTag("color"); ok {
		privateMessage.User.Color = color
	}

	if time, ok := message.GetTag("tmi-sent-ts"); ok {
		privateMessage.Time = parseTime(time)
	}

	privateMessage.User.Name = message.User

	privateMessage.Channel = strings.TrimPrefix(message.Params[0], "#")
	privateMessage.Text = message.Params[1]

	return privateMessage, nil
}

// WhisperMessage represents a parsed whisper.
type WhisperMessage struct {
	MessageID string
	ThreadID  string
	User      *User
	Emotes    []*Emote
	Text      string

	Raw *Message
}

func parseWhisper(message *Message) (*WhisperMessage, error) {
	whisperMessage := &WhisperMessage{
		User: &User{},
		Raw:  message,
	}

	if badges, ok := message.GetTag("badges"); ok {
		if badges != "" {
			whisperMessage.User.Badges = parseBadges(badges)
		}
	}

	if emotes, ok := message.GetTag("emotes"); ok {
		if emotes != "" {
			whisperMessage.Emotes = parseEmotes(emotes, message.Params[1])
		}
	}

	if displayName, ok := message.GetTag("display-name"); ok {
		whisperMessage.User.DisplayName = displayName
	}

	if userID, ok := message.GetTag("user-id"); ok {
		whisperMessage.User.ID = userID
	}

	if color, ok := message.GetTag("color"); ok {
		whisperMessage.User.Color = color
	}

	if id, ok := message.GetTag("message-id"); ok {
		whisperMessage.MessageID = id
	}

	if threadID, ok := message.GetTag("thread-id"); ok {
		whisperMessage.ThreadID = threadID
	}

	whisperMessage.Text = message.Params[1]
	whisperMessage.User.Name = message.User

	return whisperMessage, nil
}

// ClearChatMessage represents a parsed clearchat.
type ClearChatMessage struct {
	TargetUserID string
	RoomID       string
	TargetUser   string
	Time         time.Time
	BanDuration  int
	Channel      string

	Raw *Message
}

func parseClearChat(message *Message) (*ClearChatMessage, error) {
	clearchatMessage := &ClearChatMessage{
		Raw: message,
	}

	if targetUserID, ok := message.GetTag("target-user-id"); ok {
		clearchatMessage.TargetUserID = targetUserID
	}

	if time, ok := message.GetTag("tmi-sent-ts"); ok {
		clearchatMessage.Time = parseTime(time)
	}

	if rawBanDuration, ok := message.GetTag("ban-duration"); ok {
		duration, _ := strconv.Atoi(rawBanDuration)
		clearchatMessage.BanDuration = duration
	}

	if roomID, ok := message.GetTag("room-id"); ok {
		clearchatMessage.RoomID = roomID
	}

	if len(message.Params) > 1 {
		clearchatMessage.Channel = strings.TrimLeft(message.Params[0], "#")
		clearchatMessage.TargetUser = message.Params[1]
	}

	return clearchatMessage, nil
}
