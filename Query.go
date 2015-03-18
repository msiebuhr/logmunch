package logmunch

import (
	"time"
)

type Query struct {
	// A prefix to group things by
	Filter string

	// When the log start and end
	Start time.Time
	End   time.Time

	// How many lines to fetch
	Limit int

	// TODO: Filters?
	// TODO: Keep/discard some key/value pairs?
	// TODO: Some display niceness?
}
