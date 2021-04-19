package api

import "layeh.com/gumble/gumble"

type Stat struct {
	Packets      uint32
	PingAverage  float32
	PingVariance float32
}

type Stats struct {
	UDP Stat
	TCP Stat
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
			UDP: Stat{
				Packets:      user.Stats.UDPPackets,
				PingAverage:  user.Stats.UDPPingAverage,
				PingVariance: user.Stats.UDPPingVariance,
			},
			TCP: Stat{
				Packets:      user.Stats.TCPPackets,
				PingAverage:  user.Stats.TCPPingAverage,
				PingVariance: user.Stats.TCPPingVariance,
			},
		}
	}
	return u
}
