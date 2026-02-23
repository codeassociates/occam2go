# occam2go

Transpiler from occam (a CSP-based concurrent programming language) to Go.

## Build & Test

```bash
go build -o occam2go .          # build the binary
go test ./...                    # run all tests (includes e2e: transpile → compile → run)
go test ./parser                 # parser unit tests only
go test ./codegen                # codegen unit + e2e tests only
go test ./lexer                  # lexer unit tests only
go test ./codegen -run TestE2E   # e2e tests only
```

Usage:
```bash
./occam2go [-o output.go] [-I includepath]... [-D SYMBOL]... input.occ
./occam2go gen-module [-o output] [-name GUARD] <SConscript>
```

Example with `#INCLUDE`:
```bash
./occam2go -I examples -o include_demo.go examples/include_demo.occ
go run include_demo.go
```

## Architecture

```
preproc/ → lexer/ → parser/ → ast/ → codegen/
```

Six packages, one pipeline:

1. **`preproc/`** — Textual preprocessor (pre-lexer pass). Handles `#IF`/`#ELSE`/`#ENDIF`/`#DEFINE` conditional compilation, `#INCLUDE` file inclusion with search paths, and ignores `#COMMENT`/`#PRAGMA`/`#USE`. Produces a single expanded string for the lexer.
   - `preproc.go` — Preprocessor with condition stack and expression evaluator

2. **`lexer/`** — Tokenizer with indentation tracking. Produces `INDENT`/`DEDENT` tokens from whitespace changes (2-space indent = 1 level). Suppresses INDENT/DEDENT/NEWLINE inside parentheses (`parenDepth` tracking, like Python). Key files:
   - `token.go` — Token types and keyword lookup
   - `lexer.go` — Lexer with `indentStack`, `pendingTokens` queue, and `parenDepth` counter

3. **`parser/`** — Recursive descent parser with Pratt expression parsing. Produces AST.
   - `parser.go` — All parsing logic in one file

4. **`ast/`** — AST node definitions. Every construct has a struct.
   - `ast.go` — All node types: `Program`, `SeqBlock`, `ParBlock`, `VarDecl`, `Assignment`, `ProcDecl`, `FuncDecl`, etc.

5. **`codegen/`** — AST → Go source code. Two-pass: first collects metadata (imports, proc signatures), then generates.
   - `codegen.go` — Generator with `strings.Builder` output
   - `codegen_test.go` — Unit tests (transpile, check output strings)
   - `e2e_test.go` — End-to-end tests (transpile → `go build` → execute → check stdout)

6. **`modgen/`** — Generates `.module` files from KRoC SConscript build files. Uses regex-based pattern matching (not Python execution) to extract `Split('''...''')` source lists and `OccamLibrary` calls. Only works with simple, declarative SConscript files; files using Python control flow (loops, conditionals) are not supported.
   - `modgen.go` — SConscript parser and module file generator

7. **`main.go`** — CLI entry point wiring the pipeline together

## Occam → Go Mapping

