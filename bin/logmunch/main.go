package main

import (
	"flag"
	"fmt"
	"os"
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
var outputJson bool
var filterHerokuLogs bool
var outputGnuplotCount string
var outputTableCount string
var limit int
var bucketizeKeys string
var normalisePaths string
var pickKeys string
var compoundKeys string

func init() {
	flag.StringVar(&source, "source", "file:-", "Log source (default: stdin)")
	flag.StringVar(&filter, "filter", "", "Prefix to fetch")

	flag.DurationVar(&start, "start", time.Hour*-24, "When to start fetching data")
	flag.DurationVar(&end, "end", time.Duration(0), "When to stop fetching data")

	flag.IntVar(&limit, "limit", -1, "How many lines to fetch")

	// Output-control
	flag.BoolVar(&outputJson, "output-json", false, "Output as lines of JSON")
	flag.StringVar(&outputGnuplotCount, "output-gnuplot-count", "", "Output as lines of Gnuplot of frequency counts")
	flag.StringVar(&outputTableCount, "output-table-count", "", "Output as table of counts")

	// Filtering
	flag.DurationVar(&roundTime, "round-time", time.Nanosecond, "Round timestamps to nearest (ex: '1h10m')")
	flag.BoolVar(&filterHerokuLogs, "-filter-heroku-logs", true, "Magic parsing of Heroku logs")
	flag.StringVar(&bucketizeKeys, "bucketize", "", "Bucketize this key")
	flag.StringVar(&normalisePaths, "normalise-paths", "", "Normalize URL paths with `:name` placeholders")
	flag.StringVar(&pickKeys, "pick", "", "Keep only these keys")
	flag.StringVar(&compoundKeys, "compound", "", "Combine new,old1,old2,…")

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

	if filterHerokuLogs {
		filters = append(filters, logmunch.MakeRemoveHerokuDrainId())
	}

	if normalisePaths != "" {
		keys := strings.Split(normalisePaths, ",")
		if len(keys) < 2 {
			fmt.Println("Cannot use -normalise-paths withe one argument (ex: `path,/users/:uid`)")
		} else {
			filters = append(
				filters,
				logmunch.MakeNormaliseUrlPaths(keys[0], keys[1:]),
			)
		}
	}

	if bucketizeKeys != "" {
		for _, key := range strings.Split(bucketizeKeys, ",") {
			filters = append(filters, logmunch.MakeBucketizeKey(key))
		}
	}

	if compoundKeys != "" {
		keys := strings.Split(compoundKeys, ",")
		if len(keys) <= 2 {
			fmt.Println("Cannot use -compound with less than two arguments.")
		} else {
			filters = append(
				filters,
				logmunch.MakeCompondKey(keys[0], keys[1:]),
			)
		}
	}

	if pickKeys != "" {
		keys := strings.Split(pickKeys, ",")
		filters = append(filters, logmunch.MakePickFilter(keys))
	}

	if roundTime != 0 {
		filters = append(filters, logmunch.MakeRoundTimestampFilter(roundTime))
	}

	go logmunch.FilterLogChan(filters, logs, filtered)

	if outputJson {
		logmunch.DrainJson()(filtered, os.Stdout)
	} else if outputGnuplotCount != "" {
		logmunch.DrainGnuplotDistinctKeyCount(outputGnuplotCount)(filtered, os.Stdout)
	} else if outputTableCount != "" {
		logmunch.DrainCountOverTime(outputTableCount)(filtered, os.Stdout)
	} else {
		logmunch.DrainStandard()(filtered, os.Stdout)
	}
}
