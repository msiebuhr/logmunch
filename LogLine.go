package logmunch

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"github.com/kr/logfmt"
)

type LogLine struct {
	Time    time.Time
	RawLine []byte `json:"-"`
	Name    string
	Entries map[string]string
}

func NewLogLine(when time.Time, line string) LogLine {
	return LogLine{
		Time:    when,
		RawLine: []byte(line),
	}
}

func (l *LogLine) String() string {
	parts := make([]string, 2, len(l.Entries)+2)

	parts[0] = l.Time.Format(time.RFC3339Nano)
	parts[1] = l.Name

	// Key-value pairs
	for key, value := range l.Entries {
		if strings.Contains(value, " ") {
			value = fmt.Sprintf(`"%s"`, value)
		}
		if strings.Contains(key, " ") {
			key = fmt.Sprintf(`"%s"`, key)
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(parts[2:])

	return strings.Join(parts, " ")
}

func (l LogLine) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Time    time.Time         `json:"time"`
		Unix    int64             `json:"unixtime"`
		Name    string            `json:"name"`
		Entries map[string]string `json:"entries"`
	}{
		Time:    l.Time,
		Unix:    l.Time.UnixNano() / 1e6,
		Name:    l.Name,
		Entries: l.Entries,
	})
}

func (l *LogLine) HandleLogfmt(key, val []byte) error {
	if l.Entries == nil {
		l.Entries = make(map[string]string)
	}
	l.Entries[string(key)] = string(val)
	return nil
}

func (l *LogLine) parseLogEntries() error {
	// Don't parse if we already have values
	//if l.Entries != nil { return nil }

	// Give up: Just take the first thing that doesn't have a = in it.
	names := bytes.Fields(l.RawLine)
	for _, name := range names {
		if bytes.Index(name, []byte{'='}) == -1 {
			l.Name = strings.Trim(string(name), " ")
			break
		}
	}

	// Parse values
	return logfmt.Unmarshal(l.RawLine, l)
}

func (l *LogLine) HasKey(key string) bool {
	l.parseLogEntries()
	_, exists := l.Entries[key]
	return exists
}

func (l *LogLine) KeyEqualsString(key, value string) bool {
	l.parseLogEntries()
	val, exists := l.Entries[key]
	if !exists {
		return false
	}
	return val == value
}

func (l *LogLine) GetNumber(key string) float64 {
	l.parseLogEntries()

	value, exists := l.Entries[key]
	if !exists {
		return 0
	}

	numbersInValue := strings.Map(func(r rune) rune {
		if (r >= '0' && '9' >= r) || r == '.' {
			return r
		}
		return -1
	}, value)

	n, err := strconv.ParseFloat(numbersInValue, 64)
	if err != nil {
		return 0
	}

	return n
}

func (l *LogLine) HasPrefix(prefix string) bool {
	l.parseLogEntries()
	return strings.HasPrefix(l.Name, prefix)
}
