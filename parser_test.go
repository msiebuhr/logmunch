package logmunch

import (
	"testing"
	"time"
)

func TestParseLines(t *testing.T) {
	testLogLine := NewLogLine(
		time.Date(2015, 3, 29, 12, 29, 30, 5000000, time.UTC),
		"some prefix",
		map[string]string{
			"a": "first",
			"z": "last",
		},
	)

	var tests = []struct {
		in  string
		out LogLine
	}{
		// Programs own stringer
		{
			in:  testLogLine.String(),
			out: testLogLine,
		},

		// Simple hand-written test
		{
			in: `2015-06-12T00:11:22.333Z someName {"num": 123}`,
			out: LogLine{
				Time:    time.Date(2015, 6, 12, 0, 11, 22, 333000000, time.UTC),
				Name:    "someName",
				Entries: map[string]string{"num": "123"},
			},
		},

		// Heroku runtime output
		{
			in: `296 <158>1 2015-03-20T19:22:56.023454+00:00 d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 heroku router - - at=info method=POST path="/v1/oauth/token" host=api.g2m.me request_id=4ce69d2b-fd28-44f0-809c-05e192a0b2e0 fwd="54.160.189.106,173.245.56.103" dyno=web.2 connect=1ms service=4ms status=200 bytes=455`,
			out: LogLine{
				Time: time.Date(2015, 3, 20, 19, 22, 56, 23454000, time.UTC),
				Name: "d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 heroku router",
				Entries: map[string]string{
					"at":              "info",
					"method":          "POST",
					"path":            "/v1/oauth/token",
					"host":            "api.g2m.me",
					"request_id":      "4ce69d2b-fd28-44f0-809c-05e192a0b2e0",
					"fwd":             "54.160.189.106,173.245.56.103",
					"dyno":            "web.2",
					"connect":         "1ms",
					"service":         "4ms",
					"status":          "200",
					"bytes":           "455",
					"syslog.severity": "6",
					"syslog.facility": "19",
				},
			},
		},
		// Node.js' Winston output
		{
			in: `2015-03-20T19:30:35.520Z info middleware.auth.jwt.success {"tokenData":{},"clientId":"g2m-free-web","requestId":"798bbab8-7544-4513-9943-26cfdddec183","dyno":"web.1"}`,
			out: LogLine{
				Time: time.Date(2015, 3, 20, 19, 30, 35, 520000000, time.UTC),
				Name: "info middleware.auth.jwt.success",
				Entries: map[string]string{
					//"tokenData": "{}",
					"clientId":  "g2m-free-web",
					"requestId": "798bbab8-7544-4513-9943-26cfdddec183",
					"dyno":      "web.1",
				},
			},
		},

		// logmunch's own output

	}

	for _, tt := range tests {
		in := make(chan string, 1)
		out := make(chan LogLine)

		go ParseLogEntries(in, out)

		in <- tt.in
		close(in)
		log := <-out

		// Check name
		if log.Name != tt.out.Name {
			t.Errorf(
				"Expected line\n\t%s\nto have name `%s` but got `%s`",
				//"Line `%s` parsed out name `%s`, expected `%s`.",
				tt.in,
				tt.out.Name,
				log.Name,
			)
		}

		// Timestamp
		if !log.Time.Equal(tt.out.Time) {
			t.Errorf(
				"Expected line\n\t%s\nto have time `%s` but got `%s`",
				//"Line `%s` parsed out time `%s`, expected `%s`.",
				tt.in,
				tt.out.Time,
				log.Time,
			)
		}

		// Check parsed out items are all present in the expected output
		for key, expectedValue := range tt.out.Entries {
			value, hasValue := log.Entries[key]
			if !hasValue {
				t.Errorf("Line `%s` misses key %s (=%v)", tt.in, key, expectedValue)
			} else if expectedValue != value {
				t.Errorf("Line `%s` should have value %s for key %s, got %s", tt.in, expectedValue, key, value)
			}
		}

		// Check there is no extraneous keys parsed out
		for key := range log.Entries {
			if _, ok := tt.out.Entries[key]; !ok {
				t.Errorf("Line `%s` has unexpected key %s", tt.in, key)
			}
		}
	}
}