| Occam | Go |
|---|---|
| `SEQ` | Sequential statements (Go default) |
| `SEQ i = 0 FOR n` | `for i := 0; i < n; i++` |
| `SEQ i = 0 FOR n STEP s` | Counter-based `for` with `i := start + counter * s` |
| `PAR` | goroutines + `sync.WaitGroup` |
| `PAR i = 0 FOR n` | Loop spawning goroutines + WaitGroup |
| `IF` (multi-branch) | `if / else if` chain |
| `WHILE cond` | `for cond` |
| `CASE x` | `switch x` |
| `STOP` | `fmt.Fprintln(os.Stderr, ...)` + `select {}` |
| `ALT` | `select` |
| `ALT i = 0 FOR n` | `reflect.Select` with runtime case slice |
| `CHAN OF INT c:` | `c := make(chan int)` |
| `c ! expr` | `c <- expr` |
| `c ? x` | `x = <-c` |
| `PROC name(...)` | `func name(...)` |
| `INT FUNCTION name(...) IS expr` | `func name(...) int { return expr }` |
| `INT INLINE FUNCTION name(...)` | `func name(...) int { ... }` (INLINE ignored) |
| `INT, INT FUNCTION name(...)` | `func name(...) (int, int) { ... }` |
| `RESULT expr1, expr2` | `return expr1, expr2` |
| `a, b := func(...)` | `a, b = func(...)` (multi-assignment) |
| `x[0], x[1] := x[1], x[0]` | `x[0], x[1] = x[1], x[0]` (indexed multi-assignment) |
| `TIMER` / `tim ? t` | `time.Now().UnixMicro()` |
| `=` / `<>` | `==` / `!=` |
| `AND` / `OR` / `NOT` | `&&` / `||` / `!` |
| `REAL32 x:` / `REAL64 x:` | `var x float32` / `var x float64` |
| `INT16 x:` / `INT32 x:` / `INT64 x:` | `var x int16` / `var x int32` / `var x int64` |
| `INT expr`, `BYTE expr`, etc. | `int(expr)`, `byte(expr)`, etc. (type conversions) |
| `INT16 expr` / `INT32 expr` / `INT64 expr` | `int16(expr)` / `int32(expr)` / `int64(expr)` (type conversions) |
| `REAL32 expr` / `REAL64 expr` | `float32(expr)` / `float64(expr)` (type conversions) |
| `INT ROUND expr` (float→int) | `int(math.Round(float64(expr)))` |
| `INT TRUNC expr` (float→int) | `int(expr)` (Go default truncates) |
| `REAL32 ROUND expr` / `REAL32 TRUNC expr` | `float32(expr)` (qualifier irrelevant for int→float) |
| `BOOL expr` (numeric→bool) | `((expr) != 0)` |
| `INT boolExpr` (bool→numeric) | `_boolToInt(expr)` / `goType(_boolToInt(expr))` |
| `PROTOCOL X IS INT` | `type _proto_X = int` (simple protocol) |
| `PROTOCOL X IS INT ; BYTE` | `type _proto_X struct { _0 int; _1 byte }` (sequential) |
| `PROTOCOL X CASE tag; INT ...` | Interface + concrete structs per tag (variant) |
| `c ! 42 ; 65` (sequential send) | `c <- _proto_X{42, 65}` |
| `c ? x ; y` (sequential recv) | `_tmp := <-c; x = _tmp._0; y = _tmp._1` |
| `c ! tag ; val` (variant send) | `c <- _proto_X_tag{val}` |
| `c ? CASE ...` (variant recv) | `switch _v := (<-c).(type) { ... }` |
| `RECORD POINT { INT x: }` | `type POINT struct { x int }` |
| `POINT p:` | `var p POINT` |
| `p[x] := 10` (field assign) | `p.x = 10` |
| `p[x]` (field access) | `p.x` |
| `PLUS` / `MINUS` / `TIMES` | `+` / `-` / `*` (modular/wrapping arithmetic) |
| `\` (modulo) | `%` |
| `/\` / `\/` / `><` | `&` / `\|` / `^` (bitwise AND/OR/XOR) |
| `~` | `^` (bitwise NOT) |
| `<<` / `>>` | `<<` / `>>` (shifts) |
| `[5]CHAN OF INT cs:` | `cs := make([]chan int, 5)` + init loop |
| `cs[i] ! 42` | `cs[i] <- 42` |
| `cs[i] ? x` | `x = <-cs[i]` |
| `PROC f([]CHAN OF INT cs)` | `func f(cs []chan int)` |
| `PROC f([]CHAN OF INT cs?)` | `func f(cs []chan int)` (direction dropped for array params) |
| `PROC f([]CHAN OF INT cs!)` | `func f(cs []chan int)` (direction dropped for array params) |
| `PROC f(CHAN OF INT c?)` | `func f(c <-chan int)` (input/receive-only) |
| `PROC f(CHAN OF INT c!)` | `func f(c chan<- int)` (output/send-only) |
| `f(out!, in?)` (call-site dir) | `f(out, in)` (direction annotations ignored) |
| Non-VAL params | `*type` pointer params, callers pass `&arg` |
| `PROC f([]INT arr)` | `func f(arr []int)` (open array param, slice) |
| `PROC f(VAL []INT arr)` | `func f(arr []int)` (VAL open array, also slice) |
| `PROC f([2]INT arr)` | `func f(arr *[2]int)` (fixed-size array param) |
| `PROC f(RESULT INT x)` | `func f(x *int)` (RESULT qualifier, same as non-VAL) |
| `PROC f(CHAN INT a?, b?)` | Shared-type params (type applies to all until next type) |
| `VAL INT x IS 42:` | `x := 42` (abbreviation/named constant) |
| `VAL []BYTE s IS "hi":` | `var s []byte = []byte("hi")` (open array abbreviation) |
| `INT y IS z:` | `y := z` (non-VAL abbreviation) |
| `INITIAL INT x IS 42:` | `x := 42` (mutable variable with initial value) |
| `#INCLUDE "file"` | Textual inclusion (preprocessor, pre-lexer) |
| `#IF`/`#ELSE`/`#ENDIF` | Conditional compilation (preprocessor) |
| `#DEFINE SYMBOL` | Define preprocessor symbol |
| `#COMMENT`/`#PRAGMA`/`#USE` | Ignored (blank line) |
| `#FF`, `#80000000` | `0xFF`, `0x80000000` (hex integer literals) |
| `SIZE arr` / `SIZE "str"` | `len(arr)` / `len("str")` |
| `MOSTNEG INT` / `MOSTPOS INT` | `math.MinInt` / `math.MaxInt` |
| `MOSTNEG INT16` / `MOSTPOS INT16` | `math.MinInt16` / `math.MaxInt16` |
| `MOSTNEG INT32` / `MOSTPOS INT32` | `math.MinInt32` / `math.MaxInt32` |
| `MOSTNEG INT64` / `MOSTPOS INT64` | `math.MinInt64` / `math.MaxInt64` |
| `MOSTNEG BYTE` / `MOSTPOS BYTE` | `0` / `255` |
| `MOSTNEG REAL32` / `MOSTPOS REAL32` | `-math.MaxFloat32` / `math.MaxFloat32` |
| `MOSTNEG REAL64` / `MOSTPOS REAL64` | `-math.MaxFloat64` / `math.MaxFloat64` |
| `[arr FROM n FOR m]` | `arr[n : n+m]` (array slice) |
| `[arr FOR m]` | `arr[0 : m]` (shorthand slice, FROM 0 implied) |
| `[arr FROM n FOR m] := src` | `copy(arr[n:n+m], src)` (slice assignment) |
| Nested `PROC`/`FUNCTION` | `name := func(...) { ... }` (Go closure) |
| `VAL x IS 42:` (untyped) | `var x = 42` (Go type inference) |
| `[1, 2, 3]` (array literal) | `[]int{1, 2, 3}` |
| `VAL INT X RETYPES X :` | `X := int(int32(math.Float32bits(float32(X))))` |
| `VAL [2]INT X RETYPES X :` | `X := []int{lo, hi}` via `math.Float64bits` |
| `CAUSEERROR()` | `panic("CAUSEERROR")` |
| `LONGPROD` / `LONGDIV` etc. | Go helper functions using `uint64`/`math/bits` |

