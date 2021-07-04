package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type Users struct {
	Users map[int]*User
}

func (api *API) getUsers() map[int]*User {
	users := make(map[int]*User, len(api.client.Mumble.Users))

	for _, user := range api.client.Mumble.Users {
		user.RequestStats()
	}

	time.Sleep(100 * time.Millisecond)

	for i, user := range api.client.Mumble.Users {
		users[int(i)] = NewUser(user)
	}

	return users
}

func (api *API) handleUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	users := api.getUsers()
	err := json.NewEncoder(w).Encode(&Users{Users: users})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
