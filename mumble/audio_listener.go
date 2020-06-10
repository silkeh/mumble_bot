package mumble

import (
	"log"
	"sync"
	"time"

	"layeh.com/gumble/gumble"
)

// AudioListener implements a simple listener that can record audio.
type AudioListener struct {
	sync.Mutex
	buffer []gumble.AudioPacket
}

// OnAudioStream handles AudioStreamEvents.
func (al *AudioListener) OnAudioStream(e *gumble.AudioStreamEvent) {
	go func() {
		for p := range e.C {
			al.Lock()
			if al.buffer != nil {
				log.Printf("Storing %v/%v", len(al.buffer), cap(al.buffer))
				al.buffer = append(al.buffer, *p)
			}
			al.Unlock()
		}
	}()
}

// setBuffer initializes the buffer to a given size in packets.
func (al *AudioListener) setBuffer(size int) {
	al.Lock()
	defer al.Unlock()

	al.buffer = make([]gumble.AudioPacket, 0, size)
}

// getBuffer returns and clears the current audio buffer.
func (al *AudioListener) getBuffer() []gumble.AudioPacket {
	al.Lock()
	defer al.Unlock()

	buf := al.buffer
	al.buffer = nil
	return buf
}

// Record records for a certain time and returns the recorded audio samples.
func (al *AudioListener) Record(duration time.Duration) []gumble.AudioPacket {
	al.setBuffer(int(duration * gumble.AudioSampleRate / time.Second))
	time.Sleep(duration)
	return al.getBuffer()
}