## Key Parser Patterns

### Indentation Tracking
- `p.indentLevel` is incremented/decremented in `nextToken()` when INDENT/DEDENT tokens pass through
- **startLevel pattern**: After consuming INDENT, save `startLevel := p.indentLevel`. Loop with `for p.curTokenIs(DEDENT) { if p.indentLevel < startLevel { return } }` to distinguish nested DEDENTs from block-ending DEDENTs
- Used in: `parseBlockStatements()`, `parseAltCases()`, `parseIfStatement()`, `parseCaseStatement()`

### Token Flow Conventions
- Callers consume the INDENT token before calling block-parsing functions
- Block-parsing functions call `p.nextToken()` first to move past INDENT to the first real token
- `parseAssignment()` / `parseExpression()` leave the cursor on the last consumed token
- After `parseStatement()` returns, callers must advance if not already at NEWLINE/DEDENT/EOF

### Expression Parsing
- Pratt parser with precedence levels: OR < AND < EQUALS < COMPARISON < SUM < PRODUCT < PREFIX < INDEX
- `parseExpression()` handles prefix (IDENT, INT, STRING, TRUE/FALSE, LPAREN, MINUS, NOT, BITNOT, INT_TYPE/BYTE_TYPE/BOOL_TYPE/REAL_TYPE for type conversions) then infix loop
- Function calls detected by `IDENT` followed by `LPAREN`

## Adding a New Feature

Typical workflow for a new language construct:

