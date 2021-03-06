package main

import (
	"container/heap"
	"fmt"
	"math/bits"
	"strings"
)

const max = 9 // TODO Remove

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

type Grid struct {
	width, height int

	grid    [][]int
	choices [][]Choices // cache of choices at x,y

	// The found values in each row, col and area
	row  []uint
	col  []uint
	area []uint

	nextChoice ChoicesHeap
}

func MakeGrid(grid [][]int) *Grid {
	// TODO some kind of assertions as to each row having same cols
	// TODO Setup the width/height correct
	// TODO Check the original configuration has no invalid squares - This will break our row bit code.

	// Defensively copy the grid, so we don't share the passed in slice.
	g := &Grid{
		width:  9,
		height: 9,
		grid:   grid,
	}
	return g.Clone()
}

type Choices struct {
	choices uint // TODO Consider changing to uint16
	x, y    int
}

// ChoicesHeap is a min-heap of Choices.
type ChoicesHeap []*Choices

func (h ChoicesHeap) Len() int { return len(h) }
func (h ChoicesHeap) Less(i, j int) bool {
	return bits.OnesCount(h[i].choices) < bits.OnesCount(h[j].choices)
}
func (h ChoicesHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *ChoicesHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify
	// the slice's length, not just its contents.
	*h = append(*h, x.(*Choices))
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

func (g *Grid) updateFoundBits(x, y, value int) {

	// Keep track of the found numbers per row/col/area
	if g.grid[y][x] != 0 {
		// Remove the old value
		mask := uint(1) << (g.grid[y][x] - 1)
		g.col[x] &^= mask
		g.row[y] &^= mask
		g.area[(y/3)*3+(x/3)] &^= mask
	}

	if value != 0 {
		// Add new the value
		mask := uint(1) << (value - 1)
		g.col[x] |= mask
		g.row[y] |= mask
		g.area[(y/3)*3+(x/3)] |= mask
	}
}

func (g *Grid) Set(x, y, value int) {
	g.updateFoundBits(x, y, value)

	g.grid[y][x] = value
	g.invalidateChoices(x, y)
}

func (g *Grid) Clone() *Grid {
	grid := make([][]int, len(g.grid))
	for y := range grid {
		grid[y] = make([]int, len(g.grid[y]))
		copy(grid[y], g.grid[y])
	}

	cg := &Grid{
		width:  g.width,
		height: g.height,

		grid: grid,
	}
	cg.init()
	return cg
}

func (g *Grid) init() {
	g.row = make([]uint, g.width)
	g.col = make([]uint, g.height)
	g.area = make([]uint, g.width/3*g.height/3)

	g.initChoices()
	g.updateChoices()
}

func (g *Grid) initChoices() {
	g.choices = make([][]Choices, len(g.grid))

	for y := range g.choices {
		g.choices[y] = make([]Choices, len(g.grid[y]))

		for x := range g.choices[y] {
			g.choices[y][x].x = x
			g.choices[y][x].y = y

			// Add the squares which are missing a number
			if g.grid[y][x] == 0 {
				g.nextChoice = append(g.nextChoice, &g.choices[y][x])
			} else {
				g.updateFoundBits(x, y, g.grid[y][x])
			}
		}
	}
}

// updateChoices updates any invalid choices.
func (g *Grid) updateChoices() {
	// Recalulate invalid choices
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			g.choices[y][x].choices = g.choicesInternal(x, y)
		}
	}

	heap.Init(&g.nextChoice)
}

// Marks all choices on the same row/col and area as (x, y) as invalid.
func (g *Grid) invalidateChoices(x, y int) {
	// Row
	for xx := 0; xx < g.width; xx++ {
		g.choices[y][xx].choices = g.choicesInternal(xx, y)
	}

	// Column
	for yy := 0; yy < g.height; yy++ {
		g.choices[yy][x].choices = g.choicesInternal(x, yy)
	}

	// The area (3x3 grid)
	// TODO We could unroll this, or be smarter. We set the choices that were already set above
	for xx := x / 3 * 3; xx < x/3*3+3; xx++ {
		for yy := y / 3 * 3; yy < y/3*3+3; yy++ {
			g.choices[yy][xx].choices = g.choicesInternal(xx, yy)
		}
	}

	// Reshuffle the heap
	heap.Init(&g.nextChoice)
}

// Choices returns the possible values that posititon x, y could be.
func (g *Grid) Choices(x, y int) uint {
	if g.grid[y][x] != 0 {
		return 0
	}

	return g.choices[y][x].choices
}

func (g *Grid) choicesInternal(x, y int) uint {
	var found uint
	found |= g.col[x]
	found |= g.row[y]
	found |= g.area[(y/3)*3+(x/3)]

	// Negate the value (to be not founds), and mask off the uneeded bits.
	return (^found) & (1<<max - 1)
}

func (g *Grid) String() string {
	var s strings.Builder
	for y := 0; y < g.height; y++ {
		s.WriteString("{")
		for x := 0; x < g.width; x++ {
			if x < g.width-1 {
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
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
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
// not have already been solved. If there are no more the 3rd returned value will be false.
func (g *Grid) Next() (int, int, bool) {
	// Jump to the square that isn't solved, and has least choices
	if g.nextChoice.Len() == 0 {
		// No valid choices left
		return 0, 0, false
	}

	next := g.nextChoice[0]
	return next.x, next.y, true
}

// Pop removes the element most recently returned by next().
func (g *Grid) Pop() {
	assert(len(g.nextChoice) > 0, "too few next choices")
	heap.Pop(&g.nextChoice)
}

// Push readds the element at x,y. This element must have been popped earlier.
func (g *Grid) Push(x, y int) {
	g.nextChoice.Push(&g.choices[y][x])
	assert(len(g.nextChoice) < g.width*g.height, "too many next choices")
}

type SudokuSolver struct {
	solutions []*Grid // solutions

	Iterations int
}

// Simple backtrack solver
func (s *SudokuSolver) Solve(grid *Grid) []*Grid {
	s.Iterations = 0
	s.solutions = nil

	s.solve(grid)

	return s.solutions
}

func (s *SudokuSolver) solve(grid *Grid) {
	s.Iterations++

	x, y, more := grid.Next()

	if !more {
		// found a solution!
		s.solutions = append(s.solutions, grid.Clone())
		return // Now backtrack to find more
	}

	assert(grid.At(x, y) == 0, "solving a grid that already has a value")

	choices := grid.Choices(x, y)
	if choices > 0 {
		grid.Pop()

		// Find all set bits,
		// TODO I think there is a quicker way to find the index of all set bits
		for i := 1; i <= max; i++ {
			if choices&1 == 1 {
				grid.Set(x, y, i)
				s.solve(grid)
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
