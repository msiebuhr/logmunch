package logmunch

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"encoding/json"
)

type LogLine struct {
	Time    time.Time
	Name    string
	Entries map[string]string
}

func NewLogLine(when time.Time, name string, entries map[string]string) LogLine {
	return LogLine{
		Time:    when,
		Name:    name,
		Entries: entries,
	}
}

func (l *LogLine) String() string {
	parts := make([]string, 2, len(l.Entries)+2)

	parts[0] = l.Time.Format(time.RFC3339Nano)
	parts[1] = l.Name

	// Key-value pairs
	for key, value := range l.Entries {
		value = strings.ReplaceAll(value, "\n", "\\n")
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

// Implements interface from "github.com/kr/logfmt"
func (l *LogLine) HandleLogfmt(key, val []byte) error {
	l.Entries[string(key)] = string(val)
	return nil
}

func (l *LogLine) HasKey(key string) bool {
	_, exists := l.Entries[key]
	return exists
}

func (l *LogLine) GetNumber(key string) float64 {

	value, exists := l.Entries[key]
	if !exists {
		return 0
	}

	numbersInValue := strings.Map(func(r rune) rune {
		if (r >= '0' && '9' >= r) || r == '.' || r == '-' {
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

func (l *LogLine) SetNumber(key string, val float64) {
	l.Entries[key] = strconv.FormatFloat(val, 'f', -1, 64)
}

func (l *LogLine) HasPrefix(prefix string) bool {
	return strings.HasPrefix(l.Name, prefix)
}

func (l LogLine) Equal(other LogLine) bool {
	// Compare name
	if l.Name != other.Name {
		return false
	}

	// Compare time
	if !l.Time.Equal(other.Time) {
		return false
	}

	// Check all Entries in l match ones in other
	for key, expectedValue := range l.Entries {
		actualValue, hasKey := other.Entries[key]
		if !hasKey || expectedValue != actualValue {
			return false
		}
	}

	// Check other don't have too many entries
	for key := range other.Entries {
		_, hasKey := l.Entries[key]
		if !hasKey {
			return false
		}
	}

	return true
}
