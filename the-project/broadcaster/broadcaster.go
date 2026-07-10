package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/nats-io/nats.go"
)

func readEnv(env, fallback string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	} else {
		return fallback
	}
}

func main() {
	natsUrl := readEnv("NATS_URL", "nats://localhost:4222")
	token := readEnv("DISCORD_TOKEN", "")
	channel := readEnv("DISCORD_CHANNEL_ID", "")
	sendExternal := token != "" && channel != ""

	nc, err := nats.Connect(natsUrl)
	if err != nil {
		panic(err)
	}

	var dc *discordgo.Session

	if sendExternal {
		dc, err = discordgo.New("Bot " + token)
		if err != nil {
			panic(err)
		}

		err = dc.Open()
		if err != nil {
			panic(err)
		}
		defer dc.Close()
	}

	nc.QueueSubscribe("created_todo", "broadcasters", func(msg *nats.Msg) {
		str := fmt.Sprintf("created todo: %s", msg.Data)
		fmt.Println(str)
		if sendExternal {
			_, err = dc.ChannelMessageSend(channel, str)
		}
	})

	nc.QueueSubscribe("marked_done", "broadcasters", func(msg *nats.Msg) {
		str := fmt.Sprintf("marked done: %s", msg.Data)
		fmt.Println(str)
		if sendExternal {
			_, err = dc.ChannelMessageSend(channel, str)
		}
	})

	select {}
}
