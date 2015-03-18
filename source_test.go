package logmunch

import (
	"net/url"
	"testing"
	"time"
)

func TestSourceLoaderGetSource(t *testing.T) {
	s := SourceLoader{
		Config: map[string]url.URL{
			"logentries": url.URL{
				Scheme: "logentries",
				User:   url.UserPassword("", "logentries_password"),
			},
			"file": url.URL{
				Path: "-", // stdin
			},
		},
	}

	_, config, _ := s.GetConfig("logentries:///Path/name/to/fetch")

	if config.User.String() != ":logentries_password" {
		t.Errorf("Expected %v to have User info ':logentries_password', got %s", config, config.User)
	}

	if config.Path != "/Path/name/to/fetch" {
		t.Errorf("Expected %v to have Path '/Path/name/to/fetch', got %s", config, config.Path)
	}
}

func TestSourceLoaderGetData(t *testing.T) {
	s := SourceLoader{
		Config: map[string]url.URL{
			"file": url.URL{
				Scheme: "file",
				Path:   "-", // stdin
			},
		},
	}

	out := make(chan string, 0)
	lines := make([]string, 0)

	go func() {
		for line := range out {
			lines = append(lines, line)
		}
	}()

	_, err := s.GetData("file:/./source_test.go", Query{}, out)

	if err != nil {
		t.Fatalf("s.GetData() error: '%s'.", err)
	}

	if len(lines) == 0 {
		t.Errorf("No lines fetched")
	}
}
