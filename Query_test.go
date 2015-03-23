package logmunch

import (
	"time"
	"testing"
)

func TestQueryGroup(t *testing.T) {
	l := NewLogLine(time.Now(), "a=b c=d")
    l.parseLogEntries()

	var tests = []struct {
		in  QueryGroup
		out string // The generated value
	}{
		{
			in:  NewQueryGroup("out", []string{"a", "c"}),
			out: "b×d",
		},
		{
			in:  NewQueryGroup("out", []string{"a", "missing"}),
			out: "b×-",
		},
	}

	for _, tt := range tests {
		log := &l

		tt.in.Group(log)

		val, ok := l.Entries[tt.in.Name]

		if !ok {
			t.Errorf("Expected %s to have key %s.", log.Entries, tt.in.Name)
		} else if val != tt.out {
			t.Errorf("Expected value %s to equal %s.", log.Entries[tt.in.Name], tt.out)
		}
	}
}
