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

### Language Constructs

| Feature | Notes |
|---------|-------|
| **Abbreviations** | `name IS expr:` and `VAL name IS expr:` for aliasing. Only partially used in FUNCTION IS form. |
| **PRI ALT / PRI PAR** | Priority variants of ALT and PAR. |
| **Complex ALT guards** | Only simple boolean + channel guards work currently. |
| **PLACED PAR** | Assigning processes to specific hardware. |
| **PORT OF** | Hardware port mapping. |
