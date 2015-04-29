package logmunch

import (
	"fmt"
	"math"
	"regexp"
	"strings"
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

// Make a new key, `newKey`, based on a concatenation of `sourceKeys`.
func MakeCompondKey(newKey string, sourceKeys []string) func(*LogLine) *LogLine {
	return func(in *LogLine) *LogLine {
		newKeyParts := make([]string, len(sourceKeys))

		for i, key := range sourceKeys {
			val, ok := in.Entries[key]
			if !ok {
				val = "∅"
			}
			newKeyParts[i] = val
		}

		in.Entries[newKey] = strings.Join(newKeyParts, "-")

		return in
	}
}

// Normalise paths from a given URL after given express-style templates.
//
// Ex `path=/users/msiebuhr/add` w. template `/users/:uid/add` will normalize
// to `path=/users/:uid/add uid=msiebuhr`.
func MakeNormaliseUrlPaths(key string, urlTemplates []string) func(*LogLine) *LogLine {
	regexps := make([]*regexp.Regexp, len(urlTemplates))
	converterRegex, err := regexp.Compile(":[^/]+")

	if err != nil {
		panic(err)
	}

	// Convert URL templates to Regular expressions
	for i, templ := range urlTemplates {
		// Quite slashes &c.
		withQuotedMeta := regexp.QuoteMeta(templ)

		// Switch out :[^\/]+ for capture groups
		withQuotedMeta = converterRegex.ReplaceAllStringFunc(withQuotedMeta, func(group string) string {
			return fmt.Sprintf("(?P<%s>[^/]+)", group[1:])
		})

		// Add start and end guards
		withQuotedMeta = fmt.Sprintf("^%s$", withQuotedMeta)

		re, err := regexp.Compile(withQuotedMeta)

		if err != nil {
			panic(err)
		}

		regexps[i] = re
	}

	return func(in *LogLine) *LogLine {
		// TODO: Does it have the key we need to check against?

		for i, re := range regexps {
			// Get key and dump query-string
			name := strings.SplitN(in.Entries[key], "?", 2)[0]

			// Check key against regexes
			if !re.MatchString(name) {
				continue
			}

			subMatchNames := re.SubexpNames()
			subMatchValues := re.FindStringSubmatch(name)

			for j := 1; j < len(subMatchNames); j += 1 {
				in.Entries[subMatchNames[j]] = subMatchValues[j]
			}

			in.Entries[key] = urlTemplates[i]
		}

		return in
	}
}

// Make a new key, `newKey`, based on a concatenation of `sourceKeys`.
func MakeRemoveHerokuDrainId() func(*LogLine) *LogLine {
	var indices = []struct {
		at   int
		char uint8
	}{
		{0, 'd'},
		{1, '.'},
		{10, '-'},
		{15, '-'},
		{20, '-'},
		{25, '-'},
		{38, ' '},
	}
	return func(in *LogLine) *LogLine {
		// FIXME: Correct offset?
		if len(in.Name) < 38 {
			return in
		}

		// Check the required chars are there
		for _, t := range indices {
			if in.Name[t.at] != t.char {
				return in
			}
		}

		// TODO: Check everything else is hexadecimal

		// It's a valid prefix! Dump it to a key!
		in.Entries["drainId"] = in.Name[:38]
		in.Name = in.Name[39:]

		return in
	}
}
