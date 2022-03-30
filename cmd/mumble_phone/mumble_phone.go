package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/serial"
	"layeh.com/gumble/gumble"

	"github.com/silkeh/mumble_bot/audio"
	"github.com/silkeh/mumble_bot/mumble"
)

const (
	atBaudRate    = 115200
	audioBaudRate = 3000000
	mumbleRate    = gumble.AudioSampleRate
	modemRate     = 8000
)

var connectcommands = []string{
	"ATE0",                                  // disable echo
	"AT+IPR=" + strconv.Itoa(audioBaudRate), // set baud rate
}

var pickupCommands = []string{
	"ATA", // pick up
}

var startCommands = []string{
	"AT+CPCMREG=1", // enable USB audio
}

var endCommands = []string{
	"AT+CPCMREG=0", // disable USB audio
}

type AT struct {
	io.ReadWriteCloser
	r *bufio.Reader
}

func NewAT(f io.ReadWriteCloser) *AT {
	return &AT{
		ReadWriteCloser: f,
		r:               bufio.NewReader(f),
	}
}

func (at *AT) Exec(cmd string) {
	log.Printf("Executing AT command: %s", cmd)

	_, err := at.Write([]byte(cmd + "\r\n"))
	if err != nil {
		log.Fatalf("Error writing to AT device: %s", err)
	}

	log.Printf("Response: %s", at.Readline())
}

func (at *AT) Readline() string {
	line, err := at.r.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading AT device: %s", err)
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return at.Readline()
	}

	return line
}

func main() {
	var mumbleServer, mumbleUser, atDevice, audioDevice string
	var rings, ringMax int

	flag.StringVar(&mumbleServer, "server", "chat.slxh.eu:64738", "Mumble server")
	flag.StringVar(&mumbleUser, "user", "PhoneBot", "Mumble username")
	flag.StringVar(&atDevice, "modem", "/dev/ttyUSB3", "AT modem")
	flag.StringVar(&audioDevice, "audio", "/dev/ttyUSB4", "Audio USB device")
	flag.IntVar(&ringMax, "rings", 1, "Number of rings before pick up")
	flag.Parse()

	atFile, err := serial.OpenPort(&serial.Config{Name: atDevice, Baud: atBaudRate})
	if err != nil {
		log.Fatalf("Error opening AT device: %s", err)
	}
	defer atFile.Close()

	at := NewAT(atFile)

	for _, cmd := range connectcommands {
		at.Exec(cmd)
	}

	audioFile, err := serial.OpenPort(&serial.Config{Name: audioDevice, Baud: audioBaudRate})
	if err != nil {
		log.Fatal(err)
	}
	defer audioFile.Close()

	streamer := new(mumble.AudioStreamer)
	phoneIn := make(chan int16)
	phoneOut := make(chan int16)
	mumbleIn := make(chan int16)
	mumbleOut := make(chan int16)

	// Resample phone input and send to mumble
	go audio.UpSample(int16(mumbleRate/modemRate), phoneIn, mumbleIn)

	// Resample mumble input and send to phone
	go audio.DownSample(mumbleRate/modemRate, mumbleOut, phoneOut)

	// Stream phone audio from/to file
	go audio.Read(audioFile, phoneIn)
	go audio.Write(audioFile, phoneOut)

	// Stream mumble audio to phone
	go streamer.Stream(mumbleOut)

	var client *mumble.Client

	log.Print("Starting")
	for {
		line := at.Readline()

		log.Printf("AT: %q", line)

		// Pick up on ring
		if line == "RING" {
			if rings < ringMax {
				rings++
				continue
			}
			client, err = mumble.NewClient(mumbleServer, mumbleUser, streamer)
			if err != nil {
				log.Fatalf("Error creating client: %s", err)
			}

			log.Println("Picking up phone")
			for _, cmd := range pickupCommands {
				at.Exec(cmd)
			}

			// Read second response
			log.Printf("AT: %s", at.Readline())

			time.Sleep(100 * time.Millisecond)

			for _, cmd := range startCommands {
				at.Exec(cmd)
			}

			go client.StreamAudio(mumbleIn)
		}

		// Hang up on end
		if strings.HasPrefix(line, "VOICE CALL: END") {
			log.Print("Stopping audio and Mumble client")

			time.Sleep(1000 * time.Millisecond)

			for _, cmd := range endCommands {
				at.Exec(cmd)
			}

			if client != nil {
				client.StopAudio()
				client.Disconnect()
				client = nil
			}

			rings = 0

			log.Print("Audio and Mumble client stopped")
		}
	}
}
