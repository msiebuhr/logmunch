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

func TestCompoundKey(t *testing.T) {
	log := NewLogLine(time.Now(), "what", map[string]string{
		"num": "123",
		"str": "abc",
	})

	nl := MakeCompondKey("com", []string{"num", "missing", "str"})(&log)

	if val, ok := nl.Entries["com"]; !ok || val != "123-∅-abc" {
		t.Errorf("Expected key 'com' to be '123-∅-abc', got %v", val)
	}
}

func TestNormailseUrlPaths(t *testing.T) {
	log := NewLogLine(time.Now(), "what", map[string]string{
		"path": "/test/user_name/action",
	})

	nl := MakeNormaliseUrlPaths("path", []string{"/test/:uid/action", "/test/:uid"})(&log)

	if val, ok := nl.Entries["path"]; !ok || val != "/test/:uid/action" {
		t.Errorf("Expected key 'path' to be '/test/:uid/action', got '%v'", val)
	}

	if val, ok := nl.Entries["uid"]; !ok || val != "user_name" {
		t.Errorf("Expected key 'uid' to be 'user_name', got '%v'", val)
	}
}
