# occam2go

A transpiler from Occam to Go, written in Go.

Occam was developed in the 1980s to support concurrent programming on the Transputer. Go, created decades later, shares similar CSP-influenced concurrency primitives. This transpiler bridges the two.

## Building

```bash
go build -o occam2go
```

## Usage

```bash
./occam2go [options] <input.occ>
```

Options:
- `-o <file>` - Write output to file (default: stdout)
- `-version` - Print version and exit

## Running an Example

Here's how to transpile, compile, and run an Occam program:

```bash
# 1. Build the transpiler (only needed once)
go build -o occam2go

# 2. Transpile an Occam file to Go
./occam2go examples/print.occ -o output.go

# 3. Compile the generated Go code
go build -o output output.go

# 4. Run the compiled program
./output
```

Or as a one-liner to see the output immediately:

```bash
./occam2go examples/print.occ -o output.go && go run output.go
```

## Example

Input (`example.occ`):
```occam
SEQ
  INT x, y:
  PAR
    x := 1
    y := 2
  x := x + y
```

Output:
```go
package main

import (
	"sync"
)

func main() {
	var x, y int
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		x = 1
	}()
	go func() {
		defer wg.Done()
		y = 2
	}()
	wg.Wait()
	x = (x + y)
}
```

## Implemented Features

| Occam | Go |
|-------|-----|
| `INT`, `BYTE`, `BOOL`, `REAL` | `int`, `byte`, `bool`, `float64` |
| `SEQ` | Sequential code |
| `PAR` | Goroutines with `sync.WaitGroup` |
| `PROC` with `VAL` params | Functions with value/pointer params |
| `:=` assignment | `=` assignment |
| Arithmetic: `+`, `-`, `*`, `/`, `\` | `+`, `-`, `*`, `/`, `%` |
| Comparison: `=`, `<>`, `<`, `>`, `<=`, `>=` | `==`, `!=`, `<`, `>`, `<=`, `>=` |
| Logic: `AND`, `OR`, `NOT` | `&&`, `\|\|`, `!` |

### Channels

| Occam | Go |
|-------|-----|
| `CHAN OF INT c:` | `c := make(chan int)` |
| `c ! x` (send) | `c <- x` |
| `c ? y` (receive) | `y = <-c` |

Example:
```occam
SEQ
  CHAN OF INT c:
  INT result:
  PAR
    c ! 42
    c ? result
  print.int(result)
```

### Built-in I/O Procedures

| Occam | Go |
|-------|-----|
| `print.int(x)` | `fmt.Println(x)` |
| `print.bool(x)` | `fmt.Println(x)` |
| `print.string(x)` | `fmt.Println(x)` |
| `print.newline()` | `fmt.Println()` |

## Not Yet Implemented

- `ALT` (alternation) → `select`
- Replicators (`PAR i = 0 FOR n`)
- Arrays
- `WHILE`, `IF` (partial)