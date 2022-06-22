package sudoku

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	s := New()
	if len(s.unitlist) != 27 {
		t.Errorf("got len=%v, want 27", len(s.unitlist))
	}

	fmt.Println(s.units[20])
	fmt.Println(s.peers[20])
}
