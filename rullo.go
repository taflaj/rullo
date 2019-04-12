// rullo.go

// A tool to solve a Rullo board.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Converts a string to its numeric value.
func convert(line int, value string) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("Line #%v contains a non-numeric value: \"%v\"", line, value))
	}
	return n
}

// This goroutine reads data from the input file and converts it to int slices.
func load(c chan []int, s *bufio.Scanner) {
	first := true
	finished := false
	lineNo := 0
	var cols, rows, rowNo int
	var horz []int // horizontal sums
	for s.Scan() {
		raw := s.Text()
		line := strings.Fields(strings.Split(raw, "#")[0]) // remove comments and separate data
		lineNo++
		switch {
		case finished || len(line) < 2:
			continue // skip comment-only and any excess lines
		case first && len(line) == 2: // first row contains board dimensions
			first = false
			rowNo = 0
			cols = convert(lineNo, line[0])
			rows = convert(lineNo, line[1])
			horz = make([]int, rows)
			c <- []int{cols, rows}
			// from here on it's only row data
		case first:
			panic(fmt.Sprintf("First line of data (#%v) should contain the board dimensions: \"%v\"", lineNo, raw))
		default:
			n := cols
			if rowNo < rows { // another data row
				rowNo++
				n = cols + 1
			} else {
				finished = true
				c <- horz // send horizontal sums
			}
			if len(line) != n {
				panic(fmt.Sprintf("Line #%v has the wrong number of numbers.\nFound %v, expected %v.\n\"%v\"", lineNo, len(line), n, raw))
			} else {
				if !finished { // extract horizontal sum
					horz[rowNo-1] = convert(lineNo, line[cols])
				}
				row := make([]int, cols)
				for i := 0; i < cols; i++ {
					row[i] = convert(lineNo, line[i])
				}
				c <- row // if finished, this row contains the vertical sums
			}
		}
	}
	if !finished {
		panic("There are missing lines on the input file.")
	}
	if err := s.Err(); err != nil {
		panic(fmt.Sprintf("Error while reading data: %v", err))
	}
	close(c)
}

// Creates a new board with data from the input file.
func newBoard(file string) ([][]int, []int, []int) {
	f, err := os.Open(file)
	if err != nil {
		panic(fmt.Sprintf("Error opening file %v", file))
	}
	defer f.Close()
	data := make(chan []int)
	go load(data, bufio.NewScanner(f))
	row := <-data // first row contains board dimensions
	rows := row[1]
	board := make([][]int, rows)
	for i := 0; i < rows; i++ {
		board[i] = <-data
	}
	horz := <-data   // horizontal sums
	vert := <-data   // vertical sums
	for range data { // discard anything that follows
		continue
	}
	return board, horz, vert
}

// Creates an exact copy of the board.
func duplicate(board [][]int) [][]int {
	rows := len(board)
	dup := make([][]int, rows)
	for i := 0; i < rows; i++ {
		dup[i] = make([]int, len(board[i]))
		copy(dup[i], board[i])
	}
	return dup
}

// Sums all elements of a row.
func sum(row []int) int {
	sum := 0
	for _, v := range row {
		sum += v
	}
	return sum
}

// Sums a given column of a board.
func vsum(board [][]int, col int) int {
	sum := 0
	for _, r := range board {
		sum += r[col]
	}
	return sum
}

var solutions int

// Explores all possible solutions using brute force.
func explore(board [][]int, solution [][]int, r int, cols int, horz []int, vert []int) {
	if r >= len(board) { // we may have a solution
		solved := true
		for i := 0; i < cols; i++ {
			if vsum(solution, i) != vert[i] {
				solved = false // this column failed validation
				break
			}
		}
		if solved {
			solutions++
			fmt.Printf("%3v: %v\n", solutions, solution)
		}
	} else {
		for i := 0; i < 1<<uint(cols); i++ { // exercise all cells
			k := i
			row := make([]int, cols)
			for j := 0; j < cols; j++ {
				if k%2 != 0 {
					row[j] = board[r][j]
				}
				k >>= 1
			}
			if sum(row) == horz[r] { // we have a plausibe solution
				solution[r] = row
				explore(board, solution, r+1, cols, horz, vert) // try the next row
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %v <input file>\n\n", os.Args[0])
		text := "You may use your favorite text editor to create the input file.\n" +
			"Include two numbers on the first line, specifying the width and height of the grid, not including the totals.\n" +
			"On each line, include the numbers for each row, separated by spaces. The last number should be the total of the row.\n" +
			"On the last line, include the totals for each column.\n" +
			"Any text after a # is regarded as a comment.\n" +
			"For example:\n" +
			"4 3\n1 2 3 4 6\n5 6 7 8 12\n9 10 11 12 31\n15 12 10 12"
		fmt.Println(text)
	} else {
		board, horz, vert := newBoard(os.Args[1])
		fmt.Printf("board = %v\n horz = %v\n vert = %v\n", board, horz, vert)
		explore(board, duplicate(board), 0, len(board[0]), horz, vert)
	}
}
