package audio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func Read(f io.Reader, out chan<- int16) error {
	defer close(out)

	raw := make([]byte, 2048)
	for {
		n, err := f.Read(raw)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("error reading audio: %w", err)
		}

		if n%2 != 0 {
			n -= 1
		}

		for i := 0; i < n; i += 2 {
			out <- int16(binary.LittleEndian.Uint16(raw[i : i+2]))
		}
	}
}

func Write(f io.Writer, in <-chan int16) error {
	//f, _ = os.OpenFile("/tmp/rec.pcm", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	buf := make([]byte, 320)
	n := 0
	for v := range in {
		binary.LittleEndian.PutUint16(buf[n:n+2], uint16(v))

		n += 2
		if n == len(buf) {
			_, err := f.Write(buf)
			if err != nil {
				return fmt.Errorf("unable to write audio: %w", err)
			}

			n = 0
		}
	}

	return nil
}
