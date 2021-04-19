package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type Users struct {
	Users map[int]*User
}

func (api *API) handleUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	users := make(map[int]*User, len(api.client.Mumble.Users))
	for i, user := range api.client.Mumble.Users {
		user.RequestStats()
		time.Sleep(100 * time.Millisecond)

		users[int(i)] = NewUser(user)
	}

	err := json.NewEncoder(w).Encode(&Users{Users: users})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
