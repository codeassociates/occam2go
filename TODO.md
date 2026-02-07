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
- **INT, BYTE, BOOL, REAL** — Scalar types (REAL maps to float64)
- **Variable declarations** — `INT x, y, z:`
- **Arrays** — `[n]TYPE arr:` with index expressions
- **Channels** — `CHAN OF TYPE c:` with send (`!`) and receive (`?`)
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

---

## Not Yet Implemented

### Core Language Gaps

| Feature | Notes |
|---------|-------|
| **STOP** | Token exists in lexer but not parsed. Occam's deadlock/termination primitive. |
| **String literals** | Token type exists but no expression parsing or codegen. Only `print.string("...")` works as a built-in special case. |
| **Bitwise operators** | No bitwise AND, OR, XOR, or shift operators. |
| **Type conversions** | No explicit casting (e.g., `INT x` converting BYTE to INT). |

### Data Structures

| Feature | Notes |
|---------|-------|
| **Record types** | Structured data (like Go structs). |
| **Channel arrays** | `[n]CHAN OF TYPE` — only scalar channel declarations work. |
| **REAL32 / REAL64** | Only a single REAL type exists. Occam distinguishes the two. |

### Channel & Protocol Features

| Feature | Notes |
|---------|-------|
| **Protocols** | Simple, sequential, and variant (tagged union) protocols on channels. |
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

1. **String literals** — Unlocks text processing programs
2. **Channel arrays** — Essential for scalable concurrent patterns (e.g., worker pools with replicated PAR)
3. **STOP** — Simple to add, completes the process algebra primitives
4. **Bitwise operators** — Needed for systems-level programs
5. **Protocols** — Needed for realistic multi-message channel communication
6. **Record types** — Needed for structured data
