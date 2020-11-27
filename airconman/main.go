package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type EnvSensor struct {
	T1 float64
	P1 float64
	H1 float64
	T2 float64
	P2 float64
	H2 float64
	L  float64
}

var (
	envValues     EnvSensor
	envValuesLock sync.RWMutex
)

func main() {
	nc, err := nats.Connect(os.Getenv("NATS_SERVER"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	_, err = nc.Subscribe("work.wtks.home.envsensor", func(msg *nats.Msg) {
		var values EnvSensor
		if err := json.Unmarshal(msg.Data, &values); err != nil {
			log.Error(err)
			return
		}
		envValuesLock.Lock()
		envValues = values
		envValuesLock.Unlock()
	})
	if err != nil {
		log.Fatal(err)
	}

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions)
	dg.AddHandler(func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		if msg.Author.ID == s.State.User.ID {
			return
		}

		if !strings.HasPrefix(msg.Content, "/") {
			return
		}

		args := strings.Split(msg.Content, " ")

		switch strings.ToLower(strings.TrimPrefix(args[0], "/")) {
		case "on":
			if len(args) < 2 {
				return
			}

			command := map[string]interface{}{}
			switch strings.ToLower(args[1]) {
			case "h", "heat", "heater":
				command["mode"] = "heater"
				command["temp"] = 26
			case "c", "cool", "cooler":
				command["mode"] = "cooler"
				command["temp"] = 28
			default:
				return
			}

			if len(args) > 2 {
				t, err := strconv.Atoi(args[2])
				if err != nil {
					return
				}
				if t < 16 {
					t = 16
				}
				if t > 30 {
					t = 30
				}
				command["temp"] = t
			}

			b, _ := json.Marshal(command)
			if err := nc.Publish("work.wtks.home.aircon", b); err != nil {
				log.Error(err)
				return
			}
			if err := s.MessageReactionAdd(msg.ChannelID, msg.ID, "🆗"); err != nil {
				log.Error(err)
			}

		case "off":
			if err := nc.Publish("work.wtks.home.aircon", []byte(`{"mode":"off"}`)); err != nil {
				log.Error(err)
				return
			}
			if err := s.MessageReactionAdd(msg.ChannelID, msg.ID, "🆗"); err != nil {
				log.Error(err)
			}

		case "ondo":
			envValuesLock.RLock()
			v := envValues
			envValuesLock.RUnlock()
			if _, err := s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("%.1f ℃ (%.1f %%H)", v.T1, v.H1)); err != nil {
				log.Error(err)
			}
		}
	})

	if err = dg.Open(); err != nil {
		log.Fatal(err)
	}
	defer dg.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
