package api

import (
	"encoding/json"
	"net/http"
)

type Aliases struct {
	Aliases map[string]string
}

func (api *API) handleAliases(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	var aliases Aliases
	if api.client.Config.Mumble != nil {
		aliases.Aliases = api.client.Config.Mumble.Alias
	}

	json.NewEncoder(w).Encode(&aliases)
}
