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
- **STOP** — Error + deadlock

### Data Types & Declarations
- **INT, BYTE, BOOL, REAL, REAL32, REAL64** — Scalar types (REAL/REAL64 map to float64, REAL32 maps to float32)
- **Variable declarations** — `INT x, y, z:`
- **Arrays** — `[n]TYPE arr:` with index expressions
- **Channels** — `CHAN OF TYPE c:` with send (`!`) and receive (`?`)
- **Channel arrays** — `[n]CHAN OF TYPE cs:` with indexed send/receive and `[]CHAN OF TYPE` proc params
- **Channel direction** — `CHAN OF INT c?` (receive-only) and `CHAN OF INT c!` (send-only)
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
- **Bitwise** — `/\`, `\/`, `><`, `~`, `<<`, `>>`
- **AFTER** — As boolean expression (maps to `>`)
- **Parenthesized expressions**
- **Array indexing** — `arr[i]`, `arr[expr]`
- **String literals** — Double-quoted strings
- **Type conversions** — `INT expr`, `BYTE expr`, `REAL32 expr`, `REAL64 expr`

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

### Required for course.module

These features are needed to transpile the KRoC course module (`kroc/modules/course/libsrc/`), listed roughly in order of priority. Features used across many course module files are marked with frequency.

| Feature | Notes | Used in |
|---------|-------|---------|
| ~~**Abbreviations**~~ | ~~`VAL INT x IS 1:`, `VAL BYTE ch IS 'A':` — named constants.~~ **DONE** | consts.inc, all .occ files |
| ~~**`CHAN BYTE` shorthand**~~ | ~~`CHAN BYTE out!` without `OF`. KRoC allows omitting `OF` for channel types.~~ **DONE** | all .occ files |
| ~~**Open array params**~~ | ~~`VAL []BYTE s`, `[]BYTE s` — unsized array/slice parameters for PROCs and FUNCTIONs. (`[]CHAN OF T` is already supported.)~~ **DONE** | utils.occ, string.occ, file_in.occ, stringbuf.occ |
| ~~**BYTE literals**~~ | ~~`'A'`, `'0'`, `' '` — single-quoted character literals.~~ **DONE** | utils.occ, file_in.occ, string.occ |
| ~~**Occam escape sequences**~~ | ~~`*n` (newline), `*c` (carriage return), `*t` (tab) — occam uses `*` not `\` for escapes in strings and byte literals.~~ **DONE** | utils.occ, file_in.occ |
| ~~**PROC terminator `:`**~~ | ~~Standalone `:` at the end of a PROC/FUNCTION body (KRoC style).~~ **DONE** | all .occ files |
| ~~**Nested PROCs/FUNCTIONs**~~ | ~~Local PROC/FUNCTION definitions inside a PROC body.~~ **DONE** | float_io.occ, stringbuf.occ |
| ~~**Multi-result FUNCTIONs**~~ | ~~`INT, INT FUNCTION f(...)` returning multiple values via `RESULT a, b`.~~ **DONE** | random.occ, utils.occ, string.occ, float_io.occ |
| ~~**Replicated IF**~~ | ~~`IF i = 0 FOR n` — replicated conditional.~~ **DONE** | utils.occ, file_in.occ, string.occ, float_io.occ |
| ~~**Hex integer literals**~~ | ~~`#FF`, `#80000000` — prefixed with `#`.~~ **DONE** | float_io.occ, stringbuf.occ |
| **Checked arithmetic** | `TIMES`, `PLUS`, `MINUS` — modular (wrapping) arithmetic operators. | demo_cycles.occ, random.occ, utils.occ |
| **`MOSTNEG INT`** | Most-negative integer constant. | utils.occ |
| **`INITIAL` declarations** | `INITIAL INT i IS 0:` — mutable variable with initial value. | stringbuf.occ |
| ~~**Array slices**~~ | ~~`[a FROM n FOR m]` — sub-array references.~~ **DONE** | string.occ, stringbuf.occ, float_io.occ |
| **Replicator STEP** | `SEQ i = n FOR m STEP -1` — step value in replicators. | stringbuf.occ |
| **Multi-assignment** | `a, b := x, y` — parallel assignment to multiple variables. | stringbuf.occ, utils.occ |

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
