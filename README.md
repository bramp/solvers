# Solvers
Solvers for sudoku and more, that were written for fun, and run really quickly.

## Sudoku

```shell
go run sudoku.go
```

The sudoku is a backtracking solver, that uses a min-heap to determine the next square to try. This shrinks the search space and ensure it can solve even the hardest puzzles quickly.

## Testing

```shell
go test
```