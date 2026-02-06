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
| `WHILE` | `for` loop |
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

### ALT (Alternation)

| Occam | Go |
|-------|-----|
| `ALT` | `select` |
| `guard & c ? x` | Conditional channel with nil pattern |
| `SEQ i = 0 FOR n` | `for i := 0; i < n; i++` |
| `PAR i = 0 FOR n` | Parallel `for` loop with goroutines |

Example:
```occam
ALT
  c1 ? x
    print.int(x)
  c2 ? y
    print.int(y)
```

Generates:
```go
select {
case x = <-c1:
    fmt.Println(x)
case y = <-c2:
    fmt.Println(y)
}
```

ALT with guards (optional boolean conditions):
```occam
ALT
  enabled & c1 ? x
    process(x)
  TRUE & c2 ? y
    process(y)
```

### Replicators

Replicators allow you to repeat a block of code a specified number of times.

| Occam | Go |
|-------|-----|
| `SEQ i = 0 FOR n` | `for i := 0; i < n; i++` |
| `PAR i = 0 FOR n` | Parallel for loop with goroutines |

Example with replicated SEQ:
```occam
SEQ i = 1 FOR 5
  print.int(i)
```

This prints 1, 2, 3, 4, 5.

Example with replicated PAR (spawns n concurrent processes):
```occam
PAR i = 0 FOR 4
  c ! i
```

### Built-in I/O Procedures

| Occam | Go |
|-------|-----|
| `print.int(x)` | `fmt.Println(x)` |
| `print.bool(x)` | `fmt.Println(x)` |
| `print.string(x)` | `fmt.Println(x)` |
| `print.newline()` | `fmt.Println()` |

## Not Yet Implemented

- `IF` (guarded commands)
- Arrays

## How Channels are Mapped

Both Occam and Go draw from Tony Hoare's Communicating Sequential Processes (CSP) model, making channel communication a natural fit for transpilation.

### Conceptual Mapping

In Occam, channels are the primary mechanism for communication between parallel processes. A channel is a synchronous, unbuffered, point-to-point connection. Go channels share these characteristics by default.

| Concept | Occam | Go |
|---------|-------|-----|
| Declaration | `CHAN OF INT c:` | `c := make(chan int)` |
| Send (blocks until receiver ready) | `c ! value` | `c <- value` |
| Receive (blocks until sender ready) | `c ? variable` | `variable = <-c` |
| Synchronisation | Implicit in `!` and `?` | Implicit in `<-` |

### Synchronous Communication

Both languages use synchronous (rendezvous) communication by default:

```occam
PAR
  c ! 42      -- blocks until receiver is ready
  c ? x       -- blocks until sender is ready
```

The sender and receiver must both be ready before the communication occurs. This is preserved in the generated Go code, where unbuffered channels have the same semantics.

### Differences and Limitations

1. **Channel direction**: Occam channels are inherently unidirectional. Go channels can be bidirectional but can be restricted using types (`chan<-` for send-only, `<-chan` for receive-only). The transpiler currently generates bidirectional Go channels.

2. **Protocol types**: Occam 2 and later versions support protocol types for structured channel communication. These are not currently supported.

3. **Channel arrays**: Occam allows arrays of channels. Not yet implemented.

4. **ALT construct**: Occam's `ALT` maps to Go's `select` statement. Basic ALT and guards are supported. Priority ALT (`PRI ALT`) and replicated ALT are not yet implemented.