package logmunch

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

type Drainer func(<-chan LogLine, io.WriteCloser)

func DrainStandard() func(<-chan LogLine, io.WriteCloser) {
	return func(in <-chan LogLine, out io.WriteCloser) {
		defer out.Close()
		for l := range in {
			out.Write([]byte(l.String()))
			out.Write([]byte{'\n'})
		}
	}
}

func DrainJson() Drainer {
	return func(in <-chan LogLine, out io.WriteCloser) {
		defer out.Close()
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

func DrainCountOverTime(key string) func(<-chan LogLine, io.WriteCloser) {
	return func(in <-chan LogLine, out io.WriteCloser) {
		defer out.Close()
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

		// Loop over timestamps, then keys and print it all
		for _, t := range sortedTimestamps {
			out.Write([]byte(t.Format(time.RFC3339Nano)))
			out.Write([]byte{'\n'})

			for _, k := range sortedKeys {
				v := "0"

				if val, ok := data[t][k]; ok {
					v = fmt.Sprintf("%d", val)
				}

				out.Write([]byte(fmt.Sprintf("\t%s:%s\n", k, v)))
			}
		}
	}
}

func DrainGnuplotDistinctKeyCount(key string) func(<-chan LogLine, io.WriteCloser) {
	return func(in <-chan LogLine, out io.WriteCloser) {
		defer out.Close()
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

func DrainSqlite3() func(<-chan LogLine, io.WriteCloser) {
	return func(in <-chan LogLine, out io.WriteCloser) {
		defer out.Close()
		data := make([]LogLine, 0)
		keyNames := make(map[string]bool)

		for l := range in {
			// Fish out different values of the keys
			for keyName, _ := range l.Entries {
				keyNames[keyName] = true
			}

			data = append(data, l)
		}

		// Sort the keys
		sortedKeys := make([]string, 0, len(keyNames))
		for k := range keyNames {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// Print GNUPLOT stuffs
		out.Write([]byte("CREATE TABLE logs (time, unix, name, "))
		// Join keys and rewrite `.` to `_` (so we don't need escpaing in all cases).
		out.Write([]byte(strings.Replace(strings.Join(sortedKeys, ", "), ".", "_", -1)))
		out.Write([]byte(");\n\n"))

		// Loop over timestamps, then keys and print it all
		for _, k := range data {
			arr := make([]string, len(sortedKeys))
			for i, name := range sortedKeys {
				arr[i] = ""
				if val, ok := k.Entries[name]; ok {
					// In Sqlite (and SQL in general, I believe), `'`s are escaped with `''`.
					arr[i] = strings.Replace(val, "'", "''", -1)
				}
			}

			out.Write([]byte(fmt.Sprintf(
				"INSERT INTO logs VALUES('%s', %d, '%s', '%s');\n",
				k.Time.Format(time.RFC3339Nano),
				k.Time.Unix(),
				k.Name,
				strings.Join(arr, "', '"),
			)))
		}
	}
}
