package logmunch

import (
	"testing"
	"time"
)

func TestPickFilter(t *testing.T) {
	log := NewLogLine(
		time.Now(),
		"what",
		map[string]string{
			"keep":    "this",
			"discard": "that",
		},
	)

	filter := MakePickFilter([]string{"keep"})

	filteredLog := filter(&log)

	if _, ok := filteredLog.Entries["discard"]; ok {
		t.Errorf("Didn't expect %v to have key 'disclard'", filteredLog)
	}
}

func TestRoundTimestampFilter(t *testing.T) {
	n := time.Now()
	log := NewLogLine(n, "what", map[string]string{})

	filter := MakeRoundTimestampFilter(time.Hour)

	filteredLog := filter(&log)

	if n.Round(time.Hour) != filteredLog.Time {
		t.Errorf("Expected %v to be %v", filteredLog.Time, n.Round(time.Hour))
	}
}

// TODO: Table-driven test
func TestBucketizeKey(t *testing.T) {
	n := time.Now()
	log := NewLogLine(n, "what", map[string]string{"in": "11"})

	filter := MakeBucketizeKey("in")

	fl := filter(&log)

	if val, ok := fl.Entries["in"]; !ok || val != "10" {
		t.Errorf("Expected %v to be %v", fl.Entries["in"], "10")
	}
}
