package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/silkeh/mumble_bot/matrix"
	"github.com/silkeh/mumble_bot/mumble"
	"github.com/silkeh/mumble_bot/util"
)

func main() {
	var err error
	var mumbleServer, mumbleUser string
	var matrixServer, matrixUser, matrixToken, matrixRoom string

	flag.StringVar(&mumbleServer, "mumble.server", "localhost:64738", "Mumble server")
	flag.StringVar(&mumbleUser, "mumble.user", "MatrixBot", "Mumble user name")
	flag.StringVar(&matrixServer, "matrix.server", "https://matrix.org", "Matrix server")
	flag.StringVar(&matrixUser, "matrix.user", "", "Matrix user name")
	flag.StringVar(&matrixRoom, "matrix.room", "", "Matrix room ID")
	flag.Parse()

	util.SetStringFromEnv(&mumbleServer, "MUMBLE_SERVER")
	util.SetStringFromEnv(&mumbleUser, "MUMBLE_USER")
	util.SetStringFromEnv(&matrixServer, "MATRIX_SERVER")
	util.SetStringFromEnv(&matrixUser, "MATRIX_USER")
	util.SetStringFromEnv(&matrixRoom, "MATRIX_ROOM")
	util.SetStringFromEnv(&matrixToken, "MATRIX_TOKEN")

	if matrixUser == "" || matrixRoom == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Matrix
	mat, err := matrix.NewClient(matrixServer, matrixUser, matrixToken, matrixRoom)
	if err != nil {
		log.Fatal("Error connecting to Matrix:", err)
	}

	// Mumble
	mum, err := mumble.NewClient(mumbleServer, mumbleUser)
	if err != nil {
		log.Fatal("Error connecting to Mumble:", err)
	}
	defer mum.Disconnect()
	mum.Self.SetSelfDeafened(true)

	log.Printf("Waiting for events...")
	for c := range mum.Out {
		var sticker *matrix.Sticker

		switch {
		case c == mumble.Join && len(mum.Users) > 1:
			sticker = matrix.Stickers["welcome"]
		case strings.HasPrefix(c, "!"):
			sticker, _ = matrix.Stickers[c[1:]]
		}

		if sticker != nil {
			_, err := mat.SendSticker(sticker)

			if err != nil {
				log.Printf("Error posting sticker: %s", err)
			}
		}
	}
}
