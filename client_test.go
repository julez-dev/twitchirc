package twitchirc

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_sendAuth(t *testing.T) {

	clients := []*Client{
		NewClient("testnick", "testpass", &Config{}),
		NewAnonymousClient(&Config{}),
	}

	for _, c := range clients {
		buffer := &bytes.Buffer{}
		c.sendAuth(buffer)

		got := buffer.String()
		expected := fmt.Sprintf("PASS %s\r\nNICK %s\r\n", c.pass, c.nick)

		assert.Equal(t, expected, got, "should be equal")
	}

}

func TestClient_sendCaptures(t *testing.T) {
	table := []struct {
		name    string
		config  *Config
		reqLine string
	}{
		{
			"empty",
			&Config{},
			"",
		},
		{
			"commands",
			&Config{CaptureCommands: true},
			"CAP REQ :twitch.tv/commands\r\n",
		},
		{
			"membership",
			&Config{CaptureMembership: true},
			"CAP REQ :twitch.tv/membership\r\n"},
		{
			"tags",
			&Config{CaptureTags: true},
			"CAP REQ :twitch.tv/tags\r\n",
		},
		{
			"all captures",
			&Config{CaptureTags: true, CaptureCommands: true, CaptureMembership: true},
			"CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership\r\n",
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAnonymousClient(tt.config)

			buffer := &bytes.Buffer{}
			client.sendCaptures(buffer)

			assert.Equal(t, tt.reqLine, buffer.String(), "should be equal")
		})
	}

}
