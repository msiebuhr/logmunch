package logmunch

import (
	"testing"
	"time"
)

func TestLuaFilter(t *testing.T) {
	log := NewLogLine(
		time.Now(), "heroku web.1", map[string]string{
			"load": "100",
		},
	)

	var tests = []struct{
		prog string
		keep bool
		err bool
	}{
		{ prog: "true", keep: true, err: false },
		{ prog: "false", keep: false, err:false  },

		{ prog: "entries.load == 99", keep: false, err: false},

		{ prog: "entries.load > 100", keep: false, err: false},
		{ prog: "entries.load >= 100", keep: true, err: false},
		{ prog: "entries.load == 100", keep: true, err: false},
		{ prog: "entries.load <= 100", keep: true, err: false},
		{ prog: "entries.load < 100", keep: false, err: false},

		{ prog: "entries.load == 101", keep: false, err: false},
	}

	for _, tt := range tests {
		keep, err := filterLineByLua(tt.prog, &log)

		if keep != tt.keep || (err == nil) == tt.err {
			t.Errorf("filterLineByLua(%s, %s) = %t, %s, want %t, %t", tt.prog, log, keep, err, tt.keep, tt.err);
		}
	}
}

func BenchmarkLuaFilter(b *testing.B) {
	log := NewLogLine(
		time.Now(), "heroku web.1", map[string]string{
			"load": "100",
		},
	)
	b.SetBytes(int64(len(log.String())))

    for i := 0; i < b.N; i++ {
        filterLineByLua("load >= 100", &log)
    }
}
