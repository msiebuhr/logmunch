package logmunch

import (
	"math"
	"time"
)

// TODO(msiebuhr): RoundValueFilter
// TODO(msiebuhr): QueryFilter
// TODO(msiebuhr): CombineKeysFilter

// Run the LogLines on `in` through the given filters and put them on `out`.
// Note: Closes `out` when `in` does so.
func FilterLogChan(filters []Filterer, in <-chan LogLine, out chan<- LogLine) {
	defer close(out)

	for l := range in {
		lp := &l
		for _, filter := range filters {
			lp = filter(lp)

			if lp == nil {
				goto Next
			}
		}
		out <- *lp

		// Process the next logLine
	Next:
	}

}

// Given a log line, transforms and returns the logline (or perhaps a new
// one/none)
type Filterer func(*LogLine) *LogLine

// Creates a No-Op filter
func NoOpFilter(in *LogLine) *LogLine { return in }

// Creates a filter that keeps only the given keys for each log-line
func MakePickFilter(keys []string) func(*LogLine) *LogLine {
	return func(in *LogLine) *LogLine {
		// Remove all keys except the ones mentioned in `keys`
		for key := range in.Entries {
			found := false
			for _, testKey := range keys {
				if key == testKey {
					found = true
					continue
				}
			}

			if !found {
				delete(in.Entries, key)
			}
		}
		return in
	}
}

// Creates a filter that rounds the LogEntries time to the given duration
func MakeRoundTimestampFilter(d time.Duration) func(*LogLine) *LogLine {
	return func(in *LogLine) *LogLine {
		in.Time = in.Time.Round(d)
		return in
	}
}

// Creates a filter that will bucketize the named key.
//
// Numbers are basically rounded to their nearest value per their order or
// magnitude; i.e. numbers in the hundredes, will be rounded down to nearest
// hundred, numbers in the tousands will be rounded down towards nearest
// thousand.
//
// Note: Round down here actually meands "towards zero"
func MakeBucketizeKey(key string) func(*LogLine) *LogLine {
	return func(in *LogLine) *LogLine {
		if !in.HasKey(key) {
			return in
		}

		v := in.GetNumber(key)
		if v == 0 {
			// We can't just return, as the value could be non-string, and we
			// thus would end up setting some non-numeric keys where only
			// numeric keys are clearly expected.
			in.SetNumber(key, 0)
			return in
		}

		// Figure out how large the buckets should be
		bucketSize := math.Pow(10, math.Floor(math.Log10(math.Abs(v))))

		// Strip all digits after the bucket-size
		in.SetNumber(key, math.Trunc(v/bucketSize)*bucketSize)

		return in
	}
}
