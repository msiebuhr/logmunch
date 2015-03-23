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

// TODO(msiebuhr): Arrays and nil's
func (l *LogLine) convertJsonToKeyValues(prefix []byte, in map[string]interface{}) {
	// l.HandleLogfmt(key, val) on each set of data

	for key, value := range in {
		var k []byte
		if len(prefix) > 0 {
			k = bytes.Join([][]byte{prefix, []byte(key)}, []byte{'.'})
		} else {
			k = []byte(key)
		}

		switch t := value.(type) {
		case bool:
			if t {
				l.HandleLogfmt(k, []byte("true"))
			} else {
				l.HandleLogfmt(k, []byte("false"))
			}
		case float64:
			l.HandleLogfmt(k, []byte(strconv.FormatFloat(t, 'f', -1, 64)))
		case string:
			l.HandleLogfmt(k, []byte(t))
		case map[string]interface{}:
			l.convertJsonToKeyValues(k, t)
		default:
			//fmt.Println("Doesn't know what to do with", t)
			l.HandleLogfmt(k, []byte(fmt.Sprintf("UNSUPPORTED: %+v", t)))
		}
	}
}

func (l *LogLine) parseLogEntries() error {
	// Don't parse if we already have values
	if l.Entries != nil {
		return nil
	}

	// Try parsing substring as JSON
	if bytes.IndexRune(l.RawLine, '{') != -1 {
		var x map[string]interface{}
		err := json.Unmarshal(l.RawLine[bytes.IndexRune(l.RawLine, '{'):], &x)
		// Bingo!
		if err == nil {
			l.Name = strings.Trim(string(l.RawLine[:bytes.IndexRune(l.RawLine, '{')]), " ")
			l.convertJsonToKeyValues([]byte{}, x)
			return nil
		}
	}

	// For name, pick the first thing that isn't a KV-pair
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
