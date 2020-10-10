package bot

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// capacity contains the default buffer capacity
const capacity = 2048

// AudioStream represents an audio stream.
type AudioStream interface {
	io.Closer

	// Read a number of 16-bit PCM samples from the stream.
	// Returns the number of decoded samples.
	Read(pcm []int16) (int, error)
}

// OpenSoundFile opens
func OpenSoundFile(path string) (AudioStream, error) {
	decoder, ok := decoders[filepath.Ext(path)]
	if !ok {
		return nil, fmt.Errorf("unsupported format %q", filepath.Ext(path))
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return decoder(f)
}

// ReadAll reads all data from an Audio Stream
func ReadAll(a AudioStream) (b []int16, err error) {
	buf := make([]int16, 0, 256*capacity)
	for {
		b := make([]int16, capacity)
		n, err := a.Read(b)
		if n < 0 {
			panic("negative number of bytes returned")
		}

		buf = append(buf, b[:n]...)
		if err == io.EOF {
			return buf, nil
		}
		if err != nil {
			return nil, err
		}
	}
}
