package logmunch

import (
	"time"
	"strings"
)

// QueryGroup
type QueryGroup struct {
    Name string
    Keys []string
}

func NewQueryGroup(name string, keys []string) QueryGroup {
    return QueryGroup{
        Name: name,
        Keys: keys,
    }
}

// Create a new key in l by name and composed of the given keys.
func (q *QueryGroup) Group(l *LogLine) {
    values := make([]string, len(q.Keys))

    for i, key := range q.Keys {
        if value, ok := l.Entries[key]; ok {
            values[i] = value
        } else {
            values[i] = "-"
        }
    }

    l.Entries[q.Name] = strings.Join(values, "×")
}

type Query struct {
	// A prefix to group things by
	Filter string

	// When the log start and end
	Start time.Time
	End   time.Time

	// How many lines to fetch
	Limit int

	// TODO: Filters?
	// TODO: Keep/discard some key/value pairs?
	// TODO: Some display niceness?

    GroupBy []QueryGroup
}
