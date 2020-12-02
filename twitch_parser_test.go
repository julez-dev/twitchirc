package twitchirc

import (
	"reflect"
	"testing"
	"time"
)

const (
	whisperEmote      = "@badges=;color=#FFFFFF;display-name=shaymin_fakezz;emotes=302213289:0-11;message-id=7;thread-id=61083508_530594933;turbo=0;user-id=61083508;user-type= :shaymin_fakezz!shaymin_fakezz@shaymin_fakezz.tmi.twitch.tv WHISPER julezdev :ratirlPickle test"
	whisperEmoteBadge = "@badges=glhf-pledge/1;color=#FFFFFF;display-name=shaymin_fakezz;emotes=302213289:0-11;message-id=7;thread-id=61083508_530594933;turbo=0;user-id=61083508;user-type= :shaymin_fakezz!shaymin_fakezz@shaymin_fakezz.tmi.twitch.tv WHISPER julezdev :ratirlPickle test"
	privEmote         = "@badge-info=;badges=broadcaster/1;client-nonce=ca3248e0c8cae6f2dcf913ceed1bc6be;color=#FFFFFF;display-name=julezdev;emotes=302213289:0-11,27-38/302242139:13-25;flags=;id=5bb550d4-bd15-4a96-9de2-c0298b2d01a9;mod=0;room-id=530594933;subscriber=0;tmi-sent-ts=1591719487292;turbo=0;user-id=530594933;user-type= :julezdev!julezdev@julezdev.tmi.twitch.tv PRIVMSG #julezdev :ratirlPickle ratirlPopcorn ratirlPickle test test"
	timeout           = "@ban-duration=10;room-id=530594933;target-user-id=12427;tmi-sent-ts=1591726290782 :tmi.twitch.tv CLEARCHAT #julezdev :test"
	ban               = "@room-id=530594933;target-user-id=19510;tmi-sent-ts=1591726324865 :tmi.twitch.tv CLEARCHAT #julezdev :bla"
)

func Test_parseBadges(t *testing.T) {
	type args struct {
		rawBadges string
	}
	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		{
			name: "parse-badge-subscriber-6",
			args: args{rawBadges: "subscriber/6"},
			want: map[string]int{"subscriber": 6},
		},
		{
			name: "parse-badge-overwatch-league-insider-sub-18",
			args: args{rawBadges: "subscriber/18,overwatch-league-insider_2019A/1"},
			want: map[string]int{"overwatch-league-insider_2019A": 1, "subscriber": 18},
		},
		{
			name: "parse-badge-premium-sub-12",
			args: args{rawBadges: "subscriber/12,premium/1"},
			want: map[string]int{"premium": 1, "subscriber": 12},
		},
		{
			name: "parse-badge-glhf-pledge",
			args: args{rawBadges: "glhf-pledge/1"},
			want: map[string]int{"glhf-pledge": 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBadges(tt.args.rawBadges); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseBadges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkParsePrivateMessage(b *testing.B) {
	line := privEmote

	msg := mustParseMessage(line)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		parsePrivateMessage(msg)
	}
}

func BenchmarkParseWhisper(b *testing.B) {
	line := whisperEmote

	msg := mustParseMessage(line)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		parseWhisper(msg)
	}
}

func BenchmarkParseClearchat(b *testing.B) {
	line := "@ban-duration=20;room-id=530594933;target-user-id=61083508;tmi-sent-ts=1591716679276 :tmi.twitch.tv CLEARCHAT #julezdev :shaymin_fakezz"

	msg := mustParseMessage(line)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		parseClearChat(msg)
	}
}

