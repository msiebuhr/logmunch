package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/msiebuhr/logmunch"
)

var source string
var filter string
var roundTime time.Duration
var start time.Duration
var end time.Duration
var jsonOutput bool
var limit int
var bucketizeKeys string
var pickKeys string

func init() {
	flag.StringVar(&source, "source", "Production/api", "Log source")
	flag.StringVar(&filter, "filter", "", "Prefix to fetch")

	flag.DurationVar(&start, "start", time.Hour*-24, "When to start fetching data")
	flag.DurationVar(&end, "end", time.Duration(0), "When to stop fetching data")

	flag.IntVar(&limit, "limit", -1, "How many lines to fetch")

	// Output-control
	flag.BoolVar(&jsonOutput, "json-output", false, "Output as lines of JSON")

	// Filtering
	flag.DurationVar(&roundTime, "round-time", time.Nanosecond, "Round timestamps to nearest (ex: '1h10m')")
	flag.StringVar(&bucketizeKeys, "bucketize", "", "Bucketize this key")
	flag.StringVar(&pickKeys, "pick", "", "Keep only these keys")

	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()

	loader := logmunch.SourceLoader{}
	fileLocations := []string{"./.logmunch"}
	dir, err := homedir.Expand("~/.logmunch")
	if err == nil {
		fileLocations = append(fileLocations, dir)
	}
	loader.TryLoadConfigs(fileLocations)

	lines := make(chan string, 100)
	logs := make(chan logmunch.LogLine, 100)
	filtered := make(chan logmunch.LogLine, 100)

	// Get raw log-lines from source
	go func() {
		_, err := loader.GetData(source, logmunch.Query{
			Filter: filter,
			Limit:  limit,
			Start:  time.Now().Add(start),
			End:    time.Now().Add(end),
		}, lines)

		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
		}
	}()

	// Convert text to logs
	go logmunch.ParseLogEntries(lines, logs)

	// Filter the loglines
	filters := []logmunch.Filterer{}

	if pickKeys != "" {
		keys := strings.Split(pickKeys, ",")
		filters = append(filters, logmunch.MakePickFilter(keys))
	}

	if roundTime != 0 {
		filters = append(filters, logmunch.MakeRoundTimestampFilter(roundTime))
	}
	if bucketizeKeys != "" {
		for _, key := range strings.Split(bucketizeKeys, ",") {
			filters = append(filters, logmunch.MakeBucketizeKey(key))
		}
	}

	go logmunch.FilterLogChan(filters, logs, filtered)

	// Print the logs
	for line := range filtered {
		// Round timestamps (should probably be in a go-routine of it's own
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
