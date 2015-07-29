package logmunch

const (
	fuzz_good = 1
	fuzz_meh  = 0
	fuzz_bad  = -1
)

func Fuzz(data []byte) int {
	in := make(chan string)
	out := make(chan LogLine)

	go ParseLogEntries(in, out)

	in <- string(data)
	close(in)

	// Good if we get a line back
	for range out {
		return fuzz_good
	}

	// Less fun otherwise
	return fuzz_meh
}
