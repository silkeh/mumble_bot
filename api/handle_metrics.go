package api

import (
	"fmt"
	"net/http"
)

func writeMetric(w http.ResponseWriter, id int, u *User, name string, v interface{}) (int, error) {
	return fmt.Fprintf(w, `mumble_`+name+`{id="%v",name="%s"} %v`+"\n", id, u.Name, v)
}

func (api *API) handleMetrics(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		WriteMethodNotAllowed(w)
		return
	}

	fmt.Fprintf(w, "mumble_connected_users %v", len(api.client.Mumble.Users))
	for i, u := range api.getUsers() {
		writeMetric(w, i, u, "stats_connection_time_seconds", u.Stats.Connected)
		writeMetric(w, i, u, "stats_ping_tcp_count", u.Stats.Ping.TCP.Packets)
		writeMetric(w, i, u, "stats_ping_tcp_avg_ms", u.Stats.Ping.TCP.Average)
		writeMetric(w, i, u, "stats_ping_tcp_var_ms", u.Stats.Ping.TCP.Variance)
		writeMetric(w, i, u, "stats_ping_udp_count", u.Stats.Ping.UDP.Packets)
		writeMetric(w, i, u, "stats_ping_udp_avg_ms", u.Stats.Ping.UDP.Average)
		writeMetric(w, i, u, "stats_ping_udp_var_ms", u.Stats.Ping.UDP.Variance)
		writeMetric(w, i, u, "stats_udp_client_good", u.Stats.UDP.Client.Good)
		writeMetric(w, i, u, "stats_udp_client_late", u.Stats.UDP.Client.Late)
		writeMetric(w, i, u, "stats_udp_client_lost", u.Stats.UDP.Client.Lost)
		writeMetric(w, i, u, "stats_udp_client_resync", u.Stats.UDP.Client.Resync)
		writeMetric(w, i, u, "stats_udp_server_good", u.Stats.UDP.Server.Good)
		writeMetric(w, i, u, "stats_udp_server_late", u.Stats.UDP.Server.Late)
		writeMetric(w, i, u, "stats_udp_server_lost", u.Stats.UDP.Server.Lost)
		writeMetric(w, i, u, "stats_udp_server_resync", u.Stats.UDP.Server.Resync)
	}
}
