package api

import (
	"encoding/json"
	"net/http"
)

type Stickers struct {
	Stickers []string
}

func (api *API) handleStickers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	stickers := &Stickers{Stickers: []string{}}
	if api.client.Telegram != nil && api.client.Config.Telegram.Stickers != nil {
		stickers.Stickers = make([]string, 0, len(api.client.Config.Telegram.Stickers))
		for n := range api.client.Config.Telegram.Stickers {
			stickers.Stickers = append(stickers.Stickers, n)
		}
	}

	json.NewEncoder(w).Encode(&stickers)
}
