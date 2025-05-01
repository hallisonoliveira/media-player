package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const PLAYER_TOPIC = "player"

// Tipos de estado e ação
type PlayerState string
type Action string

const (
	Idle    PlayerState = "idle"
	Loaded  PlayerState = "loaded"
	Playing PlayerState = "playing"
	Paused  PlayerState = "paused"

	Load     Action = "load"
	Unload   Action = "unload"
	Play     Action = "play"
	Pause    Action = "pause"
	Stop     Action = "stop"
	Next     Action = "next"
	Previous Action = "previous"
)

// Payload serializável
type PlayerStatePayload struct {
	State   PlayerState `json:"state"`
	TrackID string      `json:"track_id,omitempty"`
}

// Transições válidas
var transitions = map[PlayerState]map[Action]PlayerState{
	Idle: {
		Load: Loaded,
	},
	Loaded: {
		Unload:   Idle,
		Play:     Playing,
		Next:     Loaded,
		Previous: Loaded,
	},
	Playing: {
		Stop:     Loaded,
		Next:     Loaded,
		Previous: Loaded,
		Pause:    Paused,
	},
	Paused: {
		Play:     Playing,
		Stop:     Loaded,
		Next:     Loaded,
		Previous: Loaded,
	},
}

type Time struct {
	Current string `json:current`
	Total   string `json:total`
}

type Data struct {
	Title  string `json:title`
	Artist string `json:artist`
}

type Media struct {
	Type  string      `json:type`
	State PlayerState `json:"state"`
	Time  Time        `json:time`
	Data  Data        `json:data`
}

type PlayerData struct {
	Timestamp string `json:timestamp`
	Media     Media  `json:media`
}

func NewPlayerData(tp string) *PlayerData {
	timestamp := time.Now().In(time.FixedZone("GMT-3", -3*60*60)).Format("02-01-2006T15:04:05.000")
	time := Time{
		Current: "",
		Total:   "",
	}
	data := Data{
		Title:  "",
		Artist: "",
	}
	media := Media{
		Type:  tp,
		State: PlayerState(Idle),
		Time:  time,
		Data:  data,
	}

	return &PlayerData{
		Timestamp: timestamp,
		Media:     media,
	}
}

func (data *PlayerData) UpdateState(action Action) error {
	transitionsForState, ok := transitions[data.Media.State]
	if !ok {
		return errors.New("Unknown state")
	}
	nextState, ok := transitionsForState[action]
	if !ok {
		return fmt.Errorf("ação %s inválida no estado %s", action, data.Media.State)
	}
	data.Media.State = nextState
	return nil
}

func (data *PlayerData) UpdateCurrentTime(currentTime string) {
	data.Media.Time.Current = currentTime
}

func (playerData *PlayerData) UpdateMediaData(artist string, title string) {
	data := Data{
		Artist: artist,
		Title:  title,
	}
	playerData.Media.Data = data
}

func main() {
	log.SetFlags(0)

	fmt.Println("Starting")

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	defer stop()

	run(ctx)
}

func run(ctx context.Context) {
	go playMp3("/home/hallison/sample.mp3")
	<-ctx.Done()
	fmt.Println("Shutting down...")
}

func printPlayerDataJson(playerData PlayerData) {
	playerDataJson, err := json.Marshal(playerData)
	if err != nil {
		log.Printf("Error serializing DateTime: %v", err)
	}
	fmt.Printf("Payload: %v\n", string(playerDataJson))
}

func playMp3(path string) {
	playerData := NewPlayerData("MP3")
	printPlayerDataJson(*playerData)

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Fail when opening MP3 file:", err)
		return
	}
	defer file.Close()

	if err = playerData.UpdateState(Load); err != nil {
		fmt.Println("Unable to update state:", err)
		return
	}

	playerData.UpdateState(Load)
	printPlayerDataJson(*playerData)

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		fmt.Println("Erro ao ler metadados:", err)
		return
	}

	playerData.UpdateMediaData(metadata.Artist(), metadata.Title())
	printPlayerDataJson(*playerData)

	file.Seek(0, 0) // Reset pointer before decoding
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		fmt.Println("Fail when decoding MP3 file:", err)
		return
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	positioner, ok := streamer.(beep.StreamSeekCloser)
	if !ok {
		fmt.Println("Streamer does not support seeking or position")
		return
	}
	done := make(chan bool)

	len := positioner.Len()
	total := time.Duration(len) * time.Second / time.Duration(format.SampleRate)
	playerData.Media.Time.Total = total.String()

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	playerData.UpdateState(Play)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pos := positioner.Position()
			elapsed := time.Duration(pos) * time.Second / time.Duration(format.SampleRate)
			playerData.UpdateCurrentTime(elapsed.String())
			printPlayerDataJson(*playerData)

		case <-done:
			playerData.UpdateState(Stop)
			printPlayerDataJson(*playerData)
			return
		}
	}
}
