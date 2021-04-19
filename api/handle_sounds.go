package api

import (
	"encoding/json"
	"github.com/silkeh/mumble_bot/bot"
	"net/http"
	"os"
	"path/filepath"
)

type Files struct {
	Files []string
}

func (api *API) handleHold(w http.ResponseWriter, req *http.Request) {
	api.handleFileList(w, req, api.client.Config.Mumble.Sounds.Hold, bot.SoundExtension)
}

func (api *API) handleClips(w http.ResponseWriter, req *http.Request) {
	api.handleFileList(w, req, api.client.Config.Mumble.Sounds.Clips, bot.SoundExtension)
}

func (api *API) handleFileList(w http.ResponseWriter, req *http.Request, dir, ext string) {
	if req.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	files := make([]string, 0, 100)
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if filepath.Ext(path) == ext {
				p, _ := filepath.Rel(dir, path)
				files = append(files, p[:len(p)-len(ext)])
			}
			return err
		})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.NewEncoder(w).Encode(&Files{Files: files})
}
