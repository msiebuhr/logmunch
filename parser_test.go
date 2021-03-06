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
		{
			in: `d 323 <158>1 2015-04-07T09:30:14.632370+00:00 d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 heroku router - - at=info method=POST path="/v1/session/yatzy-planet-earth/document" host=api.g2m.me request_id=2466f062-3188-4d78-94f5-07afd45c2381 fwd="202.56.197.65,108.162.225.148" dyno=web.1 connect=1ms service=6304ms status=200 bytes=268`,
			out: LogLine{
				Time: time.Date(2015, 4, 7, 9, 30, 14, 632370000, time.UTC),
				Name: "d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 heroku router",
				Entries: map[string]string{
					"at":              "info",
					"method":          "POST",
					"path":            "/v1/session/yatzy-planet-earth/document",
					"host":            "api.g2m.me",
					"request_id":      "2466f062-3188-4d78-94f5-07afd45c2381",
					"fwd":             "202.56.197.65,108.162.225.148",
					"dyno":            "web.1",
					"connect":         "1ms",
					"service":         "6304ms",
					"status":          "200",
					"bytes":           "268",
					"syslog.severity": "6",
					"syslog.facility": "19",
				},
			},
		},
		{
			in: "183 <45>1 2015-04-07T09:30:14.632370+00:00 d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 heroku api - - Starting process with command `./bin/session_chat_cleaner` by scheduler@addons.heroku.com",
			out: LogLine{
				//Time: time.Date(2015, 7, 28, 19, 40, 34, 170547000, time.UTC),
				Time: time.Date(2015, 4, 7, 9, 30, 14, 632370000, time.UTC),
				Name: "d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 heroku api",
				Entries: map[string]string{
					"syslog.severity": "5",
					"syslog.facility": "5",
					"message":         "Starting process with command `./bin/session_chat_cleaner` by scheduler@addons.heroku.com",
				},
			},
		},
		{
			// No quotes around anything!
			in: `<158>1 2015-03-20T19:22:56.023454+00:00 d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 heroku router - - at=info method=POST dyno=web.2 connect=1ms service=4ms status=200 bytes=455`,
			out: LogLine{
				Time: time.Date(2015, 3, 20, 19, 22, 56, 23454000, time.UTC),
				Name: "d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 heroku router",
				Entries: map[string]string{
					"at":              "info",
					"method":          "POST",
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

		// New-relic output through winston + logentries
		{
			in: `290 <190>1 2015-07-29T07:01:24.756617+00:00 d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 app scheduler.8962 - - {"v":0,"level":30,"name":"newrelic","hostname":"4350e3de-046d-4faf-b381-0361e797301d","pid":3,"time":"2015-07-29T07:01:24.755Z","msg":"Starting New Relic for Node.js connection process."}`,
			out: LogLine{
				Time: time.Date(2015, 7, 29, 7, 1, 24, 756617000, time.UTC),
				Name: "d.f12ee345-3239-4fde-8dc6-b5d1c5656c36 app scheduler.8962",
				Entries: map[string]string{
					"v":               "0",
					"level":           "30",
					"name":            "newrelic",
					"hostname":        "4350e3de-046d-4faf-b381-0361e797301d",
					"pid":             "3",
					"time":            "2015-07-29T07:01:24.755Z",
					"msg":             "Starting New Relic for Node.js connection process.",
					"syslog.severity": "6",
					"syslog.facility": "23",
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
		{
			in: `2015-09-01T16:04:42.747Z info server.recent-sessions.done`,
			out: LogLine{
				Time:    time.Date(2015, 9, 1, 16, 4, 42, 747000000, time.UTC),
				Name:    "info server.recent-sessions.done",
				Entries: map[string]string{},
			},
		},

		// Front-end
		{
			in: `91.199.145.22 INFO browser.name=IE level=info timestamp='2015-06-09T06:30:19.145Z' msg='session.meaningful' participants.length=2`,
			out: LogLine{
				Time: time.Date(2015, 6, 9, 6, 30, 19, 145000000, time.UTC),
				Name: "91.199.145.22 INFO",
				Entries: map[string]string{
					//"tokenData": "{}",
					"msg":                 "session.meaningful",
					"participants.length": "2",
					"browser.name":        "IE",
					"level":               "info",
				},
			},
		},
	}

	for _, tt := range tests {
		in := make(chan string, 1)
		out := make(chan LogLine)

		go ParseLogEntries(in, out)

		in <- tt.in
		close(in)
		log := <-out

		if !log.Equal(tt.out) {
			t.Errorf(
				"Expected line\n\t`%s`\nto parse as\n\t`%s`\nbut got\n\t`%s`",
				tt.in,
				tt.out,
				log,
			)
		}
	}
}

func TestParseInvalidLines(t *testing.T) {
	var tests = []string{
		// Found by go-fuzz
		"",
		"d",
	}

	for _, tt := range tests {
		in := make(chan string)
		out := make(chan LogLine)

		go ParseLogEntries(in, out)

		in <- tt
		close(in)

		for log := range out {
			t.Errorf("Line %s returned unexpected LogLine %v\n", tt, log)
		}
	}

}
