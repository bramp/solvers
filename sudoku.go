package main

import (
	"fmt"
	"strings"
)

const width = 9
const height = 9
const max = 9

//type Grid [][]int
type Grid struct {
	grid    [][]int
	choices [][]uint // cache of choices at x,y
}

func MakeGrid(grid [][]int) Grid {
	// TODO some kind of assertions as to each row having same cols

	// Defensively copy the grid, so we don't share the passed in slice.
	return Grid{
		grid: grid,
	}.Clone()
}

var grid = MakeGrid([][]int{
	//0 1  2  3  4  5  6  7  8
	{0, 9, 0, 0, 0, 0, 8, 5, 3}, // 0
	{0, 0, 0, 8, 0, 0, 0, 0, 4}, // 1
	{0, 0, 8, 2, 0, 3, 0, 6, 9}, // 2
	{5, 7, 4, 0, 0, 2, 0, 0, 0}, // 3
	{0, 0, 0, 0, 0, 0, 0, 0, 0}, // 4
	{0, 0, 0, 9, 0, 0, 6, 3, 7}, // 5
	{9, 4, 0, 1, 0, 8, 5, 0, 0}, // 6
	{7, 0, 0, 0, 0, 6, 0, 0, 0}, // 7
	{6, 8, 2, 0, 0, 0, 0, 9, 0}, // 8
})

func (g Grid) At(x, y int) int {
	return g.grid[y][x]
}

func (g Grid) Set(x, y, value int) {
	g.grid[y][x] = value

	//g.invalidCache(x, y)
}

func (g Grid) invalidCache(x, y int) {
	// Row
	for xx := 0; xx < width; xx++ {
		g.choices[y][xx] = 0
	}

	// Column
	for yy := 0; yy < height; yy++ {
		g.choices[yy][x] = 0
	}

	// The area (3x3 grid)
	for xx := x / 3 * 3; xx < x/3*3+3; xx++ {
		for yy := y / 3 * 3; yy < y/3*3+3; yy++ {
			g.choices[yy][xx] = 0
		}
	}
}

func (g Grid) Clone() Grid {
	grid := make([][]int, len(g.grid))
	choices := make([][]uint, len(g.grid))
	for y := range grid {
		grid[y] = make([]int, len(g.grid[y]))
		choices[y] = make([]uint, len(g.grid[y]))

		copy(grid[y], g.grid[y])
	}
	return Grid{
		grid:    grid,
		choices: choices,
	}
}

// Choices returns the possible values that posititon x, y could be.
func (g Grid) Choices(x, y int) uint {
	if g.grid[y][x] != 0 {
		return 0
	}
	//if g.choices[y][x] != 0 {
	//	return g.choices[y][x]
	//}

	var found uint

	// Check the row
	for xx := 0; xx < width; xx++ {
		found |= 1 << (g.grid[y][xx])
	}

	// Check the column
	for yy := 0; yy < height; yy++ {
		found |= 1 << (g.grid[yy][x])
	}

	// Check the area (3x3 grid)
	// TODO Unroll this loop
	for xx := x / 3 * 3; xx < x/3*3+3; xx++ {
		for yy := y / 3 * 3; yy < y/3*3+3; yy++ {
			found |= 1 << (g.grid[yy][xx])
		}
	}

	// Shift the 0 (1st index) off the bottom, as that isn't a valid value in the puzzle.
	// Negate the value (to be not founds)
	// Then return the indexes.
	g.choices[y][x] = ^(found >> 1) // TODO Mask off the higher bits
	return g.choices[y][x]
}

func (g Grid) String() string {
	var s strings.Builder
	for y := 0; y < len(g.grid); y++ {
		s.WriteString("{")
		for x := 0; x < len(g.grid[y]); x++ {
			if x < len(g.grid[y])-1 {
				fmt.Fprintf(&s, "%d, ", g.grid[y][x])
			} else {
				fmt.Fprintf(&s, "%d", g.grid[y][x])
			}
		}
		s.WriteString("},\n")
	}
	return s.String()
}

type SudokuSolver struct {
	solutions []Grid // solutions
}

// Simple backtrack solver
func (s *SudokuSolver) Solve(grid Grid) []Grid {
	s.solve(grid, 0, 0)
	return s.solutions
}

func (s *SudokuSolver) solve(grid Grid, x, y int) {
	if y >= height {
		// found a solution!
		s.solutions = append(s.solutions, grid.Clone())
		return
	}

	nx, ny := s.next(x, y)
	if grid.At(x, y) != 0 {
		// Fixed value, move on
		s.solve(grid, nx, ny)
		return
	}

	// Modify the grid on each choice, and revert it if we have to back trace
	//for _, choice := range grid.Choices(x, y) {
	choices := grid.Choices(x, y)
	for i := 1; i <= max; i++ { // TODO I think there is a quicker way to find the index of all set bits
		if choices&1 == 1 {
			grid.Set(x, y, i)
			s.solve(grid, nx, ny)
		}
		choices >>= 1
	}

	grid.Set(x, y, 0)

	// No solution (back track)
	return
}

// next returns the next x and y coordinate to look at. If there are no more, the returned y >= height.
func (s *SudokuSolver) next(x, y int) (int, int) {
	nextX := x + 1
	nextY := y

	if nextX >= width {
		nextX = 0
		nextY++
	}

	return nextX, nextY
}

func main() {
	var sudoku SudokuSolver
	solutions := sudoku.Solve(grid)

	if len(solutions) == 0 {
		fmt.Printf("No solution\n")
	} else {
		for i, solution := range solutions {
			fmt.Printf("Solution %d:\n%s\n\n", i+1, solution)
		}
	}
}