1. **Lexer** (`lexer/token.go`, `lexer/lexer.go`): Add token types and keywords if needed
2. **AST** (`ast/ast.go`): Define new node struct(s) implementing `Statement` or `Expression`
3. **Parser** (`parser/parser.go`): Add case to `parseStatement()` switch; implement parse function
4. **Codegen** (`codegen/codegen.go`): Add case to `generateStatement()` or `generateExpression()`; implement generation. If the new construct needs an import (sync, fmt, time), add a `containsX()` scanner
5. **Tests**: Add parser unit tests in `parser/parser_test.go`, codegen unit tests in `codegen/codegen_test.go`, and e2e tests in `codegen/e2e_test.go`
6. **Documentation**: Update TODO.md to reflect support for the new feature.

## What's Implemented

Preprocessor (`#IF`/`#ELSE`/`#ENDIF`/`#DEFINE`/`#INCLUDE` with search paths, include guards, include-once deduplication, `#COMMENT`/`#PRAGMA`/`#USE` ignored), module file generation from SConscript (`gen-module` subcommand), SEQ, PAR, IF, WHILE, CASE, ALT (with guards, timer timeouts, multi-statement bodies with scoped declarations, and replicators using `reflect.Select`), SKIP, STOP, variable/array/channel/timer declarations, abbreviations (`VAL INT x IS 42:`, `INT y IS z:`, `VAL []BYTE s IS "hi":`, untyped `VAL x IS expr:`), assignments (simple and indexed), channel send/receive, channel arrays (`[n]CHAN OF TYPE` with indexed send/receive and `[]CHAN OF TYPE` proc params), PROC (with VAL, RESULT, reference, CHAN, []CHAN, open array `[]TYPE`, fixed-size array `[n]TYPE`, and shared-type params), channel direction restrictions (`CHAN OF INT c?` → `<-chan int`, `CHAN OF INT c!` → `chan<- int`, call-site annotations `out!`/`in?` accepted), multi-line parameter lists and expressions (lexer suppresses INDENT/DEDENT/NEWLINE inside parens/brackets and after continuation operators), FUNCTION (IS and VALOF forms with multi-statement bodies, including multi-result `INT, INT FUNCTION` with `RESULT a, b`), multi-assignment (`a, b := func(...)` including indexed targets like `x[0], x[1] := x[1], x[0]`), KRoC-style colon terminators on PROC/FUNCTION (optional), INLINE function modifier (accepted and ignored), replicators on SEQ/PAR/IF/ALT (with optional STEP), arithmetic/comparison/logical/AFTER/bitwise operators, type conversions (`INT expr`, `INT16 expr`, `INT32 expr`, `INT64 expr`, `BYTE expr`, `BOOL expr`, `REAL32 expr`, `REAL64 expr`, including BOOL↔numeric via `_boolToInt` helper and `!= 0` comparison, and ROUND/TRUNC qualifiers for float↔int conversions), INT16/INT32/INT64 types, REAL32/REAL64 types, hex integer literals (`#FF`, `#80000000`), string literals, byte literals (`'A'`, `'*n'` with occam escape sequences), built-in print procedures, protocols (simple, sequential, and variant), record types (with field access via bracket syntax), SIZE operator, array slices (`[arr FROM n FOR m]` and shorthand `[arr FOR m]` with slice assignment), array literals (`[1, 2, 3]`), nested PROCs/FUNCTIONs (local definitions as Go closures), MOSTNEG/MOSTPOS (type min/max constants for INT, INT16, INT32, INT64, BYTE, REAL32, REAL64), INITIAL declarations (`INITIAL INT x IS 42:` — mutable variable with initial value), checked (modular) arithmetic (`PLUS`, `MINUS`, `TIMES` — wrapping operators), RETYPES (bit-level type reinterpretation: `VAL INT X RETYPES X :` for float32→int, `VAL [2]INT X RETYPES X :` for float64→int pair), transputer intrinsics (LONGPROD, LONGDIV, LONGSUM, LONGDIFF, NORMALISE, SHIFTRIGHT, SHIFTLEFT — implemented as Go helper functions), CAUSEERROR (maps to `panic("CAUSEERROR")`).

## Course Module Testing

The KRoC course module (`kroc/modules/course/libsrc/course.module`) is a real-world integration test:

```bash
# Transpile full course module (including float_io.occ)
./occam2go -I kroc/modules/course/libsrc -D TARGET.BITS.PER.WORD=32 -o /tmp/course_out.go kroc/modules/course/libsrc/course.module

# Verify Go output compiles (will only fail with "no main" since it's a library)
go vet /tmp/course_out.go
```

## Not Yet Implemented

PRI ALT/PRI PAR, PLACED PAR, PORT OF. See `TODO.md` for the full list with priorities.
