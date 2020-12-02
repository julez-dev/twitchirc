package twitchirc

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
)

const (
	privMSG = "@badge-info=;badges=broadcaster/1;client-nonce=ca3248e0c8cae6f2dcf913ceed1bc6be;color=#FFFFFF;display-name=julezdev;emotes=;flags=;id=5bb550d4-bd15-4a96-9de2-c0298b2d01a9;mod=0;room-id=530594933;subscriber=0;tmi-sent-ts=1591719487292;turbo=0;user-id=530594933;user-type= :julezdev!julezdev@julezdev.tmi.twitch.tv PRIVMSG #julezdev :test"
)

type testHandler struct {
	t    *testing.T
	want string
}

func (t *testHandler) HandleIRC(c *Connection, m *Message) error {
	if m.Message != t.want {
		t.t.Errorf("HandleIRC() = %v, want %v", m.Message, t.want)
	}

	return nil
}

// type testHandlerFail struct {
// 	err error
// }

// func (t *testHandlerFail) HandleIRC(c *Connection, m *Message) error {
// 	return t.err
// }

func TestConnection_Run(t *testing.T) {

	t.Run("simple-priv", func(t *testing.T) {
		server, client := net.Pipe()

		conn := &Connection{
			config:      &Config{},
			r:           bufio.NewScanner(client),
			w:           bufio.NewWriter(client),
			handlerLock: &sync.RWMutex{},
			conn:        server,
		}

		conn.channelHandler = map[string]Handler{
			"julezdev": &testHandler{t: t, want: privMSG},
		}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			fmt.Fprintln(server, privMSG)
			server.Close()
			cancel()
		}()

		if err := conn.Run(ctx); err != nil {
			t.Fatal(err)
		}
	})

	// t.Run("returns-handler-error", func(t *testing.T) {
	// 	server, client := net.Pipe()

	// 	conn := &Connection{
	// 		config: &Config{},
	// 		r:      bufio.NewScanner(client),
	// 		w:      bufio.NewWriter(client),
	// 		conn:   server,
	// 	}

	// 	errWant := errors.New("must fail")

	// 	conn.channelHandler = map[string]Handler{
	// 		"julezdev": &testHandlerFail{err: errWant},
	// 	}

	// 	ctx, cancel := context.WithCancel(context.Background())

	// 	go func() {
	// 		fmt.Fprintln(server, privMSG)
	// 		server.Close()
	// 		cancel()
	// 	}()

	// 	errGot := conn.Run(ctx)
	// 	if !errors.Is(errGot, errWant) {
	// 		t.Errorf("HandleIRC() = %v, want %v", errGot, errWant)
	// 	}

	// })

}

func BenchmarkHandleLine(b *testing.B) {
	line := "@badge-info=;badges=;client-nonce=cb0e017159d5809c5eeffcdb4c6a4c04;color=#FFFFFF;display-name=julezdev;emotes=302213289:5-16;flags=;id=fa8f59a2-aadb-4960-af48-54ef15ee036c;mod=0;room-id=57292293;subscriber=0;tmi-sent-ts=1591714937639;turbo=0;user-id=530594933;user-type= :julezdev!julezdev@julezdev.tmi.twitch.tv PRIVMSG #ratirl :test ratirlPickle"

	conn := Connection{
		config:         &Config{},
		handlerLock:    &sync.RWMutex{},
		channelHandler: map[string]Handler{"ratirl": &ChannelHandler{OnPrivateMessage: func(c *Connection, pm *PrivateMessage) {}}},
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		conn.handleLine(line)
	}
}

func TestConnection_handleLine(t *testing.T) {
	t.Run("got-pong", func(t *testing.T) {
		server, client := net.Pipe()

		conn := &Connection{
			config:      &Config{AutoPing: true},
			r:           bufio.NewScanner(client),
			w:           bufio.NewWriter(client),
			handlerLock: &sync.RWMutex{},
			conn:        server,
		}

		conn.ircHandler = &IRCHandler{}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			fmt.Fprintln(server, "PING :tmi.twitch.tv")

			s := bufio.NewScanner(server)
			s.Scan()

			got := s.Text()
			want := "PONG :tmi.twitch.tv"

			if got != want {
				t.Errorf("ping response = %v, want %v", got, want)
			}

			server.Close()
			cancel()
		}()

		if err := conn.Run(ctx); err != nil {
			t.Fatal(err)
		}

		client.Close()
	})
}
