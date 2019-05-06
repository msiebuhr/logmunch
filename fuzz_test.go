package logmunch

import (
	"testing"
)

func TestFuzz(t *testing.T) {
	interesting := [][]byte{
		[]byte(`2015-06-12T00:11:22.333Z someName {"num": 123}`),
	}

	for _, tt := range interesting {
		st := Fuzz(tt)

		if st != 1 {
			t.Errorf("Expected %s to be interesting (=1), got %d", tt, st)
		}
	}

	others := [][]byte{
		//[]byte{},
	}

	for _, tt := range others {
		st := Fuzz(tt)

		if st == 1 {
			t.Errorf("Expected %s to be interesting (!=1), got %d", tt, st)
		}
	}

}
