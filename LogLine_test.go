package logmunch

import (
	"testing"
	"time"
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
