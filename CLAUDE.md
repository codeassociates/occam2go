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

2. **`lexer/`** — Tokenizer with indentation tracking. Produces `INDENT`/`DEDENT` tokens from whitespace changes (2-space indent = 1 level). Key files:
   - `token.go` — Token types and keyword lookup
   - `lexer.go` — Lexer with `indentStack` and `pendingTokens` queue

3. **`parser/`** — Recursive descent parser with Pratt expression parsing. Produces AST.
   - `parser.go` — All parsing logic in one file

4. **`ast/`** — AST node definitions. Every construct has a struct.
   - `ast.go` — All node types: `Program`, `SeqBlock`, `ParBlock`, `VarDecl`, `Assignment`, `ProcDecl`, `FuncDecl`, etc.

5. **`codegen/`** — AST → Go source code. Two-pass: first collects metadata (imports, proc signatures), then generates.
   - `codegen.go` — Generator with `strings.Builder` output
   - `codegen_test.go` — Unit tests (transpile, check output strings)
   - `e2e_test.go` — End-to-end tests (transpile → `go build` → execute → check stdout)

6. **`modgen/`** — Generates `.module` files from KRoC SConscript build files. Parses Python-based SConscript to extract source lists and `OccamLibrary` calls.
   - `modgen.go` — SConscript parser and module file generator

7. **`main.go`** — CLI entry point wiring the pipeline together

## Occam → Go Mapping

| Occam | Go |
|---|---|
| `SEQ` | Sequential statements (Go default) |
| `SEQ i = 0 FOR n` | `for i := 0; i < n; i++` |
| `PAR` | goroutines + `sync.WaitGroup` |
| `PAR i = 0 FOR n` | Loop spawning goroutines + WaitGroup |
| `IF` (multi-branch) | `if / else if` chain |
| `WHILE cond` | `for cond` |
| `CASE x` | `switch x` |
| `STOP` | `fmt.Fprintln(os.Stderr, ...)` + `select {}` |
| `ALT` | `select` |
| `CHAN OF INT c:` | `c := make(chan int)` |
| `c ! expr` | `c <- expr` |
| `c ? x` | `x = <-c` |
| `PROC name(...)` | `func name(...)` |
| `INT FUNCTION name(...) IS expr` | `func name(...) int { return expr }` |
| `INT, INT FUNCTION name(...)` | `func name(...) (int, int) { ... }` |
| `RESULT expr1, expr2` | `return expr1, expr2` |
| `a, b := func(...)` | `a, b = func(...)` (multi-assignment) |
| `TIMER` / `tim ? t` | `time.Now().UnixMicro()` |
| `=` / `<>` | `==` / `!=` |
| `AND` / `OR` / `NOT` | `&&` / `||` / `!` |
| `REAL32 x:` / `REAL64 x:` | `var x float32` / `var x float64` |
| `INT expr`, `BYTE expr`, etc. | `int(expr)`, `byte(expr)`, etc. (type conversions) |
| `REAL32 expr` / `REAL64 expr` | `float32(expr)` / `float64(expr)` (type conversions) |
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
| `\` (modulo) | `%` |
| `/\` / `\/` / `><` | `&` / `\|` / `^` (bitwise AND/OR/XOR) |
| `~` | `^` (bitwise NOT) |
| `<<` / `>>` | `<<` / `>>` (shifts) |
| `[5]CHAN OF INT cs:` | `cs := make([]chan int, 5)` + init loop |
| `cs[i] ! 42` | `cs[i] <- 42` |
| `cs[i] ? x` | `x = <-cs[i]` |
| `PROC f([]CHAN OF INT cs)` | `func f(cs []chan int)` |
| `PROC f(CHAN OF INT c?)` | `func f(c <-chan int)` (input/receive-only) |
| `PROC f(CHAN OF INT c!)` | `func f(c chan<- int)` (output/send-only) |
| Non-VAL params | `*type` pointer params, callers pass `&arg` |
| `PROC f([]INT arr)` | `func f(arr []int)` (open array param, slice) |
| `PROC f(VAL []INT arr)` | `func f(arr []int)` (VAL open array, also slice) |
| `VAL INT x IS 42:` | `x := 42` (abbreviation/named constant) |
| `INT y IS z:` | `y := z` (non-VAL abbreviation) |
| `#INCLUDE "file"` | Textual inclusion (preprocessor, pre-lexer) |
| `#IF`/`#ELSE`/`#ENDIF` | Conditional compilation (preprocessor) |
| `#DEFINE SYMBOL` | Define preprocessor symbol |
| `#COMMENT`/`#PRAGMA`/`#USE` | Ignored (blank line) |
| `#FF`, `#80000000` | `0xFF`, `0x80000000` (hex integer literals) |
| `SIZE arr` / `SIZE "str"` | `len(arr)` / `len("str")` |

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

## What's Implemented

Preprocessor (`#IF`/`#ELSE`/`#ENDIF`/`#DEFINE`/`#INCLUDE` with search paths, include guards, `#COMMENT`/`#PRAGMA`/`#USE` ignored), module file generation from SConscript (`gen-module` subcommand), SEQ, PAR, IF, WHILE, CASE, ALT (with guards and timer timeouts), SKIP, STOP, variable/array/channel/timer declarations, abbreviations (`VAL INT x IS 42:`, `INT y IS z:`), assignments (simple and indexed), channel send/receive, channel arrays (`[n]CHAN OF TYPE` with indexed send/receive and `[]CHAN OF TYPE` proc params), PROC (with VAL, reference, CHAN, []CHAN, and open array `[]TYPE` params), channel direction restrictions (`CHAN OF INT c?` → `<-chan int`, `CHAN OF INT c!` → `chan<- int`), FUNCTION (IS and VALOF forms, including multi-result `INT, INT FUNCTION` with `RESULT a, b`), multi-assignment (`a, b := func(...)`), KRoC-style colon terminators on PROC/FUNCTION (optional), replicators on SEQ and PAR, arithmetic/comparison/logical/AFTER/bitwise operators, type conversions (`INT expr`, `BYTE expr`, `REAL32 expr`, `REAL64 expr`, etc.), REAL32/REAL64 types, hex integer literals (`#FF`, `#80000000`), string literals, byte literals (`'A'`, `'*n'` with occam escape sequences), built-in print procedures, protocols (simple, sequential, and variant), record types (with field access via bracket syntax), SIZE operator.

## Not Yet Implemented

PRI ALT/PRI PAR, PLACED PAR, PORT OF. See `TODO.md` for the full list with priorities.
