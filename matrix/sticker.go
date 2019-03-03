package matrix

// Sticker represents a sticker message
type Sticker struct {
	URL       string      `json:"url"`
	Body      string      `json:"body"`
	RelatesTo error       `json:"m.relates_to"`
	Info      StickerInfo `json:"info"`
}

// StickerInfo contains additional sticker information
type StickerInfo struct {
	MimeType string `json:"mimetype"`
	H        int    `json:"h"`
	W        int    `json:"w"`
	Size     int    `json:"size"`
}

// Stickers contains a preconfigured sticker set
var Stickers = map[string]*Sticker{
	"welcome": {
		URL:       "mxc://integrations.snt.utwente.nl/JpHxNejcIeluxjzgGDNEwfWT",
		Body:      "🍹 (tropical drink)",
		RelatesTo: nil,
		Info: StickerInfo{
			MimeType: "image/webp",
			H:        203,
			W:        256,
			Size:     101312,
		},
	},
}
