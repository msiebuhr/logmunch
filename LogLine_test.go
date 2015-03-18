package logmunch

import (
	"testing"
	"time"
)

func TestLogLineEqualsString(t *testing.T) {
	l := NewLogLine(time.Now(), "a=b c=d")

	if !l.KeyEqualsString("a", "b") {
		t.Errorf("Expected %v to have key a=b", l)
	}

	if l.KeyEqualsString("a", "d") {
		t.Errorf("Expected %v not to have key a=d", l)
	}

	if l.KeyEqualsString("non-existing", "") {
		t.Errorf("Expected %v not to have key non-existing=''", l)
	}
}

func TestLogLinesGetNumber(t *testing.T) {
	l := NewLogLine(time.Now(), "i=10 f=1.2 s=string m=12ms")

	// Integer
	n := l.GetNumber("i")
	if n != 10 {
		t.Errorf("Expected %+v to have i=10, got %+v\n", l, n)
	}

	// Float
	n = l.GetNumber("f")
	if n != 1.2 {
		t.Errorf("Expected %+v to have f=1.2, got %+v\n", l, n)
	}

	// String
	n = l.GetNumber("s")
	if n != 0 {
		t.Errorf("Expected %+v to have s=0, got %+v\n", l, n)
	}

	// Mixed
	n = l.GetNumber("m")
	if n != 12 {
		t.Errorf("Expected %+v to have m=12, got %+v\n", l, n)
	}

	// Missin
	n = l.GetNumber("missing")
	if n != 0 {
		t.Errorf("Expected %+v to have missing=0, got %+v\n", l, n)
	}
}

func TestLogLinesHasPrefix(t *testing.T) {
	l := NewLogLine(time.Now(), "some.service responseTime=12ms")

	if !l.HasPrefix("some") {
		t.Errorf("Expected `%v` to have prefix `some`", l)
	}
}

func TestLogLinesParseJsonObject(t *testing.T) {
	l := NewLogLine(time.Now(), `some.service {"string":"str", "number":123, "bool":true, "nested": {"value": 1}}`)
	l.parseLogEntries()

	// Expected contents
	kvs := map[string]string{
		"string":       "str",
		"number":       "123",
		"bool":         "true",
		"nested.value": "1",
	}

	for k, v := range kvs {
		if !l.KeyEqualsString(k, v) {
			t.Errorf("Expected %v to have key %s=%s", l, k, v)
		}

	}
}
