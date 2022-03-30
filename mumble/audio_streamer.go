package mumble

import (
	"math"
	"sync"

	"layeh.com/gumble/gumble"
)

const jitterBufferSize = 960 // 20ms

type AudioStream struct {
	samples chan int16
}

func (s *AudioStream) Start(e *gumble.AudioStreamEvent) {
	s.samples = make(chan int16, 2048)
	for pkt := range e.C {

		for _, sample := range pkt.AudioBuffer {
			s.samples <- sample
		}
	}

	close(s.samples)
}

func (s *AudioStream) NumSamples() int {
	return len(s.samples)
}

func (s *AudioStream) Sample() (v int16, ok bool) {
	select {
	case v, ok = <-s.samples:
		return v, ok
	default:
		return 0, true
	}
}

type AudioStreamer struct {
	out  chan<- int16
	ins  []*AudioStream
	lock sync.Mutex
}

// OnAudioStream handles AudioStreamEvents.
func (s *AudioStreamer) OnAudioStream(e *gumble.AudioStreamEvent) {
	s.lock.Lock()
	defer s.lock.Unlock()

	a := new(AudioStream)

	s.ins = append(s.ins, a)

	go a.Start(e)
}

func (s *AudioStreamer) Stream(out chan<- int16) {
	for {
		v, ok := s.sample()
		if ok {
			out <- v
		}
	}
}

func (s *AudioStreamer) sample() (out int16, ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Get the number of samples for each stream
	max := 0
	for _, stream := range s.ins {
		if n := stream.NumSamples(); n > max {
			max = n
		}
	}

	// Return if no stream has a buffer of samples
	if max < jitterBufferSize {
		return 0, false
	}

	for i, stream := range s.ins {
		v, chanOk := stream.Sample()
		if !chanOk {
			s.ins = append(s.ins[:i], s.ins[i+1:]...)
			i--
		}

		out = addSample(out, v)
	}

	return out, true
}

func addSample(a, b int16) (c int16) {
	if b > 0 {
		if a > math.MaxInt16-b {
			return math.MaxInt16
		}
	} else {
		if a < math.MinInt16-b {
			return math.MinInt16
		}
	}

	return a + b
}
