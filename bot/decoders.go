package bot

import (
	"encoding/binary"
	"gopkg.in/hraban/opus.v2"
	"os"
)

// decoders contains a mapping of available decoders.
var decoders = map[string]func(*os.File) (AudioStream, error){
	".raw":  rawDecoder,
	".opus": opusDecoder,
}

// rawDecoder is a decoder of .raw files.
func rawDecoder(f *os.File) (AudioStream, error) {
	return &file{File: f}, nil
}

// opusDecoder is a decoder of .opus files.
func opusDecoder(f *os.File) (AudioStream, error) {
	return opus.NewStream(f)
}

// file represents a file containing raw 16-bit PCM audio samples.
type file struct {
	*os.File
}

// Read a sample from the file.
func (f *file) Read(pcm []int16) (int, error) {
	// Read the data from file
	buf := make([]byte, 2*len(pcm))
	n, err := f.File.Read(buf)
	if err != nil {
		return n / 2, err
	}

	// Convert all data
	for i := 0; i < n-1; i += 2 {
		pcm[i/2] = int16(binary.LittleEndian.Uint16(buf[i : i+2]))
	}

	return n / 2, nil
}
