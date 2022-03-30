package audio

// UpSample resamples the given input channel to a rate a factor higher than the original.
// The ratio needs to greater or equal to one.
func UpSample(ratio int16, in <-chan int16, out chan<- int16) {
	defer close(out)

	prev := int16(0)

	for s := range in {
		// Create new samples
		step := (s - prev) / ratio
		for i := int16(1); i < ratio; i++ {
			out <- prev + (i * step)
		}

		// Write sample to out
		out <- s

		// Store previous
		prev = s
	}
}

// DownSample resamples the given input channel to a rate a factor lower than the original.
// The ratio needs to greater or equal to one.
func DownSample(ratio int, in <-chan int16, out chan<- int16) {
	defer close(out)

	num := 0
	sum := 0

	for s := range in {
		num++
		sum += int(s)

		if num >= ratio {
			out <- int16(sum / num)
			num = 0
			sum = 0
		}
	}
}
