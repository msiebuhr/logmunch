package logmunch

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
)

type Drainer func(<-chan LogLine, io.Writer)

func DrainStandard() func(<-chan LogLine, io.Writer) {
	return func(in <-chan LogLine, out io.Writer) {
		for l := range in {
			out.Write([]byte(l.String()))
			out.Write([]byte{'\n'})
		}
	}
}

func DrainJson() Drainer {
	return func(in <-chan LogLine, out io.Writer) {
		for l := range in {
			j, err := json.Marshal(l)
			if err == nil {
				out.Write(j)
				out.Write([]byte{'\n'})
			}
		}
	}
}

// Helper to sort list of timestamps
type timeList []time.Time

func (t timeList) Len() int           { return len(t) }
func (t timeList) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t timeList) Less(i, j int) bool { return t[i].Before(t[j]) }

func DrainGnuplotDistinctKeyCount(key string) func(<-chan LogLine, io.Writer) {
	return func(in <-chan LogLine, out io.Writer) {
		keyValues := make(map[string]bool)
		data := make(map[time.Time]map[string]int)

		curTime := time.Unix(0, 0)

		for l := range in {
			// Fish out different values of the keys
			keyValues[l.Entries[key]] = true

			// Keep track of current time
			if !curTime.Equal(l.Time) {
				curTime = l.Time
				// FIXME: Check if key already exists
				data[curTime] = make(map[string]int)
			}

			// Make sure the counter has been initalized
			if _, ok := data[curTime][l.Entries[key]]; !ok {
				data[curTime][l.Entries[key]] = 0
			}

			data[curTime][l.Entries[key]] += 1
		}

		// Sort the keys
		sortedKeys := make([]string, 0, len(keyValues))
		for k := range keyValues {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// Sort the timestamps
		var sortedTimestamps timeList = make([]time.Time, 0, len(data))
		for t := range data {
			sortedTimestamps = append(sortedTimestamps, t)
		}
		sort.Sort(sortedTimestamps)

		// Print GNUPLOT stuffs
		out.Write([]byte(
			`set terminal png
set output 'x.png'

set timefmt "%s"
set xdata time

plot`))

		for i, name := range sortedKeys {
			out.Write([]byte(fmt.Sprintf(" '-' using 1:2 with linespoints title '%s'", name)))

			if i != len(sortedKeys)-1 {
				out.Write([]byte(", \\\n"))
			}
		}
		out.Write([]byte("\n"))

		// Prit header
		//out.Write([]byte(fmt.Sprintf("TIME\t%s\n", strings.Join(sortedKeys, "\t"))))

		// Loop over timestamps, then keys and print it all
		/*
		   stuffs := make([]string, len(sortedKeys) + 1);
		   for _, t:= range sortedTimestamps {
		       stuffs[0] = fmt.Sprintf("%d", t.Unix())

		       for i, k := range sortedKeys {
		           stuffs[i + 1] = "0"

		           if val, ok := data[t][k]; ok {
		               stuffs[i + 1] = fmt.Sprintf("%d", val)
		           }
		       }
		       out.Write([]byte(strings.Join(stuffs, "\t")))
		       out.Write([]byte{'\n'})
		   }
		*/

		// Loop over timestamps, then keys and print it all
		for _, k := range sortedKeys {
			for _, t := range sortedTimestamps {
				v := 0
				if val, ok := data[t][k]; ok {
					v = val
				}
				out.Write([]byte(fmt.Sprintf(" %d\t%d\n", t.Unix(), v)))
			}
			out.Write([]byte("EOF\n"))
		}
	}
}
