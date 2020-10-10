package bot

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// capacity contains the default buffer capacity
const capacity = 512

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
	buf := make([]int16, capacity, 16*capacity)
	n := 0
	for {
		m, err := a.Read(buf[n:])
		if m < 0 {
			panic("negative number of bytes returned")
		}

		n += m
		if err == io.EOF {
			return buf[:n], nil
		}
		if err != nil {
			return nil, err
		}

		if m <= cap(buf) {
			buf = buf[:m]
		} else {
			old := buf
			buf = make([]int16, cap(buf)+capacity, 16*capacity)
			copy(buf[:len(old)], old)
		}
	}
}
