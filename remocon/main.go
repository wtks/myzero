package main

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/wtks/A75C4269"
	"log"
	"os"
	"strings"
)

type State struct {
	// Mode [off, cooler, heater, dehumidifier]
	Mode string
	// Temp [16~30]
	Temp int
	// Vol [auto, 0, 1, 2, 3, 4, 5]
	Vol string
	// Dir [auto, 1, 2, 3, 4, 5]
	Dir string
}

func (s *State) Convert() (*A75C4269.Controller, error) {
	c := A75C4269.Controller{}

	switch strings.ToLower(s.Mode) {
	case "off":
		c.Power = A75C4269.PowerOff
		return &c, nil
	case "cooler":
		c.Power = A75C4269.PowerOn
		c.Mode = A75C4269.ModeCooler
	case "heater":
		c.Power = A75C4269.PowerOn
		c.Mode = A75C4269.ModeHeater
	case "dehumidifier":
		c.Power = A75C4269.PowerOn
		c.Mode = A75C4269.ModeDehumidifier
	default:
		return nil, fmt.Errorf("unknown mode: %s", s.Mode)
	}

	if s.Temp > 30 || s.Temp < 16 {
		return nil, fmt.Errorf("temp out of range: %d", s.Temp)
	}
	c.PresetTemp = uint(s.Temp)

	switch strings.ToLower(s.Vol) {
	case "auto", "":
		c.AirVolume = A75C4269.AirVolumeAuto
	case "0":
		c.AirVolume = A75C4269.AirVolumeStill
	case "1":
		c.AirVolume = A75C4269.AirVolume1
	case "2":
		c.AirVolume = A75C4269.AirVolume2
	case "3":
		c.AirVolume = A75C4269.AirVolume3
	case "4":
		c.AirVolume = A75C4269.AirVolume4
	case "5":
		c.AirVolume = A75C4269.AirVolumePowerful
	default:
		return nil, fmt.Errorf("invalid vol: %s", s.Vol)
	}

	switch strings.ToLower(s.Dir) {
	case "auto", "":
		c.WindDirection = A75C4269.WindDirectionAuto
	case "1":
		c.WindDirection = A75C4269.WindDirection1
	case "2":
		c.WindDirection = A75C4269.WindDirection2
	case "3":
		c.WindDirection = A75C4269.WindDirection3
	case "4":
		c.WindDirection = A75C4269.WindDirection4
	case "5":
		c.WindDirection = A75C4269.WindDirection5
	default:
		return nil, fmt.Errorf("invalid dir: %s", s.Dir)
	}

	return &c, nil
}

func main() {
	nc, err := nats.Connect(os.Getenv("NATS_SERVER"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	ch := make(chan *nats.Msg)
	sub, err := nc.ChanSubscribe("work.wtks.home.aircon", ch)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	for msg := range ch {
		var s State
		if err := json.Unmarshal(msg.Data, &s); err != nil {
			log.Println(err)
			continue
		}
		q, err := s.Convert()
		if err != nil {
			log.Println(err)
			continue
		}
		b, _ := json.Marshal(q.GetRawSignal())
		if err := nc.Publish("work.wtks.home.ir.send_code", b); err != nil {
			log.Fatal(err)
		}
	}
}
