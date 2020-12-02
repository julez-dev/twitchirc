package twitchirc

// this file was stolen and modified from https://github.com/go-irc/irc/blob/master/parser.go

import (
	"bytes"
	"errors"
	"strings"
)

var tagDecodeSlashMap = map[rune]rune{
	':':  ';',
	's':  ' ',
	'\\': '\\',
	'r':  '\r',
	'n':  '\n',
}

var (
	// ErrZeroLengthMessage is returned when parsing if the input is
	// zero-length.
	ErrZeroLengthMessage = errors.New("irc: cannot parse zero-length message")

	// ErrMissingDataAfterPrefix is returned when parsing if there is
	// no message data after the prefix.
	ErrMissingDataAfterPrefix = errors.New("irc: no message data after prefix")

	// ErrMissingDataAfterTags is returned when parsing if there is no
	// message data after the tags.
	ErrMissingDataAfterTags = errors.New("irc: no message data after tags")

	// ErrMissingCommand is returned when parsing if there is no
	// command in the parsed message.
	ErrMissingCommand = errors.New("irc: missing message command")
)

// TagValue represents the value of a tag.
type TagValue string

// parseTagValue parses a TagValue from the connection. If you need to
// set a TagValue, you probably want to just set the string itself, so
// it will be encoded properly.
func parseTagValue(v string) TagValue {
	ret := &bytes.Buffer{}

	input := bytes.NewBufferString(v)

	for {
		c, _, err := input.ReadRune()
		if err != nil {
			break
		}

		if c == '\\' {
			c2, _, err := input.ReadRune()

			// If we got a backslash then the end of the tag value, we should
			// just ignore the backslash.
			if err != nil {
				break
			}

			if replacement, ok := tagDecodeSlashMap[c2]; ok {
				ret.WriteRune(replacement)
			} else {
				ret.WriteRune(c2)
			}
		} else {
			ret.WriteRune(c)
		}
	}

	return TagValue(ret.String())
}

// Tags represents the IRCv3 message tags.
type Tags map[string]TagValue

// parseTags takes a tag string and parses it into a tag map. It will
// always return a tag map, even if there are no valid tags.
func parseTags(line string) Tags {
	ret := Tags{}

	tags := strings.Split(line, ";")
	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) < 2 {
			ret[parts[0]] = ""
			continue
		}

		ret[parts[0]] = parseTagValue(parts[1])
	}

	return ret
}

// GetTag is a convenience method to look up a tag in the map.
func (t Tags) GetTag(key string) (string, bool) {
	ret, ok := t[key]
	return string(ret), ok
}

// Prefix represents the prefix of a message, generally the user who sent it
type Prefix struct {
	// Name will contain the nick of who sent the message, the
	// server who sent the message, or a blank string
	Name string

	// User will either contain the user who sent the message or a blank string
	User string

	// Host will either contain the host of who sent the message or a blank string
	Host string
}

// parsePrefix takes an identity string and parses it into an
// identity struct. It will always return an Prefix struct and never
// nil.
func parsePrefix(line string) *Prefix {
	// Start by creating an Prefix with nothing but the host
	id := &Prefix{
		Name: line,
	}

	uh := strings.SplitN(id.Name, "@", 2)
	if len(uh) == 2 {
		id.Name, id.Host = uh[0], uh[1]
	}

	nu := strings.SplitN(id.Name, "!", 2)
	if len(nu) == 2 {
		id.Name, id.User = nu[0], nu[1]
	}

	return id
}

// Message represents a line parsed from the server
type Message struct {
	// Each message can have IRCv3 tags
	Tags

	// Each message can have a Prefix
	*Prefix

	// Command is which command is being called.
	Command string

	// Params are all the arguments for the command.
	Params []string

	Message string
}

// mustParseMessage calls ParseMessage and either returns the message
// or panics if an error is returned.
func mustParseMessage(line string) *Message {
	m, err := parseMessage(line)
	if err != nil {
		panic(err.Error())
	}
	return m
}

// parseMessage takes a message string (usually a whole line) and
// parses it into a Message struct. This will return nil in the case
// of invalid messages.
func parseMessage(line string) (*Message, error) {
	// Trim the line and make sure we have data
	line = strings.TrimRight(line, "\r\n")
	if len(line) == 0 {
		return nil, ErrZeroLengthMessage
	}

	c := &Message{
		Tags:    Tags{},
		Prefix:  &Prefix{},
		Message: line,
	}

	if line[0] == '@' {
		loc := strings.Index(line, " ")
		if loc == -1 {
			return nil, ErrMissingDataAfterTags
		}

		c.Tags = parseTags(line[1:loc])
		line = line[loc+1:]
	}

	if line[0] == ':' {
		loc := strings.Index(line, " ")
		if loc == -1 {
			return nil, ErrMissingDataAfterPrefix
		}

		// Parse the identity, if there was one
		c.Prefix = parsePrefix(line[1:loc])
		line = line[loc+1:]
	}

	// Split out the trailing then the rest of the args. Because
	// we expect there to be at least one result as an arg (the
	// command) we don't need to special case the trailing arg and
	// can just attempt a split on " :"
	split := strings.SplitN(line, " :", 2)
	c.Params = strings.FieldsFunc(split[0], func(r rune) bool {
		return r == ' '
	})

	// If there are no args, we need to bail because we need at
	// least the command.
	if len(c.Params) == 0 {
		return nil, ErrMissingCommand
	}

	// If we had a trailing arg, append it to the other args
	if len(split) == 2 {
		c.Params = append(c.Params, split[1])
	}

	// Because of how it's parsed, the Command will show up as the
	// first arg.
	c.Command = strings.ToUpper(c.Params[0])
	c.Params = c.Params[1:]

	// If there are no params, set it to nil, to make writing tests and other
	// things simpler.
	if len(c.Params) == 0 {
		c.Params = nil
	}

	return c, nil
}
