package telegram

import (
	"log"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Target is a telebot.Recipient that can be either a User or Chat.
type Target string

// Recipient returns the ID of the target.
func (t Target) Recipient() string {
	return string(t)
}

// Client is a simplified Telegram client with a single recipient (Target).
type Client struct {
	*tb.Bot
	Target tb.Recipient
}

// NewClient returns a configured Telegram client.
func NewClient(token, target string) (c *Client, err error) {
	c = &Client{Target: Target(target)}
	c.Bot, err = tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	c.Handle(tb.OnSticker, func(m *tb.Message) {
		log.Printf("Received a sticker: %#v", m.Sticker)
	})

	return
}

// SendSticker sends a sticker to the configured recipient.
func (c *Client) SendSticker(sticker *tb.Sticker) (*tb.Message, error) {
	return sticker.Send(c.Bot, c.Target, nil)
}
