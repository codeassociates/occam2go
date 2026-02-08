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

See [TODO.md](TODO.md) for the full implementation status and roadmap.

| Occam | Go |
|-------|-----|
| `INT`, `BYTE`, `BOOL`, `REAL` | `int`, `byte`, `bool`, `float64` |
| `SEQ` | Sequential code |
| `PAR` | Goroutines with `sync.WaitGroup` |
| `IF` | `if / else if` |
| `WHILE` | `for` loop |
| `STOP` | Print to stderr + `select {}` (deadlock) |
| `PROC` with `VAL` params | Functions with value/pointer params |
| `:=` assignment | `=` assignment |
| Arithmetic: `+`, `-`, `*`, `/`, `\` | `+`, `-`, `*`, `/`, `%` |
| Comparison: `=`, `<>`, `<`, `>`, `<=`, `>=` | `==`, `!=`, `<`, `>`, `<=`, `>=` |
| Logic: `AND`, `OR`, `NOT` | `&&`, `\|\|`, `!` |
| Bitwise: `/\`, `\/`, `><`, `~` | `&`, `\|`, `^`, `^` (AND, OR, XOR, NOT) |
| Shifts: `<<`, `>>` | `<<`, `>>` |
| Type conversions: `INT x`, `BYTE n` | `int(x)`, `byte(n)` |

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

### Protocols

Protocols define the type of data carried on a channel. Three forms are supported:

| Occam | Go |
|-------|-----|
| `PROTOCOL SIG IS INT` | `type _proto_SIG = int` |
| `PROTOCOL PAIR IS INT ; BYTE` | `type _proto_PAIR struct { _0 int; _1 byte }` |
| `PROTOCOL MSG CASE tag; INT ...` | Interface + concrete structs per tag |
| `c ! 42 ; 65` (sequential send) | `c <- _proto_PAIR{42, 65}` |
| `c ? x ; y` (sequential recv) | `_tmp := <-c; x = _tmp._0; y = _tmp._1` |
| `c ! tag ; val` (variant send) | `c <- _proto_MSG_tag{val}` |
| `c ? CASE ...` (variant recv) | `switch _v := (<-c).(type) { ... }` |

Sequential protocol example:
```occam
PROTOCOL PAIR IS INT ; INT

SEQ
  CHAN OF PAIR c:
  INT x, y:
  PAR
    c ! 10 ; 20
    c ? x ; y
  print.int(x + y)
```

Variant protocol example:
```occam
PROTOCOL MSG
  CASE
    data; INT
    quit

SEQ
  CHAN OF MSG c:
  INT result:
  PAR
    c ! data ; 42
    c ? CASE
      data ; result
        print.int(result)
      quit
        SKIP
```

### Arrays

| Occam | Go |
|-------|-----|
| `[5]INT arr:` | `arr := make([]int, 5)` |
| `arr[i] := x` | `arr[i] = x` |
| `x := arr[i]` | `x = arr[i]` |

Example:
```occam
SEQ
  [5]INT arr:
  SEQ i = 0 FOR 5
    arr[i] := (i + 1) * 10
  INT sum:
  sum := 0
  SEQ i = 0 FOR 5
    sum := sum + arr[i]
  print.int(sum)
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

2. **Protocol types**: Simple, sequential, and variant protocols are supported. Nested protocols (protocols referencing other protocols) are not yet supported.

3. **Channel arrays**: Occam allows arrays of channels. Not yet implemented.

4. **ALT construct**: Occam's `ALT` maps to Go's `select` statement. Basic ALT, guards, and timer timeouts are supported. Priority ALT (`PRI ALT`) and replicated ALT are not yet implemented.

## How PAR is Mapped

Occam's `PAR` construct runs processes truly in parallel. On the Transputer this was hardware-scheduled; in Go it maps to goroutines coordinated with a `sync.WaitGroup`.

### Basic PAR

Each branch of a `PAR` block becomes a goroutine. The transpiler inserts a `WaitGroup` to ensure all branches complete before execution continues:

```occam
PAR
  c ! 42
  c ? x
```

Generates:

```go
var wg sync.WaitGroup
wg.Add(2)
go func() {
    defer wg.Done()
    c <- 42
}()
go func() {
    defer wg.Done()
    x = <-c
}()
wg.Wait()
```

The `wg.Wait()` call blocks until all goroutines have finished, preserving Occam's semantics that execution only continues after all parallel branches complete.

### Replicated PAR

A replicated `PAR` spawns N concurrent processes using a loop. Each iteration captures the loop variable to avoid closure issues:

```occam
PAR i = 0 FOR 4
  c ! i
```

Generates:

```go
var wg sync.WaitGroup
wg.Add(int(4))
for i := 0; i < 0 + 4; i++ {
    i := i  // capture loop variable
    go func() {
        defer wg.Done()
        c <- i
    }()
}
wg.Wait()
```

### Differences and Limitations

1. **Scheduling**: Occam on the Transputer had deterministic, priority-based scheduling. Go's goroutine scheduler is preemptive and non-deterministic. Programs that depend on execution order between `PAR` branches may behave differently.

2. **Shared memory**: Occam enforces at compile time that parallel processes do not share variables (the "disjointness" rule). The transpiler does not enforce this, so generated Go code may contain data races if the original Occam would have been rejected by a full Occam compiler.

3. **PLACED PAR**: Occam's `PLACED PAR` for assigning processes to specific Transputer links or processors is not supported.

## How Timers are Mapped

Occam's `TIMER` provides access to a hardware clock. The transpiler maps timer operations to Go's `time` package.

### Timer Declaration

Timer declarations are no-ops in the generated code since Go accesses time through the `time` package directly:

```occam
TIMER tim:
```

Generates:

```go
// TIMER tim
```

### Reading the Current Time

A timer read stores the current time as an integer (microseconds since epoch):

```occam
TIMER tim:
INT t:
tim ? t
```

Generates:

```go
// TIMER tim
var t int
t = int(time.Now().UnixMicro())
```

### Timer Timeouts in ALT

Timer cases in ALT allow a process to wait until a deadline. This maps to Go's `time.After` inside a `select`:

```occam
TIMER tim:
INT t:
tim ? t
ALT
  c ? x
    process(x)
  tim ? AFTER (t + 100000)
    handle.timeout()
```

Generates:

```go
// TIMER tim
var t int
t = int(time.Now().UnixMicro())
select {
case x = <-c:
    process(x)
case <-time.After(time.Duration((t + 100000) - int(time.Now().UnixMicro())) * time.Microsecond):
    handle_timeout()
}
```

The deadline expression `(t + 100000)` represents an absolute time. The generated code computes the remaining duration by subtracting the current time.

### AFTER as a Boolean Expression

The `AFTER` operator compares two time values and evaluates to `true` if the left operand is later than the right. It maps to `>`:

```occam
IF
  t2 AFTER t1
    -- t2 is later
```

Generates:

```go
if (t2 > t1) {
    // t2 is later
}
```

### Differences and Limitations

1. **Clock resolution**: Occam timers are hardware-dependent (often microsecond resolution on the Transputer). The transpiler uses `time.Now().UnixMicro()` for microsecond values, but actual resolution depends on the OS.

2. **Guarded timer ALT**: `guard & tim ? AFTER deadline` (timer cases with boolean guards) is not yet supported.

3. **Clock wraparound**: Occam's `AFTER` operator handles 32-bit clock wraparound correctly. The transpiler uses a simple `>` comparison, which does not handle wraparound.