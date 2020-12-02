# twitchirc

The twitchirc package allows interacting with twitch's IRC server.

This package is still in early development so not all IRC events are implemented yet.

This package is inspired by and uses code from the [go-twitch-irc](https://github.com/gempir/go-twitch-irc) package

## Examples

This creates a simple connection which will print all chat messages in the provided channel.

```go
func main() {
    // Create a config for the client
    conf := &twitchirc.Config{
        AutoPing:          true,
        UseTLS:            true,
    }

    // Create a new client as a anonymous user
    client := twitchirc.NewAnonymousClient(conf)

    // Create a new connection with the default IRC handler
    conn, _ := client.Connect(nil)

    defer conn.Close()

    // Join xqcow with the provided chat handler
    conn.JoinOne("xqcow", &twitchirc.ChannelHandler{
        OnPrivateMessage: func(conn *twitchirc.Connection, msg *twitchirc.PrivateMessage) {
            // Print the message
            fmt.Println(msg.Text)
        },
    })

    // Run the parser with a empty context
    conn.Run(context.Background());
}
```

You can create multiple connections and run them concurrently like this

```go
func main() {
    conf := &twitchirc.Config{
        AutoPing:          true,
        UseTLS:            true,
        CaptureCommands:   true,
        CaptureMembership: true,
        CaptureTags:       true,
    }

    client := twitchirc.NewAnonymousClient(conf)

    channelList := [][]string{
        {"lpl", "DreamHackCS", "Fextralife", "gaules", "csgomc_ru"},
        {"mokrivskyi", "duckmanzch", "stylishnoob4", "Solary", "ritOgaming"},
        {"TheKAIRI78", "fps_shaka", "CohhCarnage", "Esports_Alliance", "x2Twins"},
        {"ratirl", "yamatosdeath1"},
    }

    wg := &sync.WaitGroup{}
    wg.Add(len(channelList))

    context := context.Background()

    for i, v := range channelList {
        i := i // shadow i with a copy of i

        conn, err := client.Connect(&twitchirc.IRCHandler{
            OnPing: func(*twitchirc.Connection) {
                fmt.Printf("#%d: Ping\n", i)
            },
        })

        must(err)

        handler := &twitchirc.ChannelHandler{
            OnPrivateMessage: func(c *twitchirc.Connection, m *twitchirc.PrivateMessage) {
                fmt.Printf("#%d: [%s] %s: %s\n", i, m.Channel, m.User.DisplayName, m.Text)
            },
        }

        if err = conn.Join(v, handler); err != nil {
            log.Fatalln(err)
        }

        go func() {

            defer func() {
                wg.Done()
                conn.Close()
            }()

            if err = conn.Run(context); err != nil {
                log.Fatalln(err)
            }
        }()
    }

    wg.Wait()
}
```

## The `IRCHandler` and `ChannelHandler` handlers

The default `IRCHandler` handles all events which are not related to a specific channel.

The default `ChannelHandler` handles all events which are related to a specific channel.

## Customization with custom handlers

Both the `Connect` and the `Join` method take a `Handler` as an argument.

You can provide your own handlers by implementing the `Handler` interface
on your custom types.
