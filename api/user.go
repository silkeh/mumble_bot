package api

import "layeh.com/gumble/gumble"

type PingStats struct {
	UDP PingStat
	TCP PingStat
}

type PingStat struct {
	Packets  uint32
	Average  float32
	Variance float32
}

type UDPStats struct {
	Client gumble.UserStatsUDP
	Server gumble.UserStatsUDP
}

type Stats struct {
	Connected int64
	Ping      PingStats
	UDP       UDPStats
}

type User struct {
	Name            string
	Hash            string
	Stats           *Stats
	Registered      bool
	Muted           bool
	Deafened        bool
	Suppressed      bool
	SelfMuted       bool
	SelfDeafened    bool
	PrioritySpeaker bool
	Recording       bool
}

func NewUser(user *gumble.User) *User {
	u := &User{
		Name:            user.Name,
		Hash:            user.Hash,
		Registered:      user.IsRegistered(),
		Muted:           user.Muted,
		Deafened:        user.Deafened,
		Suppressed:      user.Suppressed,
		SelfMuted:       user.SelfMuted,
		SelfDeafened:    user.SelfDeafened,
		PrioritySpeaker: user.PrioritySpeaker,
		Recording:       user.Recording,
	}
	if user.Stats != nil {
		u.Stats = &Stats{
			Connected: user.Stats.Connected.Unix(),
			Ping: PingStats{
				UDP: PingStat{
					Packets:  user.Stats.UDPPackets,
					Average:  user.Stats.UDPPingAverage,
					Variance: user.Stats.UDPPingVariance,
				},
				TCP: PingStat{
					Packets:  user.Stats.TCPPackets,
					Average:  user.Stats.TCPPingAverage,
					Variance: user.Stats.TCPPingVariance,
				},
			},
			UDP: UDPStats{
				Client: user.Stats.FromClient,
				Server: user.Stats.FromServer,
			},
		}
	}
	return u
}
