package logmunch

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/kr/logfmt"
)

var timeformats []string = []string{
	// Output from Node.js' Winston logger
	"2006-01-02T15:04:05.000000-07:00",

	// RFC3339 / ISO.. timestamp
	time.RFC3339Nano,
	time.RFC3339,
}

// TODO(msiebuhr): Arrays and nil's
func flattenAndStringifyJSON(prefix string, data map[string]interface{}, log *LogLine) {
	for key, value := range data {

		// Compute new key
		if len(prefix) > 0 {
			key = strings.Join([]string{prefix, key}, ".")
		}

		// What to do with values
		switch t := value.(type) {
		case bool:
			if t {
				log.Entries[key] = "true"
			} else {
				log.Entries[key] = "false"
			}
		case float64:
			log.SetNumber(key, t)
		case string:
			log.Entries[key] = t
		case map[string]interface{}:
			flattenAndStringifyJSON(key, t, log)
		case nil:
			log.Entries[key] = ""
		default:
			//fmt.Println("Doesn't know what to do with", t)
			log.Entries[key] = fmt.Sprintf("UNSUPPORTED: %+v", t)
		}
	}
}

func tryParseOutJSON(line string, log *LogLine) bool {
	curlyIndex := strings.IndexRune(line, '{')
	if curlyIndex == -1 {
		return false
	}

	interfaceMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(line[curlyIndex:]), &interfaceMap)

	if err != nil {
		return false
	}

	// Set line prefix and flatten JSON to one level of strings
	log.Name = strings.Trim(line[:curlyIndex], " \t-")
	flattenAndStringifyJSON("", interfaceMap, log)

	return true
}

// Parse out Heroku's LogFmt dyno output
func tryHerokuLogFmt(line string, log *LogLine) bool {
	// Name is everything until ` - - `; logfmt follows after
	dashIndex := strings.Index(line, " - - ")
	if dashIndex == -1 {
		return false
	}

	// After the dashes, there should at least be three (chosen at random) =-signs
	equals := strings.Count(line[dashIndex+5:], "=")

	if equals < 3 {
		return false
	}

	log.Name = line[:dashIndex]
	logfmt.Unmarshal([]byte(line[dashIndex+5:]), log)
	return true
}

// Try parsing NAME NAME KEY=VALUE
func tryPrefixedLogFmt(line string, log *LogLine) bool {
	// Keep track of the last word-break before a word with '=' in it
	lastSpaceBeforeLogFmtWord := 0
	for i, letter := range line {
		if letter == '=' {
			break
		}
		if unicode.IsSpace(letter) {
			lastSpaceBeforeLogFmtWord = i
		}

		// We got to the end without breaking on a = -> There's no = -> no logfmt in here
		if i == len(line)-1 {
			return false
		}
	}
	log.Name = line[:lastSpaceBeforeLogFmtWord]
	logfmt.Unmarshal([]byte(line[lastSpaceBeforeLogFmtWord:]), log)
	return true
}

func tryPlainMessage(line string, log *LogLine) bool {
	// Name is everything until ` - - `; logfmt follows after
	dashIndex := strings.Index(line, " - - ")
	if dashIndex == -1 {
		return false
	}

	hasEquals := strings.Index(line, "=")

	if hasEquals != -1 {
		return false
	}

	log.Name = line[:dashIndex]
	log.Entries["message"] = line[dashIndex+len(" - - "):]
	return true
}

// Check is the line is tick-escaped logfmt
//
// If we see `='` more often than `="`, assume things are generally `'`-quoted
// and swap all `'` for `"`'s and vice versa.
func tryTicEscapedLogFmt(line string, log *LogLine) bool {
	equalTics := strings.Count(line, "='")
	equalQuote := strings.Count(line, "=\"")

	// Break if we find no equals-things
	if equalTics == 0 {
		return false
	}

	if equalTics > equalQuote {
		// Swap ticks and quotes
		line = strings.Map(func(r rune) rune {
			switch r {
			case '\'':
				return '"'
			case '"':
				return '\''
			default:
				return r
			}
		}, line)
	}

	return tryPrefixedLogFmt(line, log)
}

