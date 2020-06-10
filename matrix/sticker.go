package matrix

// Sticker represents a sticker message
type Sticker struct {
	URL       string      `json:"url" yaml:"url"`
	Body      string      `json:"body" yaml:"body"`
	RelatesTo error       `json:"m.relates_to" yaml:"relates_to"`
	Info      StickerInfo `json:"info" yaml:"info"`
}

// StickerInfo contains additional sticker information
type StickerInfo struct {
	MimeType string `json:"mimetype" yaml:"mime_type"`
	H        int    `json:"h" yaml:"h"`
	W        int    `json:"w" yaml:"w"`
	Size     int    `json:"size" yaml:"size"`
}
