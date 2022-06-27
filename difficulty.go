package sudoku

import (
	"fmt"
	"log"
)

// TODO: doc

// The before/after elimination distinction is very important here...
// 1. Count hints before elimination
// 2. Count hints after elimination
// 3. Count the low bound on empty rows/cols pre (or after?) elimination
// 4. Count how difficult average (maximal?) search is over a few random tries
func EvaluateDifficulty(values Values) (int, error) {
	countHits := func() int {
		hintcount := 0
		for _, d := range values {
			if d.Size() == 1 {
				hintcount++
			}
		}
		return hintcount
	}

	fmt.Println("hintcount before elimination:", countHits())

	// Count the lower bound (minimal number) of hints in rows and cols, pre
	// elimination.
	minHints := 9

	index := func(row, col int) Index {
		return row*9 + col
	}

	// ... first the rows.
	for row := 0; row < 9; row++ {
		rowCount := 0
		for col := 0; col < 9; col++ {
			if values[index(row, col)].Size() == 1 {
				rowCount++
			}
		}
		if rowCount < minHints {
			minHints = rowCount
		}
	}

	// ... then the columns.
	for col := 0; col < 9; col++ {
		colCount := 0
		for row := 0; row < 9; row++ {
			if values[index(row, col)].Size() == 1 {
				colCount++
			}
		}
		if colCount < minHints {
			minHints = colCount
		}
	}

	fmt.Println("min hints:", minHints)

	if !EliminateAll(values) {
		return 0, fmt.Errorf("contradiction in board")
	}

	fmt.Println("hintcount after elimination:", countHits())

	countSearches := func() (uint64, error) {
		_, solved := Solve(values, SolveOptions{Randomize: true})
		if !solved {
			return 0, fmt.Errorf("cannot solve")
		}

		return Stats.NumSearches, nil
	}

	EnableStats = true
	var totalSearches uint64 = 0
	iterations := 100
	for i := 0; i < iterations; i++ {
		Stats.Reset()
		count, err := countSearches()
		if err != nil {
			log.Fatal(err)
		}
		totalSearches += count
	}
	EnableStats = false

	fmt.Println("average searches:", float64(totalSearches)/float64(iterations))
	return 0, nil
}
