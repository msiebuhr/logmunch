package logmunch

import (
	"testing"
	"time"

	"encoding/json"
)

func TestLogLinesGetters(t *testing.T) {
	l := NewLogLine(time.Now(), "name", map[string]string{
		"i": "10",
		"f": "1.2",
		"s": "string",
		"m": "12ms",
	})

	// Existance and non-existance
	if !l.HasKey("i") {
		t.Errorf("HasKey(i) on %+v to be true\n", l)
	}
	if l.HasKey("missing") {
		t.Errorf("HasKey(missing) on %+v to be false\n", l)
	}

	// Integer as float
	n := l.GetNumber("i")
	if n != 10 {
		t.Errorf("Expected %+v to have i=10, got %+v\n", l, n)
	}

	// Float as float
	n = l.GetNumber("f")
	if n != 1.2 {
		t.Errorf("Expected %+v to have f=1.2, got %+v\n", l, n)
	}

	// String as float
	n = l.GetNumber("s")
	if n != 0 {
		t.Errorf("Expected %+v to have s=0, got %+v\n", l, n)
	}

	// Mixed as float
	n = l.GetNumber("m")
	if n != 12 {
		t.Errorf("Expected %+v to have m=12, got %+v\n", l, n)
	}

	// Missing as float
	n = l.GetNumber("missing")
	if n != 0 {
		t.Errorf("Expected %+v to have missing=0, got %+v\n", l, n)
	}
}

func TestLogLinesStringer(t *testing.T) {
	l := NewLogLine(
		time.Date(2015, 3, 29, 12, 29, 30, 5000000, time.UTC),
		"some prefix",
		map[string]string{
			"a": "first",
			"z": "last",
		})

	text := "2015-03-29T12:29:30.005Z some prefix a=first z=last"
	out := l.String()

	if out != text {
		t.Errorf("LogLine.String()\nreturned `%s`\nexpected `%s`", text, out)
	}
}

func TestLogLinesJSONEncoder(t *testing.T) {
	l := NewLogLine(
		time.Date(2015, 3, 29, 12, 29, 30, 5000000, time.UTC),
		"some prefix",
		map[string]string{
			"key": "first",
		},
	)

	// Encode to JSON
	j, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("Failed mashalling JSON: %s", err)
	}

	// Decode again
	var out map[string]interface{}
	err = json.Unmarshal(j, &out)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	// Check date
	if date, ok := out["time"]; !ok {
		t.Errorf("JSON %s misses `time` field.", out)
	} else if date.(string) != l.Time.Format(time.RFC3339Nano) {
		t.Errorf("JSON %s .time returned `%s` field.", out, date)
	}

	// Check unix date
	if date, ok := out["unixtime"]; !ok {
		t.Errorf("JSON %s misses .unixtime.", out)
	} else if date.(float64) != 1427632170005 {
		t.Errorf("JSON %s .unixtime returned `%d`.", out, date)
	}

	// Check name
	if name, ok := out["name"]; !ok {
		t.Errorf("JSON %s misses `name` field.", out)
	} else if name.(string) != "some prefix" {
		t.Errorf("JSON %s .name returned `%s` field.", out, name)
	}

	// TODO Check key-values
	if _, ok := out["entries"]; !ok {
		t.Errorf("JSON %s misses .entries.", out)
	}
}

func TestLogLinesEqual(t *testing.T) {
	l := NewLogLine(
		time.Date(2015, 3, 29, 12, 29, 30, 5000000, time.UTC),
		"some prefix",
		map[string]string{"key": "first"},
	)

	l2 := NewLogLine(
		time.Date(2015, 3, 29, 12, 29, 30, 5000000, time.UTC),
		"some prefix",
		map[string]string{"key": "first"},
	)

	if !l.Equal(l2) {
		t.Errorf("Expected\n\t%s\nto equal\n\t%v", l, l2)
	}

	l2.Entries["key2"] = "second"

	if l.Equal(l2) {
		t.Errorf("Expected\n\t%s\nnot to equal\n\t%v", l, l2)
	}
}
