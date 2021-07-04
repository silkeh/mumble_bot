package api

import (
	"github.com/silkeh/mumble_bot/bot"
	"net/http"
)

type API struct {
	client *bot.Client
	mux    *http.ServeMux
}

func NewAPI(c *bot.Client, mux *http.ServeMux) *API {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	api := &API{
		client: c,
		mux:    mux,
	}

	mux.HandleFunc("/metrics", api.handleMetrics)
	mux.HandleFunc("/api/v1/users", api.handleUsers)
	mux.HandleFunc("/api/v1/command", api.handleCommand)
	mux.HandleFunc("/api/v1/clips", api.handleClips)
	mux.HandleFunc("/api/v1/hold", api.handleHold)
	mux.HandleFunc("/api/v1/stickers", api.handleStickers)
	mux.HandleFunc("/api/v1/aliases", api.handleAliases)
	mux.HandleFunc("/api/v1/volume", api.handleVolume)

	return api
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	api.mux.ServeHTTP(w, r)
}

func (api *API) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, api)
}
