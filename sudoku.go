package sudoku

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

// Index represents a square on the Sudoku board; it's a number in the inclusive
// range [0, 80] that stands for row*9+col.
//
// These are the squares designated by an Index:
//
//  0  1  2 |  3  4  5 |  6  7  8
//  9 10 11 | 12 13 14 | 15 16 17
// 18 19 20 | 21 22 23 | 24 25 26
// ---------+----------+---------
// 27 28 29 | 30 31 32 | 33 34 35
// 36 37 38 | 39 40 41 | 42 43 44
// 45 46 47 | 48 49 50 | 51 52 53
// ---------+----------+---------
// 54 55 56 | 57 58 59 | 60 61 62
// 63 64 65 | 66 67 68 | 69 70 71
// 72 73 74 | 75 76 77 | 78 79 80
type Index = int

// Unit is a list of square indices that belong to the same Sudoku
// unit - a row, column or 3x3 block which should contain unique digits.
type Unit = []Index

// index calculates the linear index of a square from its row and column.
func index(row, col int) Index {
	return row*9 + col
}

// Values represents a Sudoku board in a format that's usable for solving.
// An element at index [i] in Values represents Sudoku square i (see the
// documentation of the Index type), and contains a set of all candidate
// digits for this square.
type Values []Digits

type Sudoku struct {
	// unitlist is the list of all units that exist on the board.
	unitlist []Unit

	// units maps an index to a list of units that contain that square.
	// The mapping is a slice, i.e. units[i] is a list of all the units
	// that contain the square with index i.
	units [][]Unit

	// peers maps an index to a list of unique peers - other indices that share
	// some unit with this index (it won't contain the index itself).
	peers [][]Index
}

func New() *Sudoku {
	var unitlist []Unit

	// row units
	for row := 0; row < 9; row++ {
		var rowUnit []Index
		for col := 0; col < 9; col++ {
			rowUnit = append(rowUnit, index(row, col))
		}
		unitlist = append(unitlist, rowUnit)
	}

	// column units
	for col := 0; col < 9; col++ {
		var colUnit []Index
		for row := 0; row < 9; row++ {
			colUnit = append(colUnit, index(row, col))
		}
		unitlist = append(unitlist, colUnit)
	}

	// 3x3 block units
	for blockRow := 0; blockRow < 3; blockRow++ {
		for blockCol := 0; blockCol < 3; blockCol++ {
			var blockUnit []Index

			for row := 0; row < 3; row++ {
				for col := 0; col < 3; col++ {
					blockUnit = append(blockUnit, index(blockRow*3+row, blockCol*3+col))
				}
			}
			unitlist = append(unitlist, blockUnit)
		}
	}

	// For each index i, units[i] is a list of all units that contain i.
	units := make([][]Unit, 81)
	for i := 0; i < 81; i++ {
		for _, unit := range unitlist {
			if slices.Index(unit, i) >= 0 {
				units[i] = append(units[i], slices.Clone(unit))
			}
		}
	}

	// For each index i, peers[i] is a list of unique indices that share some
	// unit with i.
	peers := make([][]Index, 81)
	for i := 0; i < 81; i++ {
		for _, unit := range units[i] {
			for _, candidate := range unit {
				// This uses linear search to ensure uniqueness, but this calculation is
				// only done once at solver creation so we don't particularly care about
				// its speed.
				if candidate != i && slices.Index(peers[i], candidate) < 0 {
					peers[i] = append(peers[i], candidate)
				}
			}
		}
	}

	return &Sudoku{
		unitlist: unitlist,
		units:    units,
		peers:    peers,
	}
}

// parseBoard parses a Sudoku board given in textual representation, and returns
// it as Values. The textual representation is as described in
// http://norvig.com/sudoku.html: a string with a sequence of 81 runes in the
// set [0123456789.], where 0 or . mean "unassigned". All other runes in the
// string are ignored.
// This function tries to end up with a valid board, so it will call `assign`
// to assign digits specified in the inpput; this may invoke some constraint
// propagation throughout the board.
// It returns an error if there was an issue parsing the board, of if the board
// isn't a valid Sudoku board (e.g. contradictions exist).
func (s *Sudoku) parseBoard(str string) (Values, error) {
	var dgs []uint16

	// Iterate and grab only the supported runes; ignore all others.
	for _, r := range str {
		if r >= '0' && r <= '9' {
			dgs = append(dgs, uint16(r)-uint16('0'))
		} else if r == '.' {
			dgs = append(dgs, 0)
		}
	}

	if len(dgs) != 81 {
		return nil, fmt.Errorf("got only %v digits in board, want 81", len(dgs))
	}

	// Start with an empty board.
	values := emptyBoard()

	// Assign square digits based on the parsed board. Note that this runs
	// constraint propagation and may discover contradictions.
	for sq, d := range dgs {
		if d != 0 && !s.assign(values, sq, d) {
			return nil, fmt.Errorf("contradiction when assigning %v to square %v", d, sq)
		}
	}

	return values, nil
}

