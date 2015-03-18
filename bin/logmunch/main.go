package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/msiebuhr/logmunch"
)

var source string
var filter string
var roundTime time.Duration
var start time.Duration
var end time.Duration
var jsonOutput bool
var limit int

func init() {
	flag.StringVar(&source, "source", "Production/api", "Log source")
	flag.StringVar(&filter, "filter", "", "Prefix to fetch")

	flag.BoolVar(&jsonOutput, "json-output", false, "Output as lines of JSON")

	flag.DurationVar(&roundTime, "round-time", time.Duration(0), "Round timestamps to nearest (ex: '1h10m')")
	flag.DurationVar(&start, "start", time.Hour*-24, "When to start fetching data")
	flag.DurationVar(&end, "end", time.Duration(0), "When to stop fetching data")

    flag.IntVar(&limit, "limit", -1, "How many lines to fetch")

	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()

	loader := logmunch.SourceLoader{}
	loader.TryLoadConfigs([]string{"~/.logmunch", "./.logmunch"})

	lines := make(chan string, 100)
	logs := make(chan logmunch.LogLine, 100)

// Get raw log-lines from source
    go func() {
        _, err := loader.GetData(source, logmunch.Query{
            Filter: filter,
            Limit: limit,
            Start: time.Now().Add(start),
            End:   time.Now().Add(end),
        }, lines)

        if err != nil {
            fmt.Printf("ERROR: %s\n", err)
        }
    }()

	// Convert text to logs
	go logmunch.ParseLogEntries(lines, logs)

	// Print the logs
		for line := range logs {
			// Round timestamps (should probably be in a go-routine of it's own
			if roundTime != 0 {
				line.Time = line.Time.Round(roundTime)
			}

			// Print as logfmt'd stuff
			if jsonOutput {
				out, err := json.Marshal(line)
				if err == nil {
					fmt.Println(string(out))
				}
			} else {
				fmt.Println(line.String())
			}
		}
}
