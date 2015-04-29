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

type filterTest struct {
	in  *LogLine
	out *LogLine
}

type filterTests []filterTest

// Run all given tests though the filter
func (f filterTests) run(filter Filterer, t *testing.T) {
	for _, tt := range f {
		res := filter(tt.in)

		// Both nil? Then OK
		if tt.out == nil && res == nil {
			return
		}

		if !tt.out.Equal(*res) {
			t.Errorf("Expected log\n\t%s\nto equal\n\t%s", res, tt.out)
		}
	}
}

func (f filterTests) normalizeTimestamps() {
	t := time.Now()

	for _, tt := range f {
		if tt.in != nil {
			tt.in.Time = t
		}
		if tt.out != nil {
			tt.out.Time = t
		}
	}
}

func TestRemoveHerokuDrainId(t *testing.T) {
	tests := filterTests{
		filterTest{
			in: &LogLine{
				Time:    time.Now(),
				Name:    "d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 what",
				Entries: map[string]string{},
			},
			out: &LogLine{
				Time:    time.Now(),
				Name:    "what",
				Entries: map[string]string{"drainId": "d.f12ee345-3239-4fde-8dc6-b5d1c5656c36"},
			},
		},
		filterTest{
			in: &LogLine{
				Time:    time.Now(),
				Name:    "no id",
				Entries: map[string]string{},
			},
			out: &LogLine{
				Time:    time.Now(),
				Name:    "no id",
				Entries: map[string]string{},
			},
		},
		filterTest{in: nil, out: nil},
	}

	tests.normalizeTimestamps()
	tests.run(MakeRemoveHerokuDrainId(), t)
}

func TestNormailseUrlPaths(t *testing.T) {
	tests := filterTests{
		filterTest{
			in:  &LogLine{time.Now(), "a", map[string]string{"path": "/users/NAME/avatar"}},
			out: &LogLine{time.Now(), "a", map[string]string{"path": "/users/:uid/avatar", "uid": "NAME"}},
		},
		filterTest{
			in:  &LogLine{time.Now(), "a", map[string]string{"path": "/users/NAME"}},
			out: &LogLine{time.Now(), "a", map[string]string{"path": "/users/:uid", "uid": "NAME"}},
		},
		filterTest{
			in:  &LogLine{time.Now(), "a", map[string]string{"path": "/no/match"}},
			out: &LogLine{time.Now(), "a", map[string]string{"path": "/no/match"}},
		},
		filterTest{
			in:  &LogLine{time.Now(), "a", map[string]string{"other": "key"}},
			out: &LogLine{time.Now(), "a", map[string]string{"other": "key"}},
		},
		filterTest{in: nil, out: nil},
	}

	tests.normalizeTimestamps()
	tests.run(MakeNormaliseUrlPaths("path", []string{"/users/:uid/avatar", "/users/:uid"}), t)
}

func TestBucketizeKey(t *testing.T) {
	tests := filterTests{
		filterTest{
			in:  &LogLine{time.Now(), "a", map[string]string{"v": "1.1"}},
			out: &LogLine{time.Now(), "a", map[string]string{"v": "1"}},
		},
		filterTest{
			in:  &LogLine{time.Now(), "a", map[string]string{"v": "411.6"}},
			out: &LogLine{time.Now(), "a", map[string]string{"v": "400"}},
		},
		filterTest{in: nil, out: nil},
	}

	tests.normalizeTimestamps()
	tests.run(MakeBucketizeKey("v"), t)
}
