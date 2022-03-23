package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/silkeh/mumble_bot/mumble"
)

const (
	mumbleRate = 48000
	modemRate  = 8000
)

var atCommands = []string{
	"ATA",          // pick up
	"AT+CPCMFRM=1", // set audio sample rate to 16k
	"AT+CPCMREG=1", // enable USB audio
}

func resampler(in <-chan int16, out chan<- int16) {
	defer close(out)

	mult := int16(mumbleRate / modemRate)
	prev := int16(0)

	for s := range in {
		// Create new samples
		step := (s - prev) / mult
		for i := int16(1); i < mult; i++ {
			out <- prev + (i * step)
		}

		// Write sample to out
		out <- s

		// Store previous
		prev = s
	}
}

func streamFile(f *os.File, out chan<- int16) {
	defer close(out)

	raw := make([]byte, 128)

	for {
		n, err := f.Read(raw)
		if err != nil {
			log.Fatalf("Error reading audio: %s", err)
		}
		for i := 0; i < n; i += 2 {
			out <- int16(binary.LittleEndian.Uint16(raw[i : i+2]))
		}
	}
}

func main() {
	var mumbleServer, mumbleUser, atDevice, audioDevice string

	flag.StringVar(&mumbleServer, "server", "chat.slxh.eu:64738", "Mumble server")
	flag.StringVar(&mumbleUser, "user", "PhoneBot", "Mumble username")
	flag.StringVar(&atDevice, "modem", "/dev/ttyUSB4", "AT modem")
	flag.StringVar(&audioDevice, "audio", "/dev/ttyUSB5", "Audio USB device")
	flag.Parse()

	client, err := mumble.NewClient(mumbleServer, mumbleUser)
	if err != nil {
		log.Fatalf("Error creating client: %s", err)
	}

	atFile, err := os.Open(atDevice)
	if err != nil {
		log.Fatalf("Error opening AT device: %s", err)
	}
	defer atFile.Close()

	audioFile, err := os.Open(audioDevice)
	if err != nil {
		log.Fatalf("Error opening audio device: %s", err)
	}
	defer audioFile.Close()

	inChan := make(chan int16)
	outChan := make(chan int16)

	go resampler(inChan, outChan)
	go streamFile(audioFile, inChan)

	reader := bufio.NewReader(atFile)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading AT device: %s", err)
		}

		log.Printf("AT: %q", line)

		// Pick up on ring
		if strings.TrimSpace(line) == "RING" {
			log.Println("Picking up phone")

			for _, cmd := range atCommands {
				atFile.WriteString(cmd)
			}

			go client.StreamAudio(outChan)
		}

		// Hang up on end
		if strings.HasPrefix(line, "VOICE CALL: END") {
			log.Print("Stopping audio")
			client.StopAudio()
		}
	}
}
