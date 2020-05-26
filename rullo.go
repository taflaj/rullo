// rullo.go

// A tool to solve a Rullo board.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/taflaj/util/reader"
)

// Row represents a row within a board.
type Row []int

// Sum sums all elements of a row.
func (row Row) Sum() int {
	sum := 0
	for _, v := range row {
		sum += v
	}
	return sum
}

// Board contains a series of rows forming a board.
type Board []Row

// Duplicate creates an exact copy of the board on a separate space.
func (board Board) Duplicate() Board {
	rows := len(board)
	dup := Board(make(Board, rows))
	for i := 0; i < rows; i++ {
		dup[i] = Row(make([]int, len(board[i])))
		copy(dup[i], board[i])
	}
	return dup
}

// Sum sums a given column of a board.
func (board Board) Sum(col int) int {
	sum := 0
	for _, r := range board {
		sum += r[col]
	}
	return sum
}

// Array contains an array of rows with valid solutions.
type Array []Row

// Append adds a new plausible solution to a given row.
func (a *Array) Append(solution Row) {
	*a = append(*a, solution)
}

// Plausibles contains all rows with valid solutions.
type Plausibles []Array

// Assemble returns a board with a plausible solution.
func (p Plausibles) Assemble(c chan Board, board *Board, rowNo int) {
	if rowNo >= len(p) { // board is assembled
		c <- (*board).Duplicate() // to avoid data contamination
	} else {
		for _, row := range p[rowNo] { // choose each plausible solution on this row
			(*board)[rowNo] = row
			p.Assemble(c, board, rowNo+1) // proceed with the following row
		}
	}
}

// Iterate is a goroutine that returns all combinations of plausible solutions.
func (p Plausibles) Iterate(c chan Board) {
	board := make(Board, len(p))
	p.Assemble(c, &board, 0)
	close(c)
}

// Converts a string to its numeric value.
func convert(line int, value string) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("Line #%v contains a non-numeric value: \"%v\"", line, value))
	}
	return n
}

// This goroutine reads data from the input file and converts it to int slices.
func load(c chan []int, in *reader.LineReader) {
	first := true
	finished := false
	lineNo := 0
	var cols, rows, rowNo int
	var horz []int // horizontal sums
	for {
		raw, ok := in.ReadLine()
		if !ok {
			break
		}
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
	close(c)
}

// Creates a new board with data from the input file.
func newBoard(file string) (Board, []int, []int) {
	data := make(chan []int)
	go load(data, reader.NewLineReader(file))
	row := <-data // first row contains board dimensions
	rows := row[1]
	board := Board(make(Board, rows))
	for i := 0; i < rows; i++ { // load rows
		board[i] = Row(<-data)
	}
	horz := <-data   // horizontal sums
	vert := <-data   // vertical sums
	for range data { // discard anything that follows
		continue
	}
	return board, horz, vert
}

// Explores all possible solutions using brute force.
func explore(board Board, horz []int, vert []int) {
	rows := len(board)
	cols := len(board[0])
	plausibles := make(Plausibles, rows)
	// find all plausible solutions for each row
	for r := 0; r < rows; r++ {
		plausibles[r] = Array{}
		for i := 0; i < 1<<uint(cols); i++ { // exercise all combinations
			k := i
			row := make(Row, cols)
			for j := 0; j < cols; j++ {
				if k%2 == 0 {
					row[j] = board[r][j]
				}
				k >>= 1
			}
			if row.Sum() == horz[r] { // we have a plausible solution
				plausibles[r].Append(row)
			}
		}
	}
	// match all plausible solutions with one another.
	n := 0
	solutions := make(chan Board, 100)
	go plausibles.Iterate(solutions)
	for solution := range solutions {
		// see if the newly assembled board is an actual solution
		solved := true
		for col := range solution[0] {
			if solution.Sum(col) != vert[col] {
				solved = false // this column failed validation
				break
			}
		}
		if solved {
			n++
			fmt.Printf("%3v: %v\n", n, solution)
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
		explore(board, horz, vert)
	}
}
