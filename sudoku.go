package main

import (
	"container/heap"
	"fmt"
	"math/bits"
	"strings"
)

const width = 9
const height = 9
const max = 9

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

type Grid struct {
	grid    [][]int
	choices [][]*ChoicesHeapItem // cache of choices at x,y

	nextChoice ChoicesHeap
}

func MakeGrid(grid [][]int) *Grid {
	// TODO some kind of assertions as to each row having same cols

	// Defensively copy the grid, so we don't share the passed in slice.
	g := &Grid{
		grid: grid,
	}
	return g.Clone()
}

type ChoicesHeapItem struct {
	choices uint // TODO Consider changing to uint16
	x, y    int
	valid   bool
}

// ChoicesHeap is a min-heap of ChoicesHeapItem.
type ChoicesHeap []*ChoicesHeapItem

func (h ChoicesHeap) Len() int { return len(h) }
func (h ChoicesHeap) Less(i, j int) bool {
	return bits.OnesCount(h[i].choices) < bits.OnesCount(h[j].choices)
}
func (h ChoicesHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *ChoicesHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify
	// the slice's length, not just its contents.
	*h = append(*h, x.(*ChoicesHeapItem))
}

func (h *ChoicesHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (g *Grid) At(x, y int) int {
	return g.grid[y][x]
}

func (g *Grid) Set(x, y, value int) {
	g.grid[y][x] = value
	g.invalidCache(x, y)
}

func (g *Grid) invalidCache(x, y int) {
	// Row
	for xx := 0; xx < width; xx++ {
		g.choices[y][xx].valid = false
	}

	// Column
	for yy := 0; yy < height; yy++ {
		g.choices[yy][x].valid = false
	}

	// The area (3x3 grid)
	for xx := x / 3 * 3; xx < x/3*3+3; xx++ {
		for yy := y / 3 * 3; yy < y/3*3+3; yy++ {
			g.choices[yy][xx].valid = false
		}
	}

	g.initCache()
}

func (g *Grid) init() {
	g.choices = make([][]*ChoicesHeapItem, len(g.grid))

	for y := range g.choices {
		g.choices[y] = make([]*ChoicesHeapItem, len(g.grid[y]))

		for x := range g.choices[y] {
			g.choices[y][x] = &ChoicesHeapItem{
				x: x,
				y: y,
			}

			if g.grid[y][x] == 0 {
				g.nextChoice = append(g.nextChoice, g.choices[y][x]) // TODO this could be a pointer to this slice entry
			}
		}
	}

	g.initCache()
}

func (g *Grid) initCache() {
	// Recalulate invalid choices
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if !g.choices[y][x].valid {
				g.choices[y][x].choices = g.choicesInternal(x, y)
				g.choices[y][x].valid = true
			}
		}
	}

	heap.Init(&g.nextChoice)
}

func (g *Grid) Clone() *Grid {
	grid := make([][]int, len(g.grid))
	for y := range grid {
		grid[y] = make([]int, len(g.grid[y]))
		copy(grid[y], g.grid[y])
	}
	cg := &Grid{
		grid: grid,
	}
	cg.init()
	return cg
}

// Choices returns the possible values that posititon x, y could be.
func (g *Grid) Choices(x, y int) uint {
	if g.grid[y][x] != 0 {
		return 0
	}

	assert(g.choices[y][x].valid, "invalid choice")
	return g.choices[y][x].choices
}

func (g *Grid) choicesInternal(x, y int) uint {
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
	return ^(found >> 1) & (1<<max - 1)

}

func (g *Grid) String() string {
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

func (g *Grid) Valid() bool {
	// Check for unsolved spots
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if g.grid[y][x] == 0 {
				return false
			}
		}
	}

	// TODO Check each row is good
	// TODO Check each col is good
	// TODO Check each area is good

	return true
}

// next returns the next x and y coordinate to look at. The next square must
// not have already been solved. If there are no more, the returned y >= height.
func (g *Grid) next() (int, int) {
	// Jump to the square that isn't solved, and has least choices
	if g.nextChoice.Len() == 0 {
		// No valid choices left
		return width, height
	}

	next := g.nextChoice[0]
	return next.x, next.y
}

func (g *Grid) Pop() {
	heap.Pop(&g.nextChoice)
}

func (g *Grid) Push(x, y int) {
	g.nextChoice.Push(g.choices[y][x])
	assert(len(g.nextChoice) < width*height, "too many next choices")
}

type SudokuSolver struct {
	solutions []*Grid // solutions

	Iterations int
}

// Simple backtrack solver
func (s *SudokuSolver) Solve(grid *Grid) []*Grid {
	s.Iterations = 0

	x, y := grid.next()
	s.solve(grid, x, y)

	return s.solutions
}

func (s *SudokuSolver) solve(grid *Grid, x, y int) {
	s.Iterations++

	if y >= height {
		// found a solution!
		s.solutions = append(s.solutions, grid.Clone())
		return // Now backtrack to find more
	}

	assert(grid.At(x, y) == 0, "solving a grid that already has a value")

	choices := grid.Choices(x, y)
	if choices > 0 {
		grid.Pop()

		for i := 1; i <= max; i++ { // TODO I think there is a quicker way to find the index of all set bits
			if choices&1 == 1 {
				grid.Set(x, y, i)

				nx, ny := grid.next()
				s.solve(grid, nx, ny)
			}
			choices >>= 1
		}

		grid.Set(x, y, 0)
		grid.Push(x, y)
	}

	// No solution (backtrack)
	return
}
func main() {

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
