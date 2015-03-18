package logmunch

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseLogEntries(in <-chan string, out chan<- LogLine) {
	defer close(out)
	for line := range in {

		// OFFSET ID TIMESTAMP LINE
		// but also
		// TIMESTAMP LINE
		lineParts := strings.Fields(line)

		if len(lineParts) < 1 {
			continue
		}

		// Try removing the first element if equal line length
		length, err := strconv.ParseInt(lineParts[0], 10, 32)
		if err == nil && int(length) == len(line) {
			lineParts = lineParts[1:]
		}

		// Usually, it's in the beginning
		var lineTime time.Time
		for i, part := range lineParts {
			// 2015-03-06T08:10:13.326080+00:00
			lineTime, err = time.Parse("2006-01-02T15:04:05.000000-07:00", part)

			if err == nil {
				newLine := make([]string, len(lineParts))
				copy(newLine[:i], lineParts[:i])
				copy(newLine[i:], lineParts[i+1:])
				lineParts = newLine
				break
			}

			// 2015-03-16T13:23:07.575Z
			lineTime, err = time.Parse("2006-01-02T15:04:05.000Z", part)

			if err == nil {
				newLine := make([]string, len(lineParts))
				copy(newLine[:i], lineParts[:i])
				copy(newLine[i:], lineParts[i+1:])
				lineParts = newLine
				break
			}
		}

		if lineTime.IsZero() {
			fmt.Printf("Counld not find timestamp in line `%s`.\n", line)
			continue
		}

		l := LogLine{Time: lineTime, RawLine: []byte(strings.Join(lineParts, " "))}
		l.parseLogEntries()
		out <- l
	}
}