func ParseLogEntries(in <-chan string, out chan<- LogLine) {
	defer close(out)
	for line := range in {
		// Skip empty lines
		if line == "" {
			continue
		}

		logLine := LogLine{
			Entries: make(map[string]string),
		}

		// Some log-lines from Heroku has a leading `d `, which I can't figure out.
		// So out it goes
		if len(line) >= 2 && line[0] == 'd' && line[1] == ' ' {
			line = line[2:]
		}

		// OFFSET ID TIMESTAMP LINE
		// but also
		// TIMESTAMP LINE
		lineParts := strings.Fields(line)

		if len(lineParts) < 1 {
			continue
		}

		// Remove the first element if equal line length
		// https://tools.ietf.org/html/rfc6587#section-3.4.1
		// LOG := LEN(LINE) + LINE
		length, err := strconv.ParseInt(lineParts[0], 10, 32)
		if err == nil && int(length) == (len(line)-len(lineParts[0])) {
			lineParts = lineParts[1:]
		}

		// Parse out PRIVAL
		// https://tools.ietf.org/html/rfc5424#section-6.2.1
		// < + (facility << 3) + severity + > + SYSLOG_VERSION
		if strings.HasPrefix(lineParts[0], "<") && strings.IndexRune(lineParts[0], '>') != -1 {
			prival, err := strconv.ParseInt(lineParts[0][1:strings.IndexRune(lineParts[0], '>')], 10, 32)
			if err == nil {
				logLine.Entries["syslog.severity"] = fmt.Sprintf("%d", prival&0x7)
				logLine.Entries["syslog.facility"] = fmt.Sprintf("%d", prival>>3)

				lineParts = lineParts[1:]
			}
		}

		// Try parsing each element in the line as various timestamps and see
		// what sticks.
		for i, part := range lineParts {
			// Seen in front-end logging system: `timestamp='TIMESTAMP'` if it starts with that - strip it
			if strings.HasPrefix(part, "timestamp='") {
				part = part[11 : len(part)-1] // Strip `timestamp='` and trailing `'`
			}

			for _, timefmt := range timeformats {
				lineTime, err := time.Parse(timefmt, part)

				if err == nil {
					logLine.Time = lineTime
					newLine := make([]string, len(lineParts)-1)
					copy(newLine[:i], lineParts[:i])
					copy(newLine[i:], lineParts[i+1:])
					lineParts = newLine
					break
				}
			}
		}

		if logLine.Time.IsZero() {
			fmt.Fprintf(os.Stderr, "Could not find timestamp in line `%s`.\n", line)
			continue
		}

		restOfLine := strings.Join(lineParts, " ")

		// The somewhat popular `NAME {… JSON …}`
		if ok := tryParseOutJSON(restOfLine, &logLine); ok {
			out <- logLine
			continue
		}

		// Heroku's `d.UUID NAME - - key=val key=val …` format.
		if ok := tryHerokuLogFmt(restOfLine, &logLine); ok {
			out <- logLine
			continue
		}

		// Logentries serialize with a='b' (not a="b")
		if ok := tryTicEscapedLogFmt(restOfLine, &logLine); ok {
			out <- logLine
			continue
		}

		// Some prefix text and=then some=logfmt
		if ok := tryPrefixedLogFmt(restOfLine, &logLine); ok {
			out <- logLine
			continue
		}

		// Give up. ` SOMETHING - - MESSAGE GOES HERE`
		if ok := tryPlainMessage(restOfLine, &logLine); ok {
			out <- logLine
			continue
		}

		// Really really give up.
		logLine.Name = restOfLine
		out <- logLine
	}
}
