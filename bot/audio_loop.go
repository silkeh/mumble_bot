package bot

import "io"

// audioLoop is a looping version of an AudioStream.
type audioLoop struct {
	AudioStream
	n, count, max int
	buffer        []int16
}

// NewAudioLoop returns an audioLoop for an AudioStream and a number of loops.
// Specifying `count` as smaller than zero will return an audioLoop that will loop
// indefinitely.
func NewAudioLoop(stream AudioStream, count int) AudioStream {
	return &audioLoop{
		AudioStream: stream,
		max:         count,
		buffer:      make([]int16, 0, 256*capacity),
	}
}

// Read a number of 16-bit PCM samples from the stream.
// Returns the number of decoded samples.
func (l *audioLoop) Read(b []int16) (int, error) {
	if l.count == 0 {
		return l.readFile(b)
	}
	if l.max < 0 || l.count < l.max {
		return l.readBuffer(b)
	}
	return 0, io.EOF
}

// readBuffer reads data from the stored buffer.
func (l *audioLoop) readBuffer(b []int16) (int, error) {
	n := len(l.buffer[l.n:])
	s := len(b)
	if n < s {
		s = n
	}

	copy(b, l.buffer[l.n:l.n+s])

	l.n += s
	if len(l.buffer[l.n:]) == 0 {
		l.count++
		l.n = 0
	}

	return s, nil
}

// readFile reads data from the embedded AudioStream.
func (l *audioLoop) readFile(b []int16) (int, error) {
	n, err := l.AudioStream.Read(b)
	if n < 0 {
		panic("negative number of bytes returned")
	}

	l.buffer = append(l.buffer, b[:n]...)
	if err == nil {
		return n, nil
	}

	if err == io.EOF {
		l.count = 1
		return n, nil
	}

	return n, err
}
