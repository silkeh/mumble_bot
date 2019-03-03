package matrix

import (
	matrix "github.com/matrix-org/gomatrix"

	"log"
	"time"
)

// Client is a simplified Matrix client that can send to a single room
type Client struct {
	*matrix.Client
	Syncer *matrix.DefaultSyncer
	roomID string
}

// NewClient returns a configured Matrix Client
func NewClient(homeserverURL, userID, accessToken, roomID string) (c *Client, err error) {
	c = &Client{roomID: roomID}
	c.Client, err = matrix.NewClient(homeserverURL, userID, accessToken)
	if err != nil {
		return
	}
	c.Syncer = c.Client.Syncer.(*matrix.DefaultSyncer)
	_, err = c.JoinRoom(roomID, "", nil)

	return
}

// SendSticker sends a sticker to the configured room
func (c *Client) SendSticker(s *Sticker) (resp *matrix.RespSendEvent, err error) {
	return c.SendMessageEvent(c.roomID, "m.sticker", s)
}

// Sync runs a blocking sync-thread
func (c *Client) Sync() {
	for {
		err := c.Client.Sync()
		if err != nil {
			log.Printf("Sync error: %s", err)
		}
		time.Sleep(1 * time.Second)
	}
}
