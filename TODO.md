# occam2go — Implementation Status

## Fully Implemented

### Core Constructs
- **SEQ** — Sequential execution, with replicators (`SEQ i = 0 FOR n`)
- **PAR** — Parallel execution via goroutines + sync.WaitGroup, with replicators
- **IF** — Multi-branch conditionals, maps to if/else if chains
- **WHILE** — Loops, maps to Go `for` loops
- **CASE** — Pattern matching with multiple cases and ELSE branch
- **ALT** — Channel alternation, maps to Go `select`; supports boolean guards and timer timeouts
- **SKIP** — No-op process

### Data Types & Declarations
- **INT, BYTE, BOOL, REAL, REAL32, REAL64** — Scalar types (REAL/REAL64 map to float64, REAL32 maps to float32)
- **Variable declarations** — `INT x, y, z:`
- **Arrays** — `[n]TYPE arr:` with index expressions
- **Channels** — `CHAN OF TYPE c:` with send (`!`) and receive (`?`)
- **Channel arrays** — `[n]CHAN OF TYPE cs:` with indexed send/receive and `[]CHAN OF TYPE` proc params
- **Timers** — `TIMER tim:` with reads and `AFTER` expressions

### Procedures & Functions
- **PROC** — Declaration with VAL, reference, and CHAN OF parameters
- **PROC calls** — With automatic `&`/`*` for reference params, pass-through for channels
- **FUNCTION (IS form)** — `INT FUNCTION square(VAL INT x) IS x * x`
- **FUNCTION (VALOF form)** — Local declarations + VALOF body + RESULT
- **Built-in print** — `print.int`, `print.bool`, `print.string`, `print.newline`

### Expressions & Operators
- **Arithmetic** — `+`, `-`, `*`, `/`, `\` (modulo)
- **Comparison** — `=`, `<>`, `<`, `>`, `<=`, `>=`
- **Logical** — `AND`, `OR`, `NOT`
- **AFTER** — As boolean expression (maps to `>`)
- **Parenthesized expressions**
- **Array indexing** — `arr[i]`, `arr[expr]`
- **String literals** — Double-quoted strings, usable in expressions, assignments, and channel communication

---

## Not Yet Implemented

### Core Language Gaps

| Feature | Notes |
|---------|-------|
| ~~**STOP**~~ | Implemented. Maps to `fmt.Fprintln(os.Stderr, "STOP encountered")` + `select {}`. |
| ~~**Bitwise operators**~~ | Implemented. `/\` (AND), `\/` (OR), `><` (XOR), `~` (NOT), `<<` (left shift), `>>` (right shift). |
| ~~**Type conversions**~~ | Implemented. `INT x` → `int(x)`, `BYTE n` → `byte(n)`, `REAL x` → `float64(x)`. |

### Data Structures

| Feature | Notes |
|---------|-------|
| ~~**Record types**~~ | Implemented. `RECORD POINT { INT x: INT y: }` → `type POINT struct { x int; y int }`. Field access via bracket syntax (`p[x]` → `p.x`). |
| ~~**Channel arrays**~~ | Implemented. `[n]CHAN OF TYPE cs:` → `make([]chan T, n)` + init loop. Indexed send/receive (`cs[i] ! x`, `cs[i] ? x`), `[]CHAN OF TYPE` proc params, and ALT with indexed channels. |
| ~~**REAL32 / REAL64**~~ | Implemented. `REAL32` maps to `float32`, `REAL64` maps to `float64`. |

### Channel & Protocol Features

| Feature | Notes |
|---------|-------|
| ~~**Protocols**~~ | Implemented. Simple (`PROTOCOL X IS INT`), sequential (`PROTOCOL X IS INT ; BYTE`), and variant (`PROTOCOL X CASE tag; INT ...`) protocols on channels. |
| **Channel direction** | Restricting channel params to input-only (`?`) or output-only (`!`). Currently all channel params are bidirectional. |

### Language Constructs

| Feature | Notes |
|---------|-------|
| **Abbreviations** | `name IS expr:` and `VAL name IS expr:` for aliasing. Only partially used in FUNCTION IS form. |
| **PRI ALT / PRI PAR** | Priority variants of ALT and PAR. |
| **Complex ALT guards** | Only simple boolean + channel guards work currently. |
| **PLACED PAR** | Assigning processes to specific hardware. |
| **PORT OF** | Hardware port mapping. |

---

## Suggested Priority

1. ~~**Channel arrays**~~ — Implemented
2. ~~**STOP**~~ — Implemented
3. ~~**Bitwise operators**~~ — Implemented
4. ~~**Protocols**~~ — Implemented
5. ~~**Record types**~~ — Implemented
