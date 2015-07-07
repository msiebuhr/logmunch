package logmunch

import (
	"io"
	"io/ioutil"
	"testing"
	"time"
)

func TestDrainSqlite3(t *testing.T) {
	l := NewLogLine(
		time.Date(2015, 3, 29, 12, 29, 30, 5000000, time.UTC),
		"some prefix",
		map[string]string{"key.name": "'first'"},
	)

	l2 := NewLogLine(
		time.Date(2015, 3, 29, 12, 29, 30, 5000000, time.UTC),
		"some prefix",
		map[string]string{"key.name": "second"},
	)

	drain := DrainSqlite3()
	reader, writer := io.Pipe()
	in := make(chan LogLine)

	// Start sending some data
	go func() {
		in <- l
		in <- l2
		close(in)
	}()

	// Process it
	go drain(in, writer)

	// Read the output
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		t.Errorf("Didn't expect read to return an error: %s", err)
	}

	expectedData := `CREATE TABLE logs (time, unix, name, key_name);
	
INSERT INTO logs VALUES('2015-03-29T12:29:30.005Z', 1427632170, 'some prefix', '''first''');
INSERT INTO logs VALUES('2015-03-29T12:29:30.005Z', 1427632170, 'some prefix', 'second');
`

	if expectedData != string(data) {
		t.Errorf("Expected\n`%s`\n\tto equal\n`%s`", data, expectedData)
	}
}
