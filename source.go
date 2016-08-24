package logmunch

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func outputLinesAndCloseChan(in io.ReadCloser, out chan<- string) error {
	defer in.Close()

	// Work on them lines!
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		b := scanner.Bytes()
		if len(b) != 0 {
			//sendBytes = make([]byte, len(b))
			//copy(sendBytes, b)
			// TODO: Does this do a copy?
			out <- string(b)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

type Source func(config *url.URL, query Query, out chan<- string) (Query, error)

// Get data from a file
func FileSource(config *url.URL, query Query, out chan<- string) (Query, error) {
	defer close(out)
	var file *os.File
	name := config.Path

	// We can't give relative urls `file:./relative.file`, but
	// `file:///./file.txt` and `file:/./file.txt` both turn into `/./file.txt`
	// - so we consider that a relative one
	if len(name) > 2 && name[0:2] == "/." {
		name = name[1:]
	}

	if name == "-" || name == "" {
		file = os.Stdin
	} else {
		var err error
		file, err = os.Open(name)

		if err != nil {
			return query, err
		}
	}

	err := outputLinesAndCloseChan(file, out)
	return query, err
}

// Get data from logentries
func LogEntriesSource(config *url.URL, query Query, out chan<- string) (Query, error) {
	defer close(out)

	if config.User == nil {
		return query, errors.New("No LogEntries password set!")
	}

	password, gotPassword := config.User.Password()

	if !gotPassword {
		return query, errors.New("No LogEntries password set!")
	}

	logentriesurl := fmt.Sprintf(
		"https://pull.logentries.com/%s/hosts/%s/?start=%d&end=%d",
		password,
		config.Path,
		query.Start.Unix()*1000,
		query.End.Unix()*1000,
	)

	// Add filter if one is given
	if query.Filter != "" {
		logentriesurl = logentriesurl + "&filter=" + url.QueryEscape(query.Filter)

		// Re-set filter, as we process it here
		query.Filter = ""
	}

	if query.Limit > 0 {
		logentriesurl = fmt.Sprintf("%s&limit=%d", logentriesurl, query.Limit)

		// Re-set limit, as we process it here
		query.Limit = -1
	}

	resp, err := http.Get(logentriesurl)

	if err != nil {
		close(out)
		return query, err
	} else if (resp.StatusCode != 200) {
		return query, fmt.Errorf("Logentries returned HTTP %d for %s", resp.StatusCode , logentriesurl)
	}

	err = outputLinesAndCloseChan(resp.Body, out)
	return query, err
}

// Keep a map of protocols -> default settings, all parsed from URLs
type SourceLoader struct {
	Config map[string]url.URL
}

func (s *SourceLoader) TryLoadConfigs(filenames []string) error {
	// Create config map if not set
	if s.Config == nil {
		s.Config = make(map[string]url.URL)
	}

	// Load files
	for _, filename := range filenames {
		// Load it
		file, err := os.Open(filename)
		if err != nil {
			continue // File doesn't exist
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			u, err := url.Parse(scanner.Text())
			if err == nil {
				s.Config[u.Scheme] = *u
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s SourceLoader) GetConfig(configUrl string) (Source, *url.URL, error) {
	u, err := url.Parse(configUrl)

	if err != nil {
		return nil, nil, err
	}

	// Test if u.Scheme is in s.config
	defaultConfig, found := s.Config[u.Scheme]

	// Got a default confg, then mix defaults + given config
	if found {
		// Copy default into new one
		newConf := &url.URL{
			Scheme: defaultConfig.Scheme,
			//Opaque:   defaultConfig.Opaque,
			User:     defaultConfig.User,
			Host:     defaultConfig.Host,
			Path:     defaultConfig.Path,
			RawQuery: defaultConfig.RawQuery,
			Fragment: defaultConfig.Fragment,
		}

		if u.User != nil {
			newConf.User = u.User
		}
		if u.Host != "" {
			newConf.Host = u.Host
		}
		if u.Path != "" {
			newConf.Path = u.Path
		} else if u.Opaque != "" {
			// HACK: Input is sometimes written as `proto:PA/TH`, which
			// makes the PA/TH end up in the opaque secion. We re-use that
			// as the path.
			newConf.Path = u.Opaque
		}
		if u.RawQuery != "" {
			newConf.RawQuery = u.RawQuery
		}
		if u.Fragment != "" {
			newConf.Fragment = u.Fragment
		}

		u = newConf
	}

	// Return Source
	switch u.Scheme {
	case "logentries":
		return LogEntriesSource, u, nil
		//case "stdin":
	case "file":
		return FileSource, u, nil
		//return SourceFile{}, nil
	}

	// No known source bu-hu.
	return nil, nil, errors.New(fmt.Sprintf("Unknown source '%s'.", u.Scheme))
}

func (s SourceLoader) GetData(configUrl string, query Query, out chan<- string) (Query, error) {
	sourceFunc, config, err := s.GetConfig(configUrl)
	if err != nil {
		close(out)
		return query, err
	}

	return sourceFunc(config, query, out)
}
