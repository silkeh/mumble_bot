package api

import (
	"encoding/json"
	"layeh.com/gumble/gumble"
	"net/http"
	"strings"
)

type Command struct {
	Command string
}

func (api *API) handleCommand(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		WriteMethodNotAllowed(w)
		return
	}

	var cmd Command
	err := json.NewDecoder(req.Body).Decode(&cmd)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	msg := api.client.HandleCommand(cmd.Command)
	if !strings.HasPrefix(msg, "Unknown command") &&
		!strings.HasPrefix(msg, "Error") {
		api.client.Mumble.Send(&gumble.TextMessage{
			Sender:   api.client.Mumble.Self,
			Channels: []*gumble.Channel{api.client.Mumble.Channels.Find()},
			Message:  msg,
		})
	}

	w.Write([]byte(msg))
}
