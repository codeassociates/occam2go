# occam2go — Implementation Status

## Fully Implemented

### Core Constructs
- **SEQ** — Sequential execution, with replicators (`SEQ i = 0 FOR n`) and optional STEP
- **PAR** — Parallel execution via goroutines + sync.WaitGroup, with replicators
- **IF** — Multi-branch conditionals, maps to if/else if chains, with replicators
- **WHILE** — Loops, maps to Go `for` loops
- **CASE** — Pattern matching with multiple cases and ELSE branch
- **ALT** — Channel alternation, maps to Go `select`; supports boolean guards and timer timeouts
- **SKIP** — No-op process
- **STOP** — Error + deadlock

### Data Types & Declarations
- **INT, BYTE, BOOL, REAL, REAL32, REAL64** — Scalar types (REAL/REAL64 map to float64, REAL32 maps to float32)
- **Variable declarations** — `INT x, y, z:`
- **Arrays** — `[n]TYPE arr:` with index expressions
- **Channels** — `CHAN OF TYPE c:` with send (`!`) and receive (`?`); `CHAN BYTE` shorthand (without `OF`)
- **Channel arrays** — `[n]CHAN OF TYPE cs:` with indexed send/receive and `[]CHAN OF TYPE` proc params
- **Channel direction** — `CHAN OF INT c?` (receive-only) and `CHAN OF INT c!` (send-only)
- **Timers** — `TIMER tim:` with reads and `AFTER` expressions
- **Abbreviations** — `VAL INT x IS 1:`, `INT y IS z:` — named constants and aliases
- **INITIAL declarations** — `INITIAL INT x IS 42:` — mutable variables with initial values
- **Byte literals** — `'A'`, `'0'` with occam escape sequences (`*n`, `*c`, `*t`)
- **Hex integer literals** — `#FF`, `#80000000`

### Procedures & Functions
- **PROC** — Declaration with VAL, reference, CHAN OF, and open array (`[]TYPE`) parameters
- **PROC calls** — With automatic `&`/`*` for reference params, pass-through for channels
- **FUNCTION (IS form)** — `INT FUNCTION square(VAL INT x) IS x * x`
- **FUNCTION (VALOF form)** — Local declarations + VALOF body + RESULT
- **Multi-result FUNCTIONs** — `INT, INT FUNCTION f(...)` returning multiple values via `RESULT a, b`
- **Nested PROCs/FUNCTIONs** — Local definitions inside a PROC body, compiled as Go closures
- **KRoC-style colon terminators** — Optional `:` at end of PROC/FUNCTION body
- **Built-in print** — `print.int`, `print.bool`, `print.string`, `print.newline`

### Expressions & Operators
- **Arithmetic** — `+`, `-`, `*`, `/`, `\` (modulo)
- **Comparison** — `=`, `<>`, `<`, `>`, `<=`, `>=`
- **Logical** — `AND`, `OR`, `NOT`
- **Bitwise** — `/\`, `\/`, `><`, `~`, `<<`, `>>`
- **AFTER** — As boolean expression (maps to `>`)
- **Parenthesized expressions**
- **Array indexing** — `arr[i]`, `arr[expr]`
- **String literals** — Double-quoted strings
- **Type conversions** — `INT expr`, `BYTE expr`, `REAL32 expr`, `REAL64 expr`
- **Checked arithmetic** — `PLUS`, `MINUS`, `TIMES` — modular (wrapping) operators
- **MOSTNEG/MOSTPOS** — Type min/max constants for INT, BYTE, REAL32, REAL64
- **SIZE operator** — `SIZE arr`, `SIZE "str"` maps to `len()`
- **Array slices** — `[arr FROM n FOR m]` with slice assignment
- **Multi-assignment** — `a, b := f(...)` including indexed targets like `x[0], x[1] := x[1], x[0]`

### Protocols
- **Simple** — `PROTOCOL SIG IS INT` (type alias)
- **Sequential** — `PROTOCOL PAIR IS INT ; BYTE` (struct)
- **Variant** — `PROTOCOL MSG CASE tag; TYPE ...` (interface + concrete types)

### Records
- **RECORD** — Struct types with field access via bracket syntax (`p[x]`)

### Preprocessor
- **`#IF` / `#ELSE` / `#ENDIF`** — Conditional compilation with `TRUE`, `FALSE`, `DEFINED()`, `NOT`, equality
- **`#DEFINE`** — Symbol definition
- **`#INCLUDE`** — File inclusion with search paths and include guards
- **`#COMMENT` / `#PRAGMA` / `#USE`** — Ignored (blank lines)
- **Predefined symbols** — `TARGET.BITS.PER.WORD = 64`

### Tooling
- **gen-module** — Generate `.module` files from KRoC SConscript build files

---

## Not Yet Implemented

### Required for shared_screen module (extends course module)

| Feature | Notes | Used in |
|---------|-------|---------|
| **`DATA TYPE X IS TYPE:`** | Simple type alias (e.g. `DATA TYPE COLOUR IS BYTE:`). | shared_screen.inc |
| **`DATA TYPE X RECORD`** | Alternative record syntax (vs current `RECORD X`). | shared_screen.inc |
| **Counted array protocol** | `BYTE::[]BYTE` — length-prefixed array in protocols. | shared_screen.inc, shared_screen.occ |
| **`RESULT` param qualifier** | `RESULT INT len` on PROC params (output-only, like a write-only reference). | float_io.occ |

### Other language features

| Feature | Notes |
|---------|-------|
| **PRI ALT / PRI PAR** | Priority variants of ALT and PAR. |
| **PLACED PAR** | Assigning processes to specific hardware. |
| **PORT OF** | Hardware port mapping. |
| **`RETYPES`** | Type punning / reinterpret cast (`VAL INT X RETYPES X :`). Used in float_io.occ. |
| **`CAUSEERROR ()`** | Built-in error-raising primitive. Used in float_io.occ. |
| **Transputer intrinsics** | `LONGPROD`, `LONGDIV`, `LONGSUM`, `LONGDIFF`, `NORMALISE`, `SHIFTLEFT`, `SHIFTRIGHT`. Used in float_io.occ. |
| **`VAL []BYTE` abbreviations** | `VAL []BYTE cmap IS "0123456789ABCDEF":` — named string constants. |
| **`#PRAGMA DEFINED`** | Compiler hint to suppress definedness warnings. Can be ignored. |
