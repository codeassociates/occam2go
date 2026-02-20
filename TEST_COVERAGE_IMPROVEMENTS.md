# Test Coverage Improvements

## Overview

This document tracks additions to the test suite that close coverage gaps identified by analyzing the "What's Implemented" feature list against existing tests.

## 2026-02-20: 69 New Tests

### Gap Analysis

Prior to this work, the test suite had ~124 e2e tests, ~49 codegen unit tests, ~80 parser tests, and ~9 lexer tests. Several implemented features had zero or near-zero test coverage:

- Transputer intrinsics (LONGPROD, LONGDIV, LONGSUM, LONGDIFF, NORMALISE, SHIFTRIGHT, SHIFTLEFT) — zero tests at any level
- RETYPES bit reinterpretation — parser tests only, no codegen or e2e verification
- RESULT qualifier on proc params — completely untested
- Fixed-size array params `[n]TYPE` — completely untested
- Shared-type channel params `PROC f(CHAN INT a?, b?)` — completely untested
- `VAL []BYTE s IS "hi":` abbreviation — completely untested
- ALT with boolean guards — parser test only, no e2e
- MOSTNEG/MOSTPOS for REAL32/REAL64 — codegen unit only, no e2e
- Modulo operator `\` — lexer tokenization only
- `print.string` / `print.newline` — no e2e execution
- Lexer paren/bracket depth suppression — only exercised indirectly
- Lexer continuation-operator line joining — only exercised indirectly
- Most keywords — only 11 of 33+ had lexer-level tests

### New Test Files

#### `codegen/e2e_intrinsics_test.go` — 13 tests

| Test | Feature |
|------|---------|
| `TestE2E_LONGPROD` | Basic 64-bit multiply |
| `TestE2E_LONGPRODWithCarry` | Multiply with carry addend |
| `TestE2E_LONGDIV` | Basic 64-bit divide |
| `TestE2E_LONGDIVLargeValue` | Roundtrip with LONGPROD |
| `TestE2E_LONGSUM` | Basic 64-bit add |
| `TestE2E_LONGSUMOverflow` | Addition with carry output |
| `TestE2E_LONGDIFF` | Basic 64-bit subtract |
| `TestE2E_LONGDIFFBorrow` | Subtraction with borrow |
| `TestE2E_NORMALISE` | Leading-zero normalization |
| `TestE2E_NORMALISEZero` | Zero input edge case |
| `TestE2E_SHIFTRIGHT` | 64-bit right shift |
| `TestE2E_SHIFTLEFT` | 64-bit left shift |
| `TestE2E_SHIFTLEFTCrossWord` | Shift across word boundary |

#### `codegen/e2e_retypes_test.go` — 6 tests

| Test | Feature |
|------|---------|
| `TestE2E_RetypesFloat32ToInt` | `VAL INT bits RETYPES x :` (float32 1.0) |
| `TestE2E_RetypesFloat32Zero` | float32 0.0 bit pattern |
| `TestE2E_RetypesFloat32NegOne` | float32 -1.0 bit pattern |
| `TestE2E_RetypesSameNameShadow` | `VAL INT X RETYPES X :` (param rename) |
| `TestE2E_RetypesFloat64ToIntPair` | `VAL [2]INT X RETYPES X :` (float64 1.0) |
| `TestE2E_RetypesFloat64Zero` | float64 0.0 split into two words |

#### `codegen/e2e_params_test.go` — 6 tests

| Test | Feature |
|------|---------|
| `TestE2E_ResultQualifier` | `PROC f(RESULT INT x)` |
| `TestE2E_ResultQualifierMultiple` | Multiple RESULT params |
| `TestE2E_FixedSizeArrayParam` | `PROC f([2]INT arr)` → pointer |
| `TestE2E_SharedTypeChanParams` | `PROC f(CHAN OF INT input?, output!)` |
| `TestE2E_SharedTypeIntParams` | `PROC f(VAL INT a, b, INT result)` |
| `TestE2E_ValOpenArrayByteParam` | `PROC f(VAL []BYTE s)` with string arg |

#### `codegen/e2e_strings_test.go` — 5 tests

| Test | Feature |
|------|---------|
| `TestE2E_ValByteArrayAbbreviation` | `VAL []BYTE s IS "hello":` |
| `TestE2E_PrintString` | `print.string("hello world")` |
| `TestE2E_PrintNewline` | `print.newline()` |
| `TestE2E_PrintStringAndNewline` | Combined string printing |
| `TestE2E_StringWithEscapes` | Occam `*t` escape in string |

#### `codegen/e2e_misc_test.go` — 24 tests

| Test | Feature |
|------|---------|
| `TestE2E_SkipStatement` | SKIP as standalone no-op |
| `TestE2E_SkipInPar` | SKIP in a PAR branch |
| `TestE2E_StopReached` | STOP causes non-zero exit (deadlock) |
| `TestE2E_ModuloOperator` | `\` → `%` |
| `TestE2E_ModuloInExpression` | Modulo in compound expression |
| `TestE2E_AltWithBooleanGuard` | FALSE guard disables ALT branch |
| `TestE2E_AltWithTrueGuard` | TRUE guard enables ALT branch |
| `TestE2E_MostNegReal32` | `MOSTNEG REAL32` is negative |
| `TestE2E_MostPosReal32` | `MOSTPOS REAL32` is positive |
| `TestE2E_MostNegReal64` | `MOSTNEG REAL64` is negative |
| `TestE2E_MostPosReal64` | `MOSTPOS REAL64` is positive |
| `TestE2E_ShorthandSliceFromZero` | `[arr FOR 3]` (FROM 0 implied) |
| `TestE2E_StringToByteSliceWrapping` | String literal → `[]byte()` for `[]BYTE` param |
| `TestE2E_GoReservedWordEscaping` | Variable named `len` works |
| `TestE2E_GoReservedWordByte` | Variable named `byte` works |
| `TestE2E_MultiLineExpression` | Continuation operator at line end |
| `TestE2E_MultiLineParenExpression` | Expression inside parens across lines |
| `TestE2E_NegativeIntLiteral` | Unary minus |
| `TestE2E_NotOperator` | `NOT TRUE` |
| `TestE2E_LogicalAndOr` | `AND` / `OR` operators |
| `TestE2E_NestedIfInSeq` | Nested IF with variable declarations |
| `TestE2E_WhileWithBreakCondition` | WHILE counting to target |
| `TestE2E_CaseWithMultipleArms` | CASE with 4 branches |
| `TestE2E_EqualNotEqual` | `=` and `<>` operators |
| `TestE2E_CompileOnly_StopInProc` | STOP in proc compiles cleanly |
| `TestE2E_NestedReplicatedSeq` | Nested `SEQ i = 0 FOR 3` loops |
| `TestE2E_ArraySliceAssignment` | `[dst FROM 1 FOR 3] := src` |
| `TestE2E_FunctionCallInCondition` | `BOOL FUNCTION` as IF condition |
| `TestE2E_RecursiveFunction` | Recursive factorial |
| `TestE2E_MultiLineProcParams` | Multi-line proc parameter list |
| `TestE2E_VetOutputClean` | `go vet` passes on generated code |

#### `lexer/lexer_test2_test.go` — 15 tests

| Test | Feature |
|------|---------|
| `TestAllKeywords` | All 33+ keywords tokenize correctly |
| `TestParenDepthSuppressesIndent` | No INDENT/DEDENT inside `(...)` |
| `TestBracketDepthSuppressesIndent` | No INDENT/DEDENT inside `[...]` |
| `TestContinuationOperator` | `+` at line end joins lines |
| `TestContinuationAND` | `AND` at line end joins lines |
| `TestStringLiteral` | `"hello world"` → STRING token |
| `TestStringEscapeSequences` | `*n` preserved raw by lexer |
| `TestByteLiteralToken` | `'A'` → BYTE_LIT token |
| `TestByteLiteralEscapeToken` | `'*n'` → BYTE_LIT with raw escape |
| `TestSendReceiveTokens` | `!` → SEND, `?` → RECEIVE |
| `TestAmpersandToken` | `&` → AMPERSAND |
| `TestSemicolonToken` | `;` → SEMICOLON |
| `TestNestedParenDepth` | Nested `((` tracks depth correctly |
| `TestMixedParenBracketDepth` | `arr[(1 + 2)]` mixed nesting |
| `TestLineAndColumnTracking` | Token line/column numbers |

### Summary

| Area | Before | Added | After |
|------|--------|-------|-------|
| Lexer unit tests | 9 | 15 | 24 |
| Parser unit tests | 80 | 0 | 80 |
| Codegen unit tests | 49 | 0 | 49 |
| E2E tests | ~124 | 54 | ~178 |
| Preprocessor tests | 22 | 0 | 22 |
| Modgen tests | 5 | 0 | 5 |
| **Total** | **~289** | **69** | **~358** |

### Remaining Gaps

Features with limited or indirect-only coverage that could benefit from future tests:

- Parser-level tests for SKIP, STOP, `VAL []BYTE` abbreviation, RESULT qualifier, fixed-size array params, transputer intrinsic calls, variant receive `c ? CASE`, timer ALT arm
- Codegen unit tests for RETYPES output, intrinsic helper emission, sequential/variant protocol send/receive code, `VAL []BYTE` abbreviation output
- `print.nl` (alias for `print.newline`, if supported)
- PRI ALT / PRI PAR (not yet implemented)