// assign attempts to assign digit to values[square], propagating
// constraints from the assignment. values is modified.
// It returns true if the assignment succeeded, and false if the assignment
// fails resulting in an invalid Sudoku board.
func (s *Sudoku) assign(values Values, square Index, digit uint16) bool {
	for d := uint16(1); d <= 9; d++ {
		// For each d 1..9 that's != digit, if d is set in
		// values[square], try to eliminate it.
		// TODO: iteration may be inefficient -- is there a beter way?
		if values[square].isMember(d) && d != digit {
			if !s.eliminate(values, square, d) {
				return false
			}
		}
	}
	return true
}

// eliminate removes digit from the candidates in values[square], propagating
// constraints. values is modified.
// It returns false if this results in an invalid Sudoku board; otherwise
// returns true.
func (s *Sudoku) eliminate(values Values, square Index, digit uint16) bool {
	if !values[square].isMember(digit) {
		// Already eliminated
		return true
	}

	// Remove digit from the candidates in square.
	values[square] = values[square].remove(digit)

	switch values[square].size() {
	case 0:
		// No remaining options for square -- this is a contradiction.
		return false
	case 1:
		// A single digit candidate remaining in the square -- this creates a new
		// constraint. Eliminate this digit from all peer squares.
		remaining := values[square].singleMemberDigit()
		for _, peer := range s.peers[square] {
			if !s.eliminate(values, peer, remaining) {
				return false
			}
		}
	}

	// Since digit was eliminated from square, it's possible that we'll find a
	// position for this digit in one of the units the square belongs to.
	for _, unit := range s.units[square] {
		// dplaces is a list of squares in this unit that have 'digit' as one of
		// their candidates.
		var dplaces []Index
		for _, sq := range unit {
			if values[sq].isMember(digit) {
				dplaces = append(dplaces, sq)
			}
		}
		if len(dplaces) == 0 {
			// Contradiction: no places left in this unit for 'digit'
			return false
		} else if len(dplaces) == 1 {
			// There's only a single place left in the unit for 'digit' to go, so
			// assign it.
			if !s.assign(values, dplaces[0], digit) {
				return false
			}
		}
	}

	return true
}

// display returns a Sudoku 2D board representation of values
func (s *Sudoku) display(values Values) string {
	// Find maximum length of one square.
	var maxlen int = 0
	for _, d := range values {
		if d.size() > maxlen {
			maxlen = d.size()
		}
	}
	width := maxlen + 1

	line := strings.Join([]string{
		strings.Repeat("-", width*3),
		strings.Repeat("-", width*3),
		strings.Repeat("-", width*3)}, "+")

	var sb strings.Builder
	for sq, d := range values {
		fmt.Fprintf(&sb, "%[1]*s", -width, fmt.Sprintf("%[1]*s", (width+d.size())/2, d))
		if sq%9 == 2 || sq%9 == 5 {
			sb.WriteString("|")
		}
		if sq%9 == 8 {
			sb.WriteRune('\n')
		}
		if sq == 26 || sq == 53 {
			sb.WriteString(line + "\n")
		}
	}
	return sb.String()
}

// emptyBoard creates an "empty" Sudoku board, where each square can potentially
// contain any digit.
func emptyBoard() Values {
	vals := make(Values, 81)
	for sq := range vals {
		vals[sq] = fullDigitsSet()
	}
	return vals
}

// isSolved checks whether values is a properly solved Sudoku board.
func (s *Sudoku) isSolved(values Values) bool {
	for _, unit := range s.unitlist {
		var dset Digits
		for _, sq := range unit {
			// Some squares have more than a single candidate? Not solved.
			if values[sq].size() != 1 {
				return false
			}
			dset = dset.add(values[sq].singleMemberDigit())
		}
		// Not all digits covered by this unit? Not solved.
		if dset != fullDigitsSet() {
			return false
		}
	}
	return true
}

// solveBoard solves a Sudoku board given in textual representation.
// It returns an error if there was an issue parsing or solving the board.
func (s *Sudoku) solveBoard(str string) (Values, error) {
	values, err := s.parseBoard(str)
	if err != nil {
		return values, err
	}

	vresult, solved := s.search(values)
	if solved {
		return vresult, nil
	} else {
		return vresult, fmt.Errorf("board is not solvable")
	}
}

// search runs a backtracking search to solve the board given in values.
// It returns true and the solved values if the search succeeded and we ended up
// with a board with only a single candidate per square; otherwise, it returns
// false.
func (s *Sudoku) search(values Values) (Values, bool) {
	// Find the next square to try assignment in: this would be the square with
	// more than 1 digit candidate, but the smallest number of such candidates.
	var squareToTry Index = -1
	var minSize int = 9
	for sq, d := range values {
		if d.size() > 1 && d.size() < minSize {
			minSize = d.size()
			squareToTry = sq
		}
	}

	// If we didn't find any square with more than one candidate, the board is
	// solved!
	if squareToTry == -1 {
		return values, true
	}

	// TODO: inefficient iteration again
	for d := uint16(1); d <= 9; d++ {
		// Try to assign sq with each one of its candidate digits. If this results
		// in a successful search() - we've solved the board!
		if values[squareToTry].isMember(d) {
			vcopy := slices.Clone(values)
			if s.assign(vcopy, squareToTry, d) {
				if vresult, solved := s.search(vcopy); solved {
					return vresult, true
				}
			}
		}
	}
	return values, false
}
