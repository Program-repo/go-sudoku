package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eliben/go-sudoku"
)

func main() {
	statsFlag := flag.Bool("stats", false, "enable stats for solving")
	flag.Parse()

	var totalDuration time.Duration = 0
	var maxDuration time.Duration = 0
	var totalSearches uint64 = 0
	var maxSearches uint64 = 0
	var numBoards int = 0
	var numSolved int = 0

	if *statsFlag {
		sudoku.EnableStats = true
	}

	// Expect one board per line, ignoring whitespace and lines starting with '#'.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		board := strings.TrimSpace(scanner.Text())
		if len(board) == 0 || strings.HasPrefix(board, "#") {
			continue
		}

		numBoards++

		tStart := time.Now()
		v, err := sudoku.ParseBoard(board)
		v, _ = sudoku.Solve(v, sudoku.SolveOptions{})
		if err != nil {
			log.Fatal(err)
		}
		tElapsed := time.Now().Sub(tStart)

		totalDuration += tElapsed
		if tElapsed > maxDuration {
			maxDuration = tElapsed
		}

		if sudoku.IsSolved(v) {
			numSolved++
		}

		if *statsFlag {
			totalSearches += sudoku.Stats.NumSearches
			if sudoku.Stats.NumSearches > maxSearches {
				maxSearches = sudoku.Stats.NumSearches
			}
			sudoku.Stats.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Solved %v/%v boards\n", numSolved, numBoards)
	fmt.Printf("Duration average=%-15v max=%v\n", totalDuration/time.Duration(numBoards), maxDuration)
	if *statsFlag {
		fmt.Printf("Searches average=%-15.2f max=%v\n", float64(totalSearches)/float64(numBoards), maxSearches)
	}
}