func Test_parseTime(t *testing.T) {

	table := []struct {
		name string
		got  string
		want time.Time
	}{
		{
			"no-time",
			"",
			time.Time{},
		},
		{
			"valid-time",
			"1591716679276",
			time.Unix(0, int64(1591716679276*1e6)),
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTime(tt.got); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEmotes(t *testing.T) {
	type args struct {
		emoteString string
		message     string
	}

	tests := []struct {
		name string
		args args
		want []*Emote
	}{
		{
			name: "no-emote",
			args: args{
				emoteString: "",
				message:     "@badge-info=;badges=broadcaster/1;client-nonce=bf1ba3b89b1d64b27fdfe3668f397d8d;color=#FFFFFF;display-name=julezdev;emotes=;flags=;id=6d0dfe2a-341f-4683-849d-c12b4432645f;mod=0;room-id=530594933;subscriber=0;tmi-sent-ts=1591718544125;turbo=0;user-id=530594933;user-type= :julezdev!julezdev@julezdev.tmi.twitch.tv PRIVMSG #julezdev :test",
			},
			want: []*Emote{},
		},
		{
			name: "emotes-and-text",
			args: args{
				emoteString: "302213289:0-11,27-38/302242139:13-25",
				message:     "ratirlPickle ratirlPopcorn ratirlPickle test test",
			},
			want: []*Emote{
				{ID: "302213289", Count: 2, Name: "ratirlPickle"},
				{ID: "302242139", Count: 1, Name: "ratirlPopcorn"},
			},
		},
		{
			name: "emote-last",
			args: args{
				emoteString: "1035663:4-7",
				message:     "bla xqcL",
			},
			want: []*Emote{
				{ID: "1035663", Count: 1, Name: "xqcL"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseEmotes(tt.args.emoteString, tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEmotes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePing(t *testing.T) {
	type args struct {
		message *Message
	}
	tests := []struct {
		name string
		args args
		want *PingMessage
	}{
		{
			name: "empty-message",
			args: args{message: &Message{}},
			want: &PingMessage{Raw: &Message{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePing(tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePrivateMessage(t *testing.T) {

	type args struct {
		message *Message
	}

	tests := []struct {
		name    string
		args    args
		want    *PrivateMessage
		wantErr bool
	}{
		{
			name: "as-broadcaster",
			args: args{
				message: mustParseMessage(privEmote),
			},
			want: &PrivateMessage{
				Channel: "julezdev",
				Emotes: []*Emote{
					{ID: "302213289", Count: 2, Name: "ratirlPickle"},
					{ID: "302242139", Count: 1, Name: "ratirlPopcorn"},
				},
				ID:     "5bb550d4-bd15-4a96-9de2-c0298b2d01a9",
				Raw:    &Message{},
				RoomID: "530594933",
				Text:   "ratirlPickle ratirlPopcorn ratirlPickle test test",
				Time:   time.Unix(0, int64(1591719487292*1e6)),
				User: &User{
					Color:       "#FFFFFF",
					Badges:      map[string]int{"broadcaster": 1},
					DisplayName: "julezdev",
					ID:          "530594933",
					Name:        "julezdev",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.Raw = tt.args.message

			got, err := parsePrivateMessage(tt.args.message)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePrivateMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePrivateMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseWhisper(t *testing.T) {
	type args struct {
		message *Message
	}
	tests := []struct {
		name    string
		args    args
		want    *WhisperMessage
		wantErr bool
	}{
		{
			name: "with-emote",
			args: args{
				mustParseMessage(whisperEmote),
			},
			want: &WhisperMessage{
				MessageID: "7",
				ThreadID:  "61083508_530594933",
				User: &User{
					ID:          "61083508",
					DisplayName: "shaymin_fakezz",
					Name:        "shaymin_fakezz",
					Color:       "#FFFFFF",
				},
				Emotes: []*Emote{
					{
						ID:    "302213289",
						Name:  "ratirlPickle",
						Count: 1,
					},
				},
				Text: "ratirlPickle test",
			},
		},
		{
			name: "with-emote-badge",
			args: args{
				mustParseMessage(whisperEmoteBadge),
			},
			want: &WhisperMessage{
				MessageID: "7",
				ThreadID:  "61083508_530594933",
				User: &User{
					ID:          "61083508",
					DisplayName: "shaymin_fakezz",
					Name:        "shaymin_fakezz",
					Color:       "#FFFFFF",
					Badges:      map[string]int{"glhf-pledge": 1},
				},
				Emotes: []*Emote{
					{
						ID:    "302213289",
						Name:  "ratirlPickle",
						Count: 1,
					},
				},
				Text: "ratirlPickle test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.Raw = tt.args.message

			got, err := parseWhisper(tt.args.message)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseWhisper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseWhisper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseClearChat(t *testing.T) {
	type args struct {
		message *Message
	}

	tests := []struct {
		name    string
		args    args
		want    *ClearChatMessage
		wantErr bool
	}{
		{
			name: "timeout-10",
			args: args{
				message: mustParseMessage(timeout),
			},
			want: &ClearChatMessage{
				Channel:      "julezdev",
				BanDuration:  10,
				RoomID:       "530594933",
				TargetUser:   "test",
				TargetUserID: "12427",
				Time:         time.Unix(0, int64(1591726290782*1e6)),
			},
		},
		{
			name: "ban",
			args: args{
				message: mustParseMessage(ban),
			},
			want: &ClearChatMessage{
				Channel:      "julezdev",
				RoomID:       "530594933",
				TargetUser:   "bla",
				TargetUserID: "19510",
				Time:         time.Unix(0, int64(1591726324865*1e6)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.Raw = tt.args.message

			got, err := parseClearChat(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseClearChat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseClearChat() = %v, want %v", got, tt.want)
			}
		})
	}
}
