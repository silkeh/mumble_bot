package api

import (
	"encoding/json"
	"github.com/silkeh/mumble_bot/bot"
	"net/http"
)

type Volume struct {
	Current, Min, Max int
}

func (api *API) handleVolume(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	volume := &Volume{
		Current: int(api.client.Volume()),
		Min:     bot.MinVolume,
		Max:     bot.MaxVolume,
	}

	err := json.NewEncoder(w).Encode(&volume)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
