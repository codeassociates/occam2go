package codegen

import (
	"fmt"
	"strings"

	"github.com/codeassociates/occam2go/ast"
)

// Generator converts an AST to Go code
type Generator struct {
	indent   int
	builder  strings.Builder
	needSync bool // track if we need sync package import
	needFmt  bool // track if we need fmt package import
	needTime bool // track if we need time package import
	needOs   bool // track if we need os package import
	needMath bool // track if we need math package import
	needMathBits bool // track if we need math/bits package import
	needBufio    bool // track if we need bufio package import
	needReflect    bool // track if we need reflect package import
	needBoolHelper bool // track if we need _boolToInt helper

	// Track procedure signatures for proper pointer handling
	procSigs map[string][]ast.ProcParam
	// Track current procedure's reference parameters
	refParams map[string]bool

	// Protocol support
	protocolDefs  map[string]*ast.ProtocolDecl
	chanProtocols map[string]string // channel name → protocol name
	tmpCounter    int               // for unique temp variable names

	// Record support
	recordDefs map[string]*ast.RecordDecl
	recordVars map[string]string // variable name → record type name

	// Channel element type tracking (for ALT guard codegen)
	chanElemTypes map[string]string // channel name → Go element type

	// Bool variable tracking (for type conversion codegen)
	boolVars map[string]bool

	// Nesting level: 0 = package level, >0 = inside a function
	nestingLevel int

	// RETYPES parameter renames: when a RETYPES declaration shadows a
	// parameter (e.g. VAL INT X RETYPES X :), the parameter is renamed
	// in the signature so := can create a new variable with the original name.
	retypesRenames map[string]string
}

// Transputer intrinsic function names
var transpIntrinsics = map[string]bool{
	"LONGPROD":   true,
	"LONGDIV":    true,
	"LONGSUM":    true,
	"LONGDIFF":   true,
	"NORMALISE":  true,
	"SHIFTRIGHT": true,
	"SHIFTLEFT":  true,
}

// Built-in print procedures
var printBuiltins = map[string]bool{
	"print.int":     true,
	"print.string":  true,
	"print.bool":    true,
	"print.newline": true,
}

// New creates a new code generator
func New() *Generator {
	return &Generator{}
}

// goIdent converts an occam identifier to a valid Go identifier.
// Occam allows dots in identifiers (e.g., out.repeat); Go does not.
// goReserved is a set of Go keywords and predeclared identifiers that cannot be
// used as variable names when they also appear as type conversions in the generated code.
var goReserved = map[string]bool{
	"byte": true, "int": true, "string": true, "len": true, "cap": true,
	"make": true, "new": true, "copy": true, "close": true, "delete": true,
	"panic": true, "recover": true, "print": true, "println": true,
	"error": true, "rune": true, "bool": true, "true": true, "false": true,
}

func goIdent(name string) string {
	name = strings.ReplaceAll(name, ".", "_")
	if goReserved[name] {
		return "_" + name
	}
	return name
}

// Generate produces Go code from the AST
func (g *Generator) Generate(program *ast.Program) string {
	g.builder.Reset()
	g.needSync = false
	g.needFmt = false
	g.needTime = false
	g.needOs = false
	g.needMath = false
	g.needMathBits = false
	g.needBufio = false
	g.needReflect = false
	g.needBoolHelper = false
	g.procSigs = make(map[string][]ast.ProcParam)
	g.refParams = make(map[string]bool)
	g.protocolDefs = make(map[string]*ast.ProtocolDecl)
	g.chanProtocols = make(map[string]string)
	g.chanElemTypes = make(map[string]string)
	g.tmpCounter = 0
	g.recordDefs = make(map[string]*ast.RecordDecl)
	g.recordVars = make(map[string]string)
	g.boolVars = make(map[string]bool)

	// Pre-pass: collect BOOL variable names (needed before containsBoolConversion)
	for _, stmt := range program.Statements {
		g.collectBoolVars(stmt)
	}

	// First pass: collect procedure signatures, protocols, and check for PAR/print
	for _, stmt := range program.Statements {
		if g.containsPar(stmt) {
			g.needSync = true
		}
		if g.containsPrint(stmt) {
			g.needFmt = true
		}
		if g.containsTimer(stmt) {
			g.needTime = true
		}
		if g.containsStop(stmt) {
			g.needOs = true
			g.needFmt = true
		}
		if g.containsMostExpr(stmt) {
			g.needMath = true
		}
		if g.containsIntrinsics(stmt) {
			g.needMathBits = true
		}
		if g.containsRetypes(stmt) {
			g.needMath = true
		}
		if g.containsAltReplicator(stmt) {
			g.needReflect = true
		}
		if g.containsBoolConversion(stmt) {
			g.needBoolHelper = true
		}
		if proc, ok := stmt.(*ast.ProcDecl); ok {
			g.procSigs[proc.Name] = proc.Params
			g.collectNestedProcSigs(proc.Body)
		}
		if fn, ok := stmt.(*ast.FuncDecl); ok {
			g.procSigs[fn.Name] = fn.Params
		}
		if proto, ok := stmt.(*ast.ProtocolDecl); ok {
			g.protocolDefs[proto.Name] = proto
		}
		if rec, ok := stmt.(*ast.RecordDecl); ok {
			g.recordDefs[rec.Name] = rec
		}
		g.collectChanProtocols(stmt)
		g.collectRecordVars(stmt)
	}

	// Separate protocol, record, procedure declarations from other statements
	var typeDecls []ast.Statement
	var procDecls []ast.Statement
	var mainStatements []ast.Statement

	// First pass: check if there are any proc/func declarations
	hasProcDecls := false
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.ProcDecl); ok {
			hasProcDecls = true
			break
		}
		if _, ok := stmt.(*ast.FuncDecl); ok {
			hasProcDecls = true
			break
		}
	}

	var abbrDecls []ast.Statement
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.ProtocolDecl, *ast.RecordDecl:
			typeDecls = append(typeDecls, stmt)
		case *ast.ProcDecl, *ast.FuncDecl:
			procDecls = append(procDecls, stmt)
		case *ast.Abbreviation:
			if hasProcDecls {
				// Top-level abbreviations need to be at package level
				// so PROCs can reference them
				abbrDecls = append(abbrDecls, stmt)
			} else {
				mainStatements = append(mainStatements, stmt)
			}
		case *ast.RetypesDecl:
			_ = s
			// RETYPES declarations are local to functions, not package-level
			mainStatements = append(mainStatements, stmt)
		default:
			mainStatements = append(mainStatements, stmt)
		}
	}

	// Detect entry point PROC so we can set import flags before writing imports
	var entryProc *ast.ProcDecl
	if len(mainStatements) == 0 {
		entryProc = g.findEntryProc(procDecls)
		if entryProc != nil {
			g.needOs = true
			g.needSync = true
			g.needBufio = true
		}
	}

	// Write package declaration
	g.writeLine("package main")
	g.writeLine("")

	// Write imports
	if g.needSync || g.needFmt || g.needTime || g.needOs || g.needMath || g.needMathBits || g.needBufio || g.needReflect {
		g.writeLine("import (")
		g.indent++
		if g.needBufio {
			g.writeLine(`"bufio"`)
		}
		if g.needFmt {
			g.writeLine(`"fmt"`)
		}
		if g.needMath {
			g.writeLine(`"math"`)
		}
		if g.needMathBits {
			g.writeLine(`"math/bits"`)
		}
		if g.needOs {
			g.writeLine(`"os"`)
		}
		if g.needReflect {
			g.writeLine(`"reflect"`)
		}
		if g.needSync {
			g.writeLine(`"sync"`)
		}
		if g.needTime {
			g.writeLine(`"time"`)
		}
		g.indent--
		g.writeLine(")")
		g.writeLine("")
	}

	// Emit transputer intrinsic helper functions
	if g.needMathBits {
		g.emitIntrinsicHelpers()
	}

	// Emit _boolToInt helper function
	if g.needBoolHelper {
		g.emitBoolHelper()
	}

	// Generate type definitions first (at package level)
	for _, stmt := range typeDecls {
		g.generateStatement(stmt)
	}

	// Generate package-level abbreviations (constants)
	for _, stmt := range abbrDecls {
		abbr := stmt.(*ast.Abbreviation)
		if abbr.Type == "" {
			// Untyped VAL: let Go infer the type
			g.builder.WriteString("var ")
			g.write(fmt.Sprintf("%s = ", goIdent(abbr.Name)))
			g.generateExpression(abbr.Value)
			g.write("\n")
		} else {
			goType := g.occamTypeToGo(abbr.Type)
			if abbr.IsOpenArray || abbr.IsFixedArray {
				goType = "[]" + goType
			}
			g.builder.WriteString("var ")
			g.write(fmt.Sprintf("%s %s = ", goIdent(abbr.Name), goType))
			// Wrap string literals with []byte() when assigned to []byte variables
			if _, isStr := abbr.Value.(*ast.StringLiteral); isStr && abbr.IsOpenArray && abbr.Type == "BYTE" {
				g.write("[]byte(")
				g.generateExpression(abbr.Value)
				g.write(")")
			} else {
				g.generateExpression(abbr.Value)
			}
			g.write("\n")
		}
	}
	if len(abbrDecls) > 0 {
		g.writeLine("")
	}

	// Generate procedure declarations (at package level)
	for _, stmt := range procDecls {
		g.generateStatement(stmt)
	}

	// Generate main function with other statements
	if len(mainStatements) > 0 {
		g.writeLine("func main() {")
		g.indent++
		g.nestingLevel++
		for _, stmt := range mainStatements {
			g.generateStatement(stmt)
		}
		g.nestingLevel--
		g.indent--
		g.writeLine("}")
	} else if entryProc != nil {
		g.generateEntryHarness(entryProc)
	}

	return g.builder.String()
}

// collectNestedProcSigs recursively collects procedure/function signatures
// from nested declarations inside PROC bodies.
func (g *Generator) collectNestedProcSigs(stmts []ast.Statement) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ProcDecl:
			g.procSigs[s.Name] = s.Params
			g.collectNestedProcSigs(s.Body)
		case *ast.FuncDecl:
			g.procSigs[s.Name] = s.Params
			g.collectNestedProcSigs(s.Body)
		case *ast.SeqBlock:
			g.collectNestedProcSigs(s.Statements)
		case *ast.ParBlock:
			g.collectNestedProcSigs(s.Statements)
		case *ast.IfStatement:
			for _, c := range s.Choices {
				g.collectNestedProcSigs(c.Body)
			}
		case *ast.WhileLoop:
			g.collectNestedProcSigs(s.Body)
		case *ast.CaseStatement:
			for _, ch := range s.Choices {
				g.collectNestedProcSigs(ch.Body)
			}
		}
	}
}

// collectNestedProcSigsScoped registers nested proc/func signatures into procSigs
// for the current scope. It saves old values into oldSigs so they can be restored
// after the scope ends (preventing name collisions between same-named nested procs
// in different parent procs).
func (g *Generator) collectNestedProcSigsScoped(stmts []ast.Statement, oldSigs map[string][]ast.ProcParam) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ProcDecl:
			if _, saved := oldSigs[s.Name]; !saved {
				oldSigs[s.Name] = g.procSigs[s.Name] // nil if not previously set
			}
			g.procSigs[s.Name] = s.Params
			g.collectNestedProcSigsScoped(s.Body, oldSigs)
		case *ast.FuncDecl:
			if _, saved := oldSigs[s.Name]; !saved {
				oldSigs[s.Name] = g.procSigs[s.Name]
			}
			g.procSigs[s.Name] = s.Params
			g.collectNestedProcSigsScoped(s.Body, oldSigs)
		case *ast.SeqBlock:
			g.collectNestedProcSigsScoped(s.Statements, oldSigs)
		case *ast.ParBlock:
			g.collectNestedProcSigsScoped(s.Statements, oldSigs)
		case *ast.IfStatement:
			for _, c := range s.Choices {
				g.collectNestedProcSigsScoped(c.Body, oldSigs)
			}
		case *ast.WhileLoop:
			g.collectNestedProcSigsScoped(s.Body, oldSigs)
		case *ast.CaseStatement:
			for _, ch := range s.Choices {
				g.collectNestedProcSigsScoped(ch.Body, oldSigs)
			}
		}
	}
}

// findEntryProc looks for the last top-level PROC with the standard occam
// entry point signature: exactly 3 CHAN OF BYTE params (keyboard?, screen!, error!).
func (g *Generator) findEntryProc(procDecls []ast.Statement) *ast.ProcDecl {
	var entry *ast.ProcDecl
	for _, stmt := range procDecls {
		proc, ok := stmt.(*ast.ProcDecl)
		if !ok {
			continue
		}
		if len(proc.Params) != 3 {
			continue
		}
		p0, p1, p2 := proc.Params[0], proc.Params[1], proc.Params[2]
		if p0.IsChan && p0.ChanElemType == "BYTE" && p0.ChanDir == "?" &&
			p1.IsChan && p1.ChanElemType == "BYTE" && p1.ChanDir == "!" &&
			p2.IsChan && p2.ChanElemType == "BYTE" && p2.ChanDir == "!" {
			entry = proc
		}
	}
	return entry
}

// generateEntryHarness emits a func main() that wires stdin/stdout/stderr
// to channels and calls the entry PROC.
func (g *Generator) generateEntryHarness(proc *ast.ProcDecl) {
	name := goIdent(proc.Name)
	g.writeLine("func main() {")
	g.indent++

	// Create channels
	g.writeLine("keyboard := make(chan byte, 256)")
	g.writeLine("screen := make(chan byte, 256)")
	g.writeLine("_error := make(chan byte, 256)")
	g.writeLine("")

	// WaitGroup for writer goroutines to finish draining
	g.writeLine("var wg sync.WaitGroup")
	g.writeLine("wg.Add(2)")
	g.writeLine("")

	// Screen writer goroutine
	g.writeLine("go func() {")
	g.indent++
	g.writeLine("defer wg.Done()")
	g.writeLine("w := bufio.NewWriter(os.Stdout)")
	g.writeLine("for b := range screen {")
	g.indent++
	g.writeLine("if b == 255 {")
	g.indent++
	g.writeLine("w.Flush()")
	g.indent--
	g.writeLine("} else {")
	g.indent++
	g.writeLine("w.WriteByte(b)")
	g.indent--
	g.writeLine("}")
	g.indent--
	g.writeLine("}")
	g.writeLine("w.Flush()")
	g.indent--
	g.writeLine("}()")
	g.writeLine("")

	// Error writer goroutine
	g.writeLine("go func() {")
	g.indent++
	g.writeLine("defer wg.Done()")
	g.writeLine("w := bufio.NewWriter(os.Stderr)")
	g.writeLine("for b := range _error {")
	g.indent++
	g.writeLine("if b == 255 {")
	g.indent++
	g.writeLine("w.Flush()")
	g.indent--
	g.writeLine("} else {")
	g.indent++
	g.writeLine("w.WriteByte(b)")
	g.indent--
	g.writeLine("}")
	g.indent--
	g.writeLine("}")
	g.writeLine("w.Flush()")
	g.indent--
	g.writeLine("}()")
	g.writeLine("")

	// Keyboard reader goroutine
	g.writeLine("go func() {")
	g.indent++
	g.writeLine("r := bufio.NewReader(os.Stdin)")
	g.writeLine("for {")
	g.indent++
	g.writeLine("b, err := r.ReadByte()")
	g.writeLine("if err != nil {")
	g.indent++
	g.writeLine("close(keyboard)")
	g.writeLine("return")
	g.indent--
	g.writeLine("}")
	g.writeLine("keyboard <- b")
	g.indent--
	g.writeLine("}")
	g.indent--
	g.writeLine("}()")
	g.writeLine("")

	// Call the entry proc
	g.writeLine(fmt.Sprintf("%s(keyboard, screen, _error)", name))
	g.writeLine("")

	// Close output channels and wait for writers to drain
	g.writeLine("close(screen)")
	g.writeLine("close(_error)")
	g.writeLine("wg.Wait()")

	g.indent--
	g.writeLine("}")
}

func (g *Generator) containsPar(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.ParBlock:
		return true
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			if g.containsPar(inner) {
				return true
			}
		}
	case *ast.AltBlock:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsPar(inner) {
					return true
				}
			}
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			if g.containsPar(inner) {
				return true
			}
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			if g.containsPar(inner) {
				return true
			}
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			if g.containsPar(inner) {
				return true
			}
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil {
				if g.containsPar(choice.NestedIf) {
					return true
				}
			}
			for _, inner := range choice.Body {
				if g.containsPar(inner) {
					return true
				}
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				if g.containsPar(inner) {
					return true
				}
			}
		}
	case *ast.VariantReceive:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsPar(inner) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Generator) containsPrint(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.ProcCall:
		return printBuiltins[s.Name]
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			if g.containsPrint(inner) {
				return true
			}
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			if g.containsPrint(inner) {
				return true
			}
		}
	case *ast.AltBlock:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsPrint(inner) {
					return true
				}
			}
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			if g.containsPrint(inner) {
				return true
			}
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			if g.containsPrint(inner) {
				return true
			}
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			if g.containsPrint(inner) {
				return true
			}
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil {
				if g.containsPrint(choice.NestedIf) {
					return true
				}
			}
			for _, inner := range choice.Body {
				if g.containsPrint(inner) {
					return true
				}
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				if g.containsPrint(inner) {
					return true
				}
			}
		}
	case *ast.VariantReceive:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsPrint(inner) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Generator) containsTimer(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.TimerDecl, *ast.TimerRead:
		return true
	case *ast.AltBlock:
		for _, c := range s.Cases {
			if c.IsTimer {
				return true
			}
			for _, inner := range c.Body {
				if g.containsTimer(inner) {
					return true
				}
			}
		}
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			if g.containsTimer(inner) {
				return true
			}
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			if g.containsTimer(inner) {
				return true
			}
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			if g.containsTimer(inner) {
				return true
			}
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			if g.containsTimer(inner) {
				return true
			}
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			if g.containsTimer(inner) {
				return true
			}
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil {
				if g.containsTimer(choice.NestedIf) {
					return true
				}
			}
			for _, inner := range choice.Body {
				if g.containsTimer(inner) {
					return true
				}
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				if g.containsTimer(inner) {
					return true
				}
			}
		}
	case *ast.VariantReceive:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsTimer(inner) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Generator) containsStop(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.Stop:
		return true
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			if g.containsStop(inner) {
				return true
			}
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			if g.containsStop(inner) {
				return true
			}
		}
	case *ast.AltBlock:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsStop(inner) {
					return true
				}
			}
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			if g.containsStop(inner) {
				return true
			}
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			if g.containsStop(inner) {
				return true
			}
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			if g.containsStop(inner) {
				return true
			}
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil {
				if g.containsStop(choice.NestedIf) {
					return true
				}
			}
			for _, inner := range choice.Body {
				if g.containsStop(inner) {
					return true
				}
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				if g.containsStop(inner) {
					return true
				}
			}
		}
	case *ast.VariantReceive:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsStop(inner) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Generator) containsMostExpr(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.Assignment:
		result := g.exprNeedsMath(s.Value)
		for _, idx := range s.Indices {
			result = result || g.exprNeedsMath(idx)
		}
		return result
	case *ast.MultiAssignment:
		for _, t := range s.Targets {
			for _, idx := range t.Indices {
				if g.exprNeedsMath(idx) {
					return true
				}
			}
		}
		for _, v := range s.Values {
			if g.exprNeedsMath(v) {
				return true
			}
		}
	case *ast.Abbreviation:
		return g.exprNeedsMath(s.Value)
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			if g.containsMostExpr(inner) {
				return true
			}
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			if g.containsMostExpr(inner) {
				return true
			}
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			if g.containsMostExpr(inner) {
				return true
			}
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			if g.containsMostExpr(inner) {
				return true
			}
		}
	case *ast.WhileLoop:
		if g.exprNeedsMath(s.Condition) {
			return true
		}
		for _, inner := range s.Body {
			if g.containsMostExpr(inner) {
				return true
			}
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil {
				if g.containsMostExpr(choice.NestedIf) {
					return true
				}
			}
			if g.exprNeedsMath(choice.Condition) {
				return true
			}
			for _, inner := range choice.Body {
				if g.containsMostExpr(inner) {
					return true
				}
			}
		}
	case *ast.CaseStatement:
		if g.exprNeedsMath(s.Selector) {
			return true
		}
		for _, choice := range s.Choices {
			for _, v := range choice.Values {
				if g.exprNeedsMath(v) {
					return true
				}
			}
			for _, inner := range choice.Body {
				if g.containsMostExpr(inner) {
					return true
				}
			}
		}
	case *ast.Send:
		if g.exprNeedsMath(s.Value) {
			return true
		}
		for _, v := range s.Values {
			if g.exprNeedsMath(v) {
				return true
			}
		}
	case *ast.ProcCall:
		for _, arg := range s.Args {
			if g.exprNeedsMath(arg) {
				return true
			}
		}
	case *ast.AltBlock:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsMostExpr(inner) {
					return true
				}
			}
		}
	case *ast.VariantReceive:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsMostExpr(inner) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Generator) exprNeedsMath(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.MostExpr:
		// BYTE uses literal 0/255, doesn't need math
		return e.ExprType != "BYTE"
	case *ast.BinaryExpr:
		return g.exprNeedsMath(e.Left) || g.exprNeedsMath(e.Right)
	case *ast.UnaryExpr:
		return g.exprNeedsMath(e.Right)
	case *ast.ParenExpr:
		return g.exprNeedsMath(e.Expr)
	case *ast.TypeConversion:
		if e.Qualifier == "ROUND" && isOccamIntType(e.TargetType) {
			return true
		}
		return g.exprNeedsMath(e.Expr)
	case *ast.SizeExpr:
		return g.exprNeedsMath(e.Expr)
	case *ast.IndexExpr:
		return g.exprNeedsMath(e.Left) || g.exprNeedsMath(e.Index)
	case *ast.FuncCall:
		for _, arg := range e.Args {
			if g.exprNeedsMath(arg) {
				return true
			}
		}
	case *ast.SliceExpr:
		return g.exprNeedsMath(e.Array) || g.exprNeedsMath(e.Start) || g.exprNeedsMath(e.Length)
	case *ast.ArrayLiteral:
		for _, elem := range e.Elements {
			if g.exprNeedsMath(elem) {
				return true
			}
		}
	}
	return false
}

func (g *Generator) generateMostExpr(e *ast.MostExpr) {
	switch e.ExprType {
	case "INT":
		if e.IsNeg {
			g.write("math.MinInt")
		} else {
			g.write("math.MaxInt")
		}
	case "INT16":
		if e.IsNeg {
			g.write("math.MinInt16")
		} else {
			g.write("math.MaxInt16")
		}
	case "INT32":
		if e.IsNeg {
			g.write("math.MinInt32")
		} else {
			g.write("math.MaxInt32")
		}
	case "INT64":
		if e.IsNeg {
			g.write("math.MinInt64")
		} else {
			g.write("math.MaxInt64")
		}
	case "BYTE":
		if e.IsNeg {
			g.write("0")
		} else {
			g.write("255")
		}
	case "REAL32":
		if e.IsNeg {
			g.write("-math.MaxFloat32")
		} else {
			g.write("math.MaxFloat32")
		}
	case "REAL64":
		if e.IsNeg {
			g.write("-math.MaxFloat64")
		} else {
			g.write("math.MaxFloat64")
		}
	}
}

func (g *Generator) writeLine(s string) {
	if s == "" {
		g.builder.WriteString("\n")
		return
	}
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.builder.WriteString(s)
	g.builder.WriteString("\n")
}

func (g *Generator) write(s string) {
	g.builder.WriteString(s)
}

func (g *Generator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		g.generateVarDecl(s)
	case *ast.ArrayDecl:
		g.generateArrayDecl(s)
	case *ast.ChanDecl:
		g.generateChanDecl(s)
	case *ast.Assignment:
		g.generateAssignment(s)
	case *ast.Send:
		g.generateSend(s)
	case *ast.Receive:
		g.generateReceive(s)
	case *ast.SeqBlock:
		g.generateSeqBlock(s)
	case *ast.ParBlock:
		g.generateParBlock(s)
	case *ast.AltBlock:
		g.generateAltBlock(s)
	case *ast.Skip:
		g.writeLine("// SKIP")
	case *ast.Stop:
		g.writeLine(`fmt.Fprintln(os.Stderr, "STOP encountered")`)
		g.writeLine("select {}")
	case *ast.ProcDecl:
		g.generateProcDecl(s)
	case *ast.FuncDecl:
		g.generateFuncDecl(s)
	case *ast.ProcCall:
		g.generateProcCall(s)
	case *ast.WhileLoop:
		g.generateWhileLoop(s)
	case *ast.IfStatement:
		g.generateIfStatement(s)
	case *ast.CaseStatement:
		g.generateCaseStatement(s)
	case *ast.TimerDecl:
		g.generateTimerDecl(s)
	case *ast.TimerRead:
		g.generateTimerRead(s)
	case *ast.ProtocolDecl:
		g.generateProtocolDecl(s)
	case *ast.VariantReceive:
		g.generateVariantReceive(s)
	case *ast.RecordDecl:
		g.generateRecordDecl(s)
	case *ast.Abbreviation:
		g.generateAbbreviation(s)
	case *ast.MultiAssignment:
		g.generateMultiAssignment(s)
	case *ast.RetypesDecl:
		g.generateRetypesDecl(s)
	}
}

func (g *Generator) generateVarDecl(decl *ast.VarDecl) {
	goType := g.occamTypeToGo(decl.Type)
	goNames := make([]string, len(decl.Names))
	for i, n := range decl.Names {
		goNames[i] = goIdent(n)
	}
	g.writeLine(fmt.Sprintf("var %s %s", strings.Join(goNames, ", "), goType))
	// Suppress "declared and not used" for each variable
	for _, n := range goNames {
		g.writeLine(fmt.Sprintf("_ = %s", n))
	}
	// Track BOOL variables for type conversion codegen
	if decl.Type == "BOOL" {
		for _, n := range decl.Names {
			g.boolVars[n] = true
		}
	}
}

func (g *Generator) generateAbbreviation(abbr *ast.Abbreviation) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	if abbr.Type != "" {
		goType := g.occamTypeToGo(abbr.Type)
		if abbr.IsOpenArray || abbr.IsFixedArray {
			goType = "[]" + goType
		}
		g.write(fmt.Sprintf("var %s %s = ", goIdent(abbr.Name), goType))
	} else {
		g.write(fmt.Sprintf("%s := ", goIdent(abbr.Name)))
	}
	// Wrap string literals with []byte() when assigned to []byte variables
	if _, isStr := abbr.Value.(*ast.StringLiteral); isStr && abbr.IsOpenArray && abbr.Type == "BYTE" {
		g.write("[]byte(")
		g.generateExpression(abbr.Value)
		g.write(")")
	} else {
		g.generateExpression(abbr.Value)
	}
	g.write("\n")
	// Suppress "declared and not used" for abbreviations inside function bodies
	if g.nestingLevel > 0 {
		g.writeLine(fmt.Sprintf("_ = %s", goIdent(abbr.Name)))
	}
}

func (g *Generator) generateChanDecl(decl *ast.ChanDecl) {
	goType := g.occamTypeToGo(decl.ElemType)
	for _, name := range decl.Names {
		g.chanElemTypes[name] = goType
	}
	if len(decl.Sizes) > 0 {
		for _, name := range decl.Names {
			n := goIdent(name)
			g.generateMultiDimChanInit(n, goType, decl.Sizes, 0)
		}
	} else {
		for _, name := range decl.Names {
			g.writeLine(fmt.Sprintf("%s := make(chan %s)", goIdent(name), goType))
		}
	}
}

// generateMultiDimChanInit generates nested make+init loops for multi-dimensional channel arrays.
// For [w][h]CHAN OF INT link: generates:
//
//	link := make([][]chan int, w)
//	for _i0 := range link { link[_i0] = make([]chan int, h)
//	    for _i1 := range link[_i0] { link[_i0][_i1] = make(chan int) }
//	}
func (g *Generator) generateMultiDimChanInit(name, goType string, sizes []ast.Expression, depth int) {
	if depth == 0 {
		// Top-level: name := make([]...[]chan goType, sizes[0])
		sliceType := strings.Repeat("[]", len(sizes)) + "chan " + goType
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("%s := make(%s, ", name, sliceType))
		g.generateExpression(sizes[0])
		g.write(")\n")
		if len(sizes) == 1 {
			// Single dim: init each channel
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("for _i0 := range %s { %s[_i0] = make(chan %s) }\n", name, name, goType))
		} else {
			// Multi dim: recurse
			ivar := "_i0"
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("for %s := range %s {\n", ivar, name))
			g.indent++
			g.generateMultiDimChanInit(name+"["+ivar+"]", goType, sizes, 1)
			g.indent--
			g.writeLine("}")
		}
	} else if depth < len(sizes)-1 {
		// Middle dimension: allocate sub-slice
		sliceType := strings.Repeat("[]", len(sizes)-depth) + "chan " + goType
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("%s = make(%s, ", name, sliceType))
		g.generateExpression(sizes[depth])
		g.write(")\n")
		ivar := fmt.Sprintf("_i%d", depth)
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("for %s := range %s {\n", ivar, name))
		g.indent++
		g.generateMultiDimChanInit(name+"["+ivar+"]", goType, sizes, depth+1)
		g.indent--
		g.writeLine("}")
	} else {
		// Innermost dimension: allocate sub-slice + init channels
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("%s = make([]chan %s, ", name, goType))
		g.generateExpression(sizes[depth])
		g.write(")\n")
		ivar := fmt.Sprintf("_i%d", depth)
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("for %s := range %s { %s[%s] = make(chan %s) }\n", ivar, name, name, ivar, goType))
	}
}

func (g *Generator) generateTimerDecl(decl *ast.TimerDecl) {
	for _, name := range decl.Names {
		g.writeLine(fmt.Sprintf("// TIMER %s", name))
	}
}

func (g *Generator) generateTimerRead(tr *ast.TimerRead) {
	g.writeLine(fmt.Sprintf("%s = int(time.Now().UnixMicro())", goIdent(tr.Variable)))
}

func (g *Generator) generateArrayDecl(decl *ast.ArrayDecl) {
	goType := g.occamTypeToGo(decl.Type)
	for _, name := range decl.Names {
		n := goIdent(name)
		if len(decl.Sizes) == 1 {
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("%s := make([]%s, ", n, goType))
			g.generateExpression(decl.Sizes[0])
			g.write(")\n")
		} else {
			g.generateMultiDimArrayInit(n, goType, decl.Sizes, 0)
		}
	}
}

// generateMultiDimArrayInit generates nested make+init loops for multi-dimensional arrays.
// For [5][3]INT arr: generates:
//
//	arr := make([][]int, 5)
//	for _i0 := range arr { arr[_i0] = make([]int, 3) }
func (g *Generator) generateMultiDimArrayInit(name, goType string, sizes []ast.Expression, depth int) {
	if depth == 0 {
		sliceType := strings.Repeat("[]", len(sizes)) + goType
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("%s := make(%s, ", name, sliceType))
		g.generateExpression(sizes[0])
		g.write(")\n")
		if len(sizes) > 1 {
			ivar := "_i0"
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("for %s := range %s {\n", ivar, name))
			g.indent++
			g.generateMultiDimArrayInit(name+"["+ivar+"]", goType, sizes, 1)
			g.indent--
			g.writeLine("}")
		}
	} else if depth < len(sizes)-1 {
		sliceType := strings.Repeat("[]", len(sizes)-depth) + goType
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("%s = make(%s, ", name, sliceType))
		g.generateExpression(sizes[depth])
		g.write(")\n")
		ivar := fmt.Sprintf("_i%d", depth)
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("for %s := range %s {\n", ivar, name))
		g.indent++
		g.generateMultiDimArrayInit(name+"["+ivar+"]", goType, sizes, depth+1)
		g.indent--
		g.writeLine("}")
	} else {
		// Innermost dimension
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("%s = make([]%s, ", name, goType))
		g.generateExpression(sizes[depth])
		g.write(")\n")
	}
}

// generateIndices emits [idx1][idx2]... for multi-dimensional index access.
func (g *Generator) generateIndices(indices []ast.Expression) {
	for _, idx := range indices {
		g.write("[")
		g.generateExpression(idx)
		g.write("]")
	}
}

// generateIndicesStr generates indices into a buffer and returns the string.
func (g *Generator) generateIndicesStr(indices []ast.Expression) string {
	var buf strings.Builder
	for _, idx := range indices {
		buf.WriteString("[")
		oldBuilder := g.builder
		g.builder = strings.Builder{}
		g.generateExpression(idx)
		buf.WriteString(g.builder.String())
		g.builder = oldBuilder
		buf.WriteString("]")
	}
	return buf.String()
}

func (g *Generator) generateSend(send *ast.Send) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write(goIdent(send.Channel))
	g.generateIndices(send.ChannelIndices)
	g.write(" <- ")

	protoName := g.chanProtocols[send.Channel]
	proto := g.protocolDefs[protoName]
	gProtoName := goIdent(protoName)

	if send.VariantTag != "" && proto != nil && proto.Kind == "variant" {
		// Variant send with explicit tag: c <- _proto_NAME_tag{values...}
		g.write(fmt.Sprintf("_proto_%s_%s{", gProtoName, goIdent(send.VariantTag)))
		for i, val := range send.Values {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(val)
		}
		g.write("}")
	} else if proto != nil && proto.Kind == "variant" && send.Value != nil && len(send.Values) == 0 {
		// Check if the send value is a bare identifier matching a variant tag
		if ident, ok := send.Value.(*ast.Identifier); ok && g.isVariantTag(protoName, ident.Value) {
			g.write(fmt.Sprintf("_proto_%s_%s{}", gProtoName, goIdent(ident.Value)))
		} else {
			g.generateExpression(send.Value)
		}
	} else if len(send.Values) > 0 && proto != nil && proto.Kind == "sequential" {
		// Sequential send: c <- _proto_NAME{val1, val2, ...}
		g.write(fmt.Sprintf("_proto_%s{", gProtoName))
		g.generateExpression(send.Value)
		for _, val := range send.Values {
			g.write(", ")
			g.generateExpression(val)
		}
		g.write("}")
	} else {
		// Simple send
		g.generateExpression(send.Value)
	}
	g.write("\n")
}

func (g *Generator) generateReceive(recv *ast.Receive) {
	chanRef := goIdent(recv.Channel)
	if len(recv.ChannelIndices) > 0 {
		chanRef += g.generateIndicesStr(recv.ChannelIndices)
	}

	if len(recv.Variables) > 0 {
		// Sequential receive: _tmpN := <-c; x = _tmpN._0; y = _tmpN._1
		tmpName := fmt.Sprintf("_tmp%d", g.tmpCounter)
		g.tmpCounter++
		g.writeLine(fmt.Sprintf("%s := <-%s", tmpName, chanRef))
		varRef := goIdent(recv.Variable)
		if len(recv.VariableIndices) > 0 {
			varRef += g.generateIndicesStr(recv.VariableIndices)
		} else if g.refParams[recv.Variable] {
			varRef = "*" + varRef
		}
		g.writeLine(fmt.Sprintf("%s = %s._0", varRef, tmpName))
		for i, v := range recv.Variables {
			vRef := goIdent(v)
			if g.refParams[v] {
				vRef = "*" + vRef
			}
			g.writeLine(fmt.Sprintf("%s = %s._%d", vRef, tmpName, i+1))
		}
	} else {
		varRef := goIdent(recv.Variable)
		if len(recv.VariableIndices) > 0 {
			varRef += g.generateIndicesStr(recv.VariableIndices)
		} else if g.refParams[recv.Variable] {
			varRef = "*" + varRef
		}
		g.writeLine(fmt.Sprintf("%s = <-%s", varRef, chanRef))
	}
}

func (g *Generator) generateProtocolDecl(proto *ast.ProtocolDecl) {
	gName := goIdent(proto.Name)
	switch proto.Kind {
	case "simple":
		goType := g.occamTypeToGoBase(proto.Types[0])
		g.writeLine(fmt.Sprintf("type _proto_%s = %s", gName, goType))
		g.writeLine("")
	case "sequential":
		g.writeLine(fmt.Sprintf("type _proto_%s struct {", gName))
		g.indent++
		for i, t := range proto.Types {
			goType := g.occamTypeToGoBase(t)
			g.writeLine(fmt.Sprintf("_%d %s", i, goType))
		}
		g.indent--
		g.writeLine("}")
		g.writeLine("")
	case "variant":
		// Interface type
		g.writeLine(fmt.Sprintf("type _proto_%s interface {", gName))
		g.indent++
		g.writeLine(fmt.Sprintf("_is_%s()", gName))
		g.indent--
		g.writeLine("}")
		g.writeLine("")
		// Concrete types for each variant
		for _, v := range proto.Variants {
			gTag := goIdent(v.Tag)
			if len(v.Types) == 0 {
				// No-payload variant: empty struct
				g.writeLine(fmt.Sprintf("type _proto_%s_%s struct{}", gName, gTag))
			} else {
				g.writeLine(fmt.Sprintf("type _proto_%s_%s struct {", gName, gTag))
				g.indent++
				for i, t := range v.Types {
					goType := g.occamTypeToGoBase(t)
					g.writeLine(fmt.Sprintf("_%d %s", i, goType))
				}
				g.indent--
				g.writeLine("}")
			}
			g.writeLine(fmt.Sprintf("func (_proto_%s_%s) _is_%s() {}", gName, gTag, gName))
			g.writeLine("")
		}
	}
}

func (g *Generator) generateVariantReceive(vr *ast.VariantReceive) {
	protoName := g.chanProtocols[vr.Channel]
	gProtoName := goIdent(protoName)
	chanRef := goIdent(vr.Channel)
	if len(vr.ChannelIndices) > 0 {
		chanRef += g.generateIndicesStr(vr.ChannelIndices)
	}
	g.writeLine(fmt.Sprintf("switch _v := (<-%s).(type) {", chanRef))
	for _, vc := range vr.Cases {
		g.writeLine(fmt.Sprintf("case _proto_%s_%s:", gProtoName, goIdent(vc.Tag)))
		g.indent++
		for i, v := range vc.Variables {
			g.writeLine(fmt.Sprintf("%s = _v._%d", goIdent(v), i))
		}
		for _, s := range vc.Body {
			g.generateStatement(s)
		}
		g.indent--
	}
	g.writeLine("}")
}

func (g *Generator) isVariantTag(protoName, tagName string) bool {
	proto := g.protocolDefs[protoName]
	if proto == nil {
		return false
	}
	for _, v := range proto.Variants {
		if v.Tag == tagName {
			return true
		}
	}
	return false
}

func (g *Generator) collectChanProtocols(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ChanDecl:
		if _, ok := g.protocolDefs[s.ElemType]; ok {
			for _, name := range s.Names {
				g.chanProtocols[name] = s.ElemType
			}
		}
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			g.collectChanProtocols(inner)
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			g.collectChanProtocols(inner)
		}
	case *ast.ProcDecl:
		// Register PROC param channels (including channel array params)
		for _, p := range s.Params {
			if p.IsChan || p.ChanArrayDims > 0 {
				if _, ok := g.protocolDefs[p.ChanElemType]; ok {
					g.chanProtocols[p.Name] = p.ChanElemType
				}
			}
		}
		for _, inner := range s.Body {
			g.collectChanProtocols(inner)
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			g.collectChanProtocols(inner)
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			g.collectChanProtocols(inner)
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil {
				g.collectChanProtocols(choice.NestedIf)
			}
			for _, inner := range choice.Body {
				g.collectChanProtocols(inner)
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				g.collectChanProtocols(inner)
			}
		}
	case *ast.AltBlock:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				g.collectChanProtocols(inner)
			}
		}
	}
}

func (g *Generator) collectBoolVars(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		if s.Type == "BOOL" {
			for _, name := range s.Names {
				g.boolVars[name] = true
			}
		}
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			g.collectBoolVars(inner)
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			g.collectBoolVars(inner)
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			g.collectBoolVars(inner)
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			g.collectBoolVars(inner)
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			g.collectBoolVars(inner)
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil {
				g.collectBoolVars(choice.NestedIf)
			}
			for _, inner := range choice.Body {
				g.collectBoolVars(inner)
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				g.collectBoolVars(inner)
			}
		}
	}
}

func (g *Generator) collectRecordVars(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		if _, ok := g.recordDefs[s.Type]; ok {
			for _, name := range s.Names {
				g.recordVars[name] = s.Type
			}
		}
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			g.collectRecordVars(inner)
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			g.collectRecordVars(inner)
		}
	case *ast.ProcDecl:
		for _, p := range s.Params {
			if !p.IsChan {
				if _, ok := g.recordDefs[p.Type]; ok {
					g.recordVars[p.Name] = p.Type
				}
			}
		}
		for _, inner := range s.Body {
			g.collectRecordVars(inner)
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			g.collectRecordVars(inner)
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			g.collectRecordVars(inner)
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil {
				g.collectRecordVars(choice.NestedIf)
			}
			for _, inner := range choice.Body {
				g.collectRecordVars(inner)
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				g.collectRecordVars(inner)
			}
		}
	case *ast.AltBlock:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				g.collectRecordVars(inner)
			}
		}
	}
}

func (g *Generator) generateRecordDecl(rec *ast.RecordDecl) {
	g.writeLine(fmt.Sprintf("type %s struct {", goIdent(rec.Name)))
	g.indent++
	for _, f := range rec.Fields {
		goType := g.occamTypeToGoBase(f.Type)
		g.writeLine(fmt.Sprintf("%s %s", goIdent(f.Name), goType))
	}
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

// occamTypeToGoBase converts a type name without checking protocol defs
// (used inside protocol generation to avoid infinite recursion)
func (g *Generator) occamTypeToGoBase(occamType string) string {
	switch occamType {
	case "INT":
		return "int"
	case "INT16":
		return "int16"
	case "INT32":
		return "int32"
	case "INT64":
		return "int64"
	case "BYTE":
		return "byte"
	case "BOOL":
		return "bool"
	case "REAL":
		return "float64"
	case "REAL32":
		return "float32"
	case "REAL64":
		return "float64"
	default:
		return occamType
	}
}

func (g *Generator) occamTypeToGo(occamType string) string {
	switch occamType {
	case "INT":
		return "int"
	case "INT16":
		return "int16"
	case "INT32":
		return "int32"
	case "INT64":
		return "int64"
	case "BYTE":
		return "byte"
	case "BOOL":
		return "bool"
	case "REAL":
		return "float64"
	case "REAL32":
		return "float32"
	case "REAL64":
		return "float64"
	default:
		// Check if it's a protocol name
		if _, ok := g.protocolDefs[occamType]; ok {
			return "_proto_" + goIdent(occamType)
		}
		// Check if it's a record type name
		if _, ok := g.recordDefs[occamType]; ok {
			return occamType
		}
		return occamType // pass through unknown types
	}
}

func isOccamIntType(t string) bool {
	switch t {
	case "INT", "INT16", "INT32", "INT64", "BYTE":
		return true
	}
	return false
}

func (g *Generator) generateAssignment(assign *ast.Assignment) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))

	if assign.SliceTarget != nil {
		// Slice assignment: [arr FROM start FOR length] := value
		// Maps to: copy(arr[start : start + length], value)
		g.write("copy(")
		g.generateExpression(assign.SliceTarget.Array)
		g.write("[")
		g.generateExpression(assign.SliceTarget.Start)
		g.write(" : ")
		g.generateExpression(assign.SliceTarget.Start)
		g.write(" + ")
		g.generateExpression(assign.SliceTarget.Length)
		g.write("], ")
		g.generateExpression(assign.Value)
		g.write(")\n")
		return
	}

	if len(assign.Indices) > 0 {
		// Check if this is a record field access (single index that is an identifier)
		if len(assign.Indices) == 1 {
			if _, ok := g.recordVars[assign.Name]; ok {
				if ident, ok := assign.Indices[0].(*ast.Identifier); ok {
					// Record field: p.x = value (Go auto-dereferences pointers)
					g.write(goIdent(assign.Name))
					g.write(".")
					g.write(goIdent(ident.Value))
					g.write(" = ")
					g.generateExpression(assign.Value)
					g.write("\n")
					return
				}
			}
		}
		// Array index: dereference if ref param
		if g.refParams[assign.Name] {
			g.write("*")
		}
		g.write(goIdent(assign.Name))
		g.generateIndices(assign.Indices)
	} else {
		// Simple assignment: dereference if ref param
		if g.refParams[assign.Name] {
			g.write("*")
		}
		g.write(goIdent(assign.Name))
	}
	g.write(" = ")
	g.generateExpression(assign.Value)
	g.write("\n")
}

func (g *Generator) generateSeqBlock(seq *ast.SeqBlock) {
	if seq.Replicator != nil {
		if seq.Replicator.Step != nil {
			// Replicated SEQ with STEP: counter-based loop
			v := goIdent(seq.Replicator.Variable)
			counter := "_repl_" + v
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("for %s := 0; %s < ", counter, counter))
			g.generateExpression(seq.Replicator.Count)
			g.write(fmt.Sprintf("; %s++ {\n", counter))
			g.indent++
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("%s := ", v))
			g.generateExpression(seq.Replicator.Start)
			g.write(fmt.Sprintf(" + %s * ", counter))
			g.generateExpression(seq.Replicator.Step)
			g.write("\n")
		} else {
			// Replicated SEQ: SEQ i = start FOR count becomes a for loop
			v := goIdent(seq.Replicator.Variable)
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("for %s := ", v))
			g.generateExpression(seq.Replicator.Start)
			g.write(fmt.Sprintf("; %s < ", v))
			g.generateExpression(seq.Replicator.Start)
			g.write(" + ")
			g.generateExpression(seq.Replicator.Count)
			g.write(fmt.Sprintf("; %s++ {\n", v))
			g.indent++
		}
		for _, stmt := range seq.Statements {
			g.generateStatement(stmt)
		}
		g.indent--
		g.writeLine("}")
	} else {
		// SEQ just becomes sequential Go code (Go's default)
		for _, stmt := range seq.Statements {
			g.generateStatement(stmt)
		}
	}
}

func (g *Generator) generateParBlock(par *ast.ParBlock) {
	if par.Replicator != nil {
		// Replicated PAR: PAR i = start FOR count becomes goroutines in a loop
		g.writeLine("var wg sync.WaitGroup")
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write("wg.Add(int(")
		g.generateExpression(par.Replicator.Count)
		g.write("))\n")

		v := goIdent(par.Replicator.Variable)
		if par.Replicator.Step != nil {
			counter := "_repl_" + v
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("for %s := 0; %s < ", counter, counter))
			g.generateExpression(par.Replicator.Count)
			g.write(fmt.Sprintf("; %s++ {\n", counter))
			g.indent++
			// Compute loop variable from counter — also serves as closure capture
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("%s := ", v))
			g.generateExpression(par.Replicator.Start)
			g.write(fmt.Sprintf(" + %s * ", counter))
			g.generateExpression(par.Replicator.Step)
			g.write("\n")
		} else {
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("for %s := ", v))
			g.generateExpression(par.Replicator.Start)
			g.write(fmt.Sprintf("; %s < ", v))
			g.generateExpression(par.Replicator.Start)
			g.write(" + ")
			g.generateExpression(par.Replicator.Count)
			g.write(fmt.Sprintf("; %s++ {\n", v))
			g.indent++
			// Capture loop variable to avoid closure issues
			g.writeLine(fmt.Sprintf("%s := %s", v, v))
		}
		g.writeLine("go func() {")
		g.indent++
		g.writeLine("defer wg.Done()")
		for _, stmt := range par.Statements {
			g.generateStatement(stmt)
		}
		g.indent--
		g.writeLine("}()")

		g.indent--
		g.writeLine("}")
		g.writeLine("wg.Wait()")
	} else {
		// PAR becomes goroutines with WaitGroup
		g.writeLine("var wg sync.WaitGroup")
		g.writeLine(fmt.Sprintf("wg.Add(%d)", len(par.Statements)))

		for _, stmt := range par.Statements {
			g.writeLine("go func() {")
			g.indent++
			g.writeLine("defer wg.Done()")
			g.generateStatement(stmt)
			g.indent--
			g.writeLine("}()")
		}

		g.writeLine("wg.Wait()")
	}
}

func (g *Generator) generateAltBlock(alt *ast.AltBlock) {
	if alt.Replicator != nil {
		g.generateReplicatedAlt(alt)
		return
	}

	// ALT becomes Go select statement
	// For guards, we use a pattern with nil channels

	// Check if any cases have guards
	hasGuards := false
	for _, c := range alt.Cases {
		if c.Guard != nil {
			hasGuards = true
			break
		}
	}

	if hasGuards {
		// Generate channel variables for guarded cases
		for i, c := range alt.Cases {
			if c.Guard != nil && !c.IsSkip {
				g.builder.WriteString(strings.Repeat("\t", g.indent))
				// Look up the channel's element type
				elemType := "int" // default fallback
				if t, ok := g.chanElemTypes[c.Channel]; ok {
					elemType = t
				}
				g.write(fmt.Sprintf("var _alt%d <-chan %s = nil\n", i, elemType))
				g.builder.WriteString(strings.Repeat("\t", g.indent))
				g.write(fmt.Sprintf("if "))
				g.generateExpression(c.Guard)
				g.write(fmt.Sprintf(" { _alt%d = %s }\n", i, goIdent(c.Channel)))
			}
		}
	}

	g.writeLine("select {")
	for i, c := range alt.Cases {
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		if c.IsSkip {
			g.write("default:\n")
		} else if c.IsTimer {
			g.write("case <-time.After(time.Duration(")
			g.generateExpression(c.Deadline)
			g.write(" - int(time.Now().UnixMicro())) * time.Microsecond):\n")
		} else if c.Guard != nil {
			varRef := goIdent(c.Variable)
			if len(c.VariableIndices) > 0 {
				varRef += g.generateIndicesStr(c.VariableIndices)
			}
			g.write(fmt.Sprintf("case %s = <-_alt%d:\n", varRef, i))
		} else if len(c.ChannelIndices) > 0 {
			varRef := goIdent(c.Variable)
			if len(c.VariableIndices) > 0 {
				varRef += g.generateIndicesStr(c.VariableIndices)
			}
			g.write(fmt.Sprintf("case %s = <-%s", varRef, goIdent(c.Channel)))
			g.generateIndices(c.ChannelIndices)
			g.write(":\n")
		} else {
			varRef := goIdent(c.Variable)
			if len(c.VariableIndices) > 0 {
				varRef += g.generateIndicesStr(c.VariableIndices)
			}
			g.write(fmt.Sprintf("case %s = <-%s:\n", varRef, goIdent(c.Channel)))
		}
		g.indent++
		guardedSkip := c.IsSkip && c.Guard != nil
		if guardedSkip {
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write("if ")
			g.generateExpression(c.Guard)
			g.write(" {\n")
			g.indent++
		}
		for _, s := range c.Body {
			g.generateStatement(s)
		}
		if guardedSkip {
			g.indent--
			g.writeLine("}")
		}
		g.indent--
	}
	g.writeLine("}")
}

func (g *Generator) generateReplicatedAlt(alt *ast.AltBlock) {
	// Replicated ALT: ALT i = start FOR count
	// Uses reflect.Select for runtime-variable case count
	if len(alt.Cases) == 0 {
		return
	}
	c := alt.Cases[0]
	rep := alt.Replicator
	v := goIdent(rep.Variable)

	// Determine receive type from scoped declarations
	recvType := "int" // default
	for _, decl := range c.Declarations {
		if vd, ok := decl.(*ast.VarDecl); ok {
			for _, name := range vd.Names {
				if name == c.Variable {
					recvType = g.occamTypeToGo(vd.Type)
					break
				}
			}
		}
	}

	// Open a block for scoping
	g.writeLine("{")
	g.indent++

	// _altCount := int(<count>)
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write("_altCount := int(")
	g.generateExpression(rep.Count)
	g.write(")\n")

	// _altCases := make([]reflect.SelectCase, _altCount)
	g.writeLine("_altCases := make([]reflect.SelectCase, _altCount)")

	// Setup loop: build select cases
	g.writeLine("for _altI := 0; _altI < _altCount; _altI++ {")
	g.indent++

	// Compute replicator variable
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	if rep.Step != nil {
		g.write(fmt.Sprintf("%s := ", v))
		g.generateExpression(rep.Start)
		g.write(" + _altI * (")
		g.generateExpression(rep.Step)
		g.write(")\n")
	} else {
		g.write(fmt.Sprintf("%s := ", v))
		g.generateExpression(rep.Start)
		g.write(" + _altI\n")
	}

	// Generate scoped abbreviations (needed for channel index computation)
	for _, decl := range c.Declarations {
		if abbr, ok := decl.(*ast.Abbreviation); ok {
			g.generateAbbreviation(abbr)
		}
	}

	// Build select case entry
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write("_altCases[_altI] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(")
	if len(c.ChannelIndices) > 0 {
		g.write(goIdent(c.Channel))
		g.generateIndices(c.ChannelIndices)
	} else {
		g.write(goIdent(c.Channel))
	}
	g.write(")}\n")

	g.indent--
	g.writeLine("}")

	// Call reflect.Select
	g.writeLine("_altChosen, _altValue, _ := reflect.Select(_altCases)")

	// Recompute replicator variable from chosen index
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	if rep.Step != nil {
		g.write(fmt.Sprintf("%s := ", v))
		g.generateExpression(rep.Start)
		g.write(" + _altChosen * (")
		g.generateExpression(rep.Step)
		g.write(")\n")
	} else {
		g.write(fmt.Sprintf("%s := ", v))
		g.generateExpression(rep.Start)
		g.write(" + _altChosen\n")
	}
	g.writeLine(fmt.Sprintf("_ = %s", v))

	// Generate scoped var declarations
	for _, decl := range c.Declarations {
		if vd, ok := decl.(*ast.VarDecl); ok {
			g.generateVarDecl(vd)
		}
	}

	// Generate scoped abbreviations
	for _, decl := range c.Declarations {
		if abbr, ok := decl.(*ast.Abbreviation); ok {
			g.generateAbbreviation(abbr)
		}
	}

	// Assign received value from reflect.Value
	varRef := goIdent(c.Variable)
	if len(c.VariableIndices) > 0 {
		varRef += g.generateIndicesStr(c.VariableIndices)
	}
	g.writeLine(fmt.Sprintf("%s = _altValue.Interface().(%s)", varRef, recvType))

	// Generate body
	for _, s := range c.Body {
		g.generateStatement(s)
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateProcDecl(proc *ast.ProcDecl) {
	// Track reference parameters for this procedure
	oldRefParams := g.refParams
	newRefParams := make(map[string]bool)
	// Scope boolVars per proc body
	oldBoolVars := g.boolVars
	newBoolVars := make(map[string]bool)
	// Inherit parent's ref params and boolVars for closure captures when nested
	if g.nestingLevel > 0 {
		for k, v := range oldRefParams {
			newRefParams[k] = v
		}
		for k, v := range oldBoolVars {
			newBoolVars[k] = v
		}
	}
	for _, p := range proc.Params {
		if !p.IsVal && !p.IsChan && p.ChanArrayDims == 0 && p.OpenArrayDims == 0 && p.ArraySize == "" {
			newRefParams[p.Name] = true
		} else {
			// Own param shadows any inherited ref param with same name
			delete(newRefParams, p.Name)
		}
		// Track BOOL params; delete non-BOOL params that shadow inherited names
		if p.Type == "BOOL" && !p.IsChan && p.ChanArrayDims == 0 {
			newBoolVars[p.Name] = true
		} else {
			delete(newBoolVars, p.Name)
		}
		// Register chan params with protocol mappings and element types
		if p.IsChan || p.ChanArrayDims > 0 {
			if _, ok := g.protocolDefs[p.ChanElemType]; ok {
				g.chanProtocols[p.Name] = p.ChanElemType
			}
			g.chanElemTypes[p.Name] = g.occamTypeToGo(p.ChanElemType)
		}
		// Register record-typed params
		if !p.IsChan {
			if _, ok := g.recordDefs[p.Type]; ok {
				g.recordVars[p.Name] = p.Type
			}
		}
	}
	g.refParams = newRefParams
	g.boolVars = newBoolVars

	// Scan proc body for RETYPES declarations that shadow parameters.
	// When VAL INT X RETYPES X :, Go can't redeclare X in the same scope,
	// so we rename the parameter (e.g. X → _rp_X) and let RETYPES declare the original name.
	oldRenames := g.retypesRenames
	g.retypesRenames = nil
	paramNames := make(map[string]bool)
	for _, p := range proc.Params {
		paramNames[p.Name] = true
	}
	for _, stmt := range proc.Body {
		if rd, ok := stmt.(*ast.RetypesDecl); ok {
			if paramNames[rd.Source] && rd.Name == rd.Source {
				if g.retypesRenames == nil {
					g.retypesRenames = make(map[string]string)
				}
				g.retypesRenames[rd.Name] = "_rp_" + goIdent(rd.Name)
			}
		}
	}

	// Generate function signature
	params := g.generateProcParams(proc.Params)
	gName := goIdent(proc.Name)
	if g.nestingLevel > 0 {
		// Nested PROC: generate as Go closure
		g.writeLine(fmt.Sprintf("%s := func(%s) {", gName, params))
	} else {
		g.writeLine(fmt.Sprintf("func %s(%s) {", gName, params))
	}
	g.indent++
	g.nestingLevel++

	// Register nested proc/func signatures for this scope so that calls
	// within this proc resolve to the correct (local) signature rather than
	// a same-named proc from a different scope.
	oldSigs := make(map[string][]ast.ProcParam)
	g.collectNestedProcSigsScoped(proc.Body, oldSigs)

	for _, stmt := range proc.Body {
		g.generateStatement(stmt)
	}

	// Restore overwritten signatures
	for name, params := range oldSigs {
		if params == nil {
			delete(g.procSigs, name)
		} else {
			g.procSigs[name] = params
		}
	}

	g.nestingLevel--
	g.indent--
	g.writeLine("}")
	g.writeLine("")

	// Restore previous context
	g.refParams = oldRefParams
	g.boolVars = oldBoolVars
	g.retypesRenames = oldRenames
}

func (g *Generator) generateProcParams(params []ast.ProcParam) string {
	var parts []string
	for _, p := range params {
		var goType string
		if p.ChanArrayDims > 0 {
			goType = strings.Repeat("[]", p.ChanArrayDims) + "chan " + g.occamTypeToGo(p.ChanElemType)
		} else if p.IsChan {
			goType = chanDirPrefix(p.ChanDir) + g.occamTypeToGo(p.ChanElemType)
		} else if p.OpenArrayDims > 0 {
			goType = strings.Repeat("[]", p.OpenArrayDims) + g.occamTypeToGo(p.Type)
		} else if p.ArraySize != "" {
			// Fixed-size array parameter: use slice for Go compatibility
			// (occam [n]TYPE and []TYPE both map to Go slices)
			goType = "[]" + g.occamTypeToGo(p.Type)
		} else {
			goType = g.occamTypeToGo(p.Type)
			if !p.IsVal {
				// Non-VAL parameters are pass by reference in Occam
				goType = "*" + goType
			}
		}
		pName := goIdent(p.Name)
		if renamed, ok := g.retypesRenames[p.Name]; ok {
			pName = renamed
		}
		parts = append(parts, fmt.Sprintf("%s %s", pName, goType))
	}
	return strings.Join(parts, ", ")
}

func chanDirPrefix(dir string) string {
	switch dir {
	case "?":
		return "<-chan " // input/receive-only
	case "!":
		return "chan<- " // output/send-only
	default:
		return "chan " // bidirectional
	}
}

func (g *Generator) generateProcCall(call *ast.ProcCall) {
	// Handle built-in print procedures
	if printBuiltins[call.Name] {
		g.generatePrintCall(call)
		return
	}

	// Handle CAUSEERROR
	if call.Name == "CAUSEERROR" {
		g.writeLine(`panic("CAUSEERROR")`)
		return
	}

	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write(goIdent(call.Name))
	g.write("(")

	// Look up procedure signature to determine which args need address-of
	params := g.procSigs[call.Name]

	for i, arg := range call.Args {
		if i > 0 {
			g.write(", ")
		}
		// If this parameter is not VAL (i.e., pass by reference), take address
		// Channels, channel arrays, open arrays, and fixed-size arrays (mapped to slices) are already reference types
		if i < len(params) && !params[i].IsVal && !params[i].IsChan && params[i].ChanArrayDims == 0 && params[i].OpenArrayDims == 0 && params[i].ArraySize == "" {
			g.write("&")
		}
		// Wrap string literals with []byte() when passed to []BYTE parameters
		if _, isStr := arg.(*ast.StringLiteral); isStr && i < len(params) && params[i].OpenArrayDims > 0 && params[i].Type == "BYTE" {
			g.write("[]byte(")
			g.generateExpression(arg)
			g.write(")")
		} else {
			g.generateExpression(arg)
		}
	}
	g.write(")")
	g.write("\n")
}

func (g *Generator) generateFuncDecl(fn *ast.FuncDecl) {
	params := g.generateProcParams(fn.Params)

	// Build return type string
	var returnTypeStr string
	if len(fn.ReturnTypes) == 1 {
		returnTypeStr = g.occamTypeToGo(fn.ReturnTypes[0])
	} else {
		goTypes := make([]string, len(fn.ReturnTypes))
		for i, rt := range fn.ReturnTypes {
			goTypes[i] = g.occamTypeToGo(rt)
		}
		returnTypeStr = "(" + strings.Join(goTypes, ", ") + ")"
	}

	// Scope boolVars per function body
	oldBoolVars := g.boolVars
	newBoolVars := make(map[string]bool)
	if g.nestingLevel > 0 {
		for k, v := range oldBoolVars {
			newBoolVars[k] = v
		}
	}
	for _, p := range fn.Params {
		if p.Type == "BOOL" && !p.IsChan && p.ChanArrayDims == 0 {
			newBoolVars[p.Name] = true
		} else {
			delete(newBoolVars, p.Name)
		}
	}
	g.boolVars = newBoolVars

	gName := goIdent(fn.Name)
	if g.nestingLevel > 0 {
		// Nested FUNCTION: generate as Go closure
		g.writeLine(fmt.Sprintf("%s := func(%s) %s {", gName, params, returnTypeStr))
	} else {
		g.writeLine(fmt.Sprintf("func %s(%s) %s {", gName, params, returnTypeStr))
	}
	g.indent++
	g.nestingLevel++

	for _, stmt := range fn.Body {
		g.generateStatement(stmt)
	}

	if len(fn.ResultExprs) > 0 {
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write("return ")
		for i, expr := range fn.ResultExprs {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(expr)
		}
		g.write("\n")
	}

	g.nestingLevel--
	g.indent--
	g.writeLine("}")
	g.writeLine("")

	// Restore previous boolVars
	g.boolVars = oldBoolVars
}

func (g *Generator) generateFuncCallExpr(call *ast.FuncCall) {
	if transpIntrinsics[call.Name] {
		g.write("_" + call.Name)
	} else {
		g.write(goIdent(call.Name))
	}
	g.write("(")
	params := g.procSigs[call.Name]
	for i, arg := range call.Args {
		if i > 0 {
			g.write(", ")
		}
		// Wrap string literals with []byte() when passed to []BYTE parameters
		if _, isStr := arg.(*ast.StringLiteral); isStr && i < len(params) && params[i].OpenArrayDims > 0 && params[i].Type == "BYTE" {
			g.write("[]byte(")
			g.generateExpression(arg)
			g.write(")")
		} else {
			g.generateExpression(arg)
		}
	}
	g.write(")")
}

func (g *Generator) generateMultiAssignment(stmt *ast.MultiAssignment) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	for i, target := range stmt.Targets {
		if i > 0 {
			g.write(", ")
		}
		if len(target.Indices) > 0 {
			// Check if this is a record field access (single index that is an identifier)
			if len(target.Indices) == 1 {
				if _, ok := g.recordVars[target.Name]; ok {
					if ident, ok := target.Indices[0].(*ast.Identifier); ok {
						g.write(goIdent(target.Name))
						g.write(".")
						g.write(goIdent(ident.Value))
						continue
					}
				}
			}
			if g.refParams[target.Name] {
				g.write("(*")
				g.write(goIdent(target.Name))
				g.write(")")
			} else {
				g.write(goIdent(target.Name))
			}
			g.generateIndices(target.Indices)
		} else {
			if g.refParams[target.Name] {
				g.write("*")
			}
			g.write(goIdent(target.Name))
		}
	}
	g.write(" = ")
	for i, val := range stmt.Values {
		if i > 0 {
			g.write(", ")
		}
		g.generateExpression(val)
	}
	g.write("\n")
}

func (g *Generator) generatePrintCall(call *ast.ProcCall) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))

	switch call.Name {
	case "print.int", "print.string", "print.bool":
		g.write("fmt.Println(")
		if len(call.Args) > 0 {
			g.generateExpression(call.Args[0])
		}
		g.write(")")
	case "print.newline":
		g.write("fmt.Println()")
	}

	g.write("\n")
}

func (g *Generator) generateWhileLoop(loop *ast.WhileLoop) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write("for ")
	g.generateExpression(loop.Condition)
	g.write(" {\n")
	g.indent++

	for _, s := range loop.Body {
		g.generateStatement(s)
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateIfStatement(stmt *ast.IfStatement) {
	if stmt.Replicator != nil {
		// Replicated IF: IF i = start FOR count → for loop with break on first match
		g.generateReplicatedIfLoop(stmt, false)
	} else {
		// Flatten non-replicated nested IFs into the parent choice list
		choices := g.flattenIfChoices(stmt.Choices)
		g.generateIfChoiceChain(choices, true)
	}
}

// flattenIfChoices inlines choices from non-replicated nested IFs into a flat list.
// Replicated nested IFs are preserved as-is (they need special loop codegen).
func (g *Generator) flattenIfChoices(choices []ast.IfChoice) []ast.IfChoice {
	var flat []ast.IfChoice
	for _, c := range choices {
		if c.NestedIf != nil && c.NestedIf.Replicator == nil {
			// Non-replicated nested IF: inline its choices recursively
			flat = append(flat, g.flattenIfChoices(c.NestedIf.Choices)...)
		} else {
			flat = append(flat, c)
		}
	}
	return flat
}

// generateReplicatedIfLoop emits a for loop that breaks on first matching choice.
// When withinFlag is true, it sets the named flag to true before breaking.
func (g *Generator) generateReplicatedIfLoop(stmt *ast.IfStatement, withinFlag bool, flagName ...string) {
	repl := stmt.Replicator
	v := goIdent(repl.Variable)
	if repl.Step != nil {
		counter := "_repl_" + v
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("for %s := 0; %s < ", counter, counter))
		g.generateExpression(repl.Count)
		g.write(fmt.Sprintf("; %s++ {\n", counter))
		g.indent++
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("%s := ", v))
		g.generateExpression(repl.Start)
		g.write(fmt.Sprintf(" + %s * ", counter))
		g.generateExpression(repl.Step)
		g.write("\n")
	} else {
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("for %s := ", v))
		g.generateExpression(repl.Start)
		g.write(fmt.Sprintf("; %s < ", v))
		g.generateExpression(repl.Start)
		g.write(" + ")
		g.generateExpression(repl.Count)
		g.write(fmt.Sprintf("; %s++ {\n", v))
		g.indent++
	}

	for i, choice := range stmt.Choices {
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		if i == 0 {
			g.write("if ")
		} else {
			g.write("} else if ")
		}
		g.generateExpression(choice.Condition)
		g.write(" {\n")
		g.indent++

		for _, s := range choice.Body {
			g.generateStatement(s)
		}
		if withinFlag && len(flagName) > 0 {
			g.writeLine(fmt.Sprintf("%s = true", flagName[0]))
		}
		g.writeLine("break")

		g.indent--
	}
	g.writeLine("}")

	g.indent--
	g.writeLine("}")
}

// generateIfChoiceChain emits a chain of if/else-if for the given choices.
// When a replicated nested IF is encountered, it splits the chain and uses
// a _ifmatched flag to determine whether remaining choices should be tried.
func (g *Generator) generateIfChoiceChain(choices []ast.IfChoice, isFirst bool) {
	// Find first replicated nested IF
	replIdx := -1
	for i, c := range choices {
		if c.NestedIf != nil && c.NestedIf.Replicator != nil {
			replIdx = i
			break
		}
	}

	if replIdx == -1 {
		// No replicated nested IFs — simple if/else-if chain
		for i, choice := range choices {
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			if i == 0 && isFirst {
				g.write("if ")
			} else {
				g.write("} else if ")
			}
			g.generateExpression(choice.Condition)
			g.write(" {\n")
			g.indent++

			for _, s := range choice.Body {
				g.generateStatement(s)
			}

			g.indent--
		}
		if len(choices) > 0 {
			g.writeLine("}")
		}
		return
	}

	// Split at the replicated nested IF
	before := choices[:replIdx]
	replChoice := choices[replIdx]
	after := choices[replIdx+1:]

	// Emit choices before the replicated IF as a normal if-else chain
	if len(before) > 0 {
		for i, choice := range before {
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			if i == 0 && isFirst {
				g.write("if ")
			} else {
				g.write("} else if ")
			}
			g.generateExpression(choice.Condition)
			g.write(" {\n")
			g.indent++
			for _, s := range choice.Body {
				g.generateStatement(s)
			}
			g.indent--
		}
		// Open else block for the replicated IF + remaining choices
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write("} else {\n")
		g.indent++
	}

	// Emit the replicated nested IF with a flag
	needFlag := len(after) > 0
	flagName := fmt.Sprintf("_ifmatched%d", g.tmpCounter)
	g.tmpCounter++
	if needFlag {
		g.writeLine(fmt.Sprintf("%s := false", flagName))
	}
	g.generateReplicatedIfLoop(replChoice.NestedIf, needFlag, flagName)

	// Emit remaining choices inside if !flagName (recursive for multiple)
	if len(after) > 0 {
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("if !%s {\n", flagName))
		g.indent++
		g.generateIfChoiceChain(after, true) // recursive for remaining
		g.indent--
		g.writeLine("}")
	}

	if len(before) > 0 {
		g.indent--
		g.writeLine("}")
	}
}

func (g *Generator) generateCaseStatement(stmt *ast.CaseStatement) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write("switch ")
	g.generateExpression(stmt.Selector)
	g.write(" {\n")

	for _, choice := range stmt.Choices {
		if choice.IsElse {
			g.writeLine("default:")
		} else {
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write("case ")
			for i, val := range choice.Values {
				if i > 0 {
					g.write(", ")
				}
				g.generateExpression(val)
			}
			g.write(":\n")
		}
		g.indent++
		for _, s := range choice.Body {
			g.generateStatement(s)
		}
		g.indent--
	}

	g.writeLine("}")
}

func (g *Generator) generateExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		if g.refParams[e.Value] {
			g.write("*" + goIdent(e.Value))
		} else {
			g.write(goIdent(e.Value))
		}
	case *ast.IntegerLiteral:
		g.write(fmt.Sprintf("%d", e.Value))
	case *ast.StringLiteral:
		g.write(fmt.Sprintf("%q", e.Value))
	case *ast.ByteLiteral:
		g.write(fmt.Sprintf("byte(%d)", e.Value))
	case *ast.BooleanLiteral:
		if e.Value {
			g.write("true")
		} else {
			g.write("false")
		}
	case *ast.BinaryExpr:
		g.generateBinaryExpr(e)
	case *ast.UnaryExpr:
		g.generateUnaryExpr(e)
	case *ast.SizeExpr:
		g.write("len(")
		g.generateExpression(e.Expr)
		g.write(")")
	case *ast.ParenExpr:
		g.write("(")
		g.generateExpression(e.Expr)
		g.write(")")
	case *ast.IndexExpr:
		// Check if this is a record field access
		if ident, ok := e.Left.(*ast.Identifier); ok {
			if _, ok := g.recordVars[ident.Value]; ok {
				if field, ok := e.Index.(*ast.Identifier); ok {
					g.generateExpression(e.Left)
					g.write(".")
					g.write(goIdent(field.Value))
					break
				}
			}
		}
		g.generateExpression(e.Left)
		g.write("[")
		g.generateExpression(e.Index)
		g.write("]")
	case *ast.SliceExpr:
		g.generateExpression(e.Array)
		g.write("[")
		g.generateExpression(e.Start)
		g.write(" : ")
		g.generateExpression(e.Start)
		g.write(" + ")
		g.generateExpression(e.Length)
		g.write("]")
	case *ast.FuncCall:
		g.generateFuncCallExpr(e)
	case *ast.TypeConversion:
		if e.TargetType == "BOOL" {
			// numeric → bool: emit ((expr) != 0)
			g.write("((")
			g.generateExpression(e.Expr)
			g.write(") != 0)")
		} else if g.isBoolExpression(e.Expr) {
			// bool → numeric: emit type(_boolToInt(expr))
			goType := g.occamTypeToGo(e.TargetType)
			if goType == "int" {
				g.write("_boolToInt(")
				g.generateExpression(e.Expr)
				g.write(")")
			} else {
				g.write(goType)
				g.write("(_boolToInt(")
				g.generateExpression(e.Expr)
				g.write("))")
			}
		} else if e.Qualifier == "ROUND" && isOccamIntType(e.TargetType) {
			// float → int with ROUND: emit goType(math.Round(float64(expr)))
			goType := g.occamTypeToGo(e.TargetType)
			g.write(goType)
			g.write("(math.Round(float64(")
			g.generateExpression(e.Expr)
			g.write(")))")
		} else {
			g.write(g.occamTypeToGo(e.TargetType))
			g.write("(")
			g.generateExpression(e.Expr)
			g.write(")")
		}
	case *ast.MostExpr:
		g.generateMostExpr(e)
	case *ast.ArrayLiteral:
		g.generateArrayLiteral(e)
	}
}

func (g *Generator) generateBinaryExpr(expr *ast.BinaryExpr) {
	g.write("(")
	g.generateExpression(expr.Left)
	g.write(" ")
	g.write(g.occamOpToGo(expr.Operator))
	g.write(" ")
	g.generateExpression(expr.Right)
	g.write(")")
}

func (g *Generator) generateUnaryExpr(expr *ast.UnaryExpr) {
	op := g.occamOpToGo(expr.Operator)
	g.write(op)
	if op == "!" || op == "^" {
		// Go's logical NOT and bitwise NOT don't need space
	} else {
		g.write(" ")
	}
	g.generateExpression(expr.Right)
}

func (g *Generator) occamOpToGo(op string) string {
	switch op {
	case "=":
		return "=="
	case "<>":
		return "!="
	case "AND":
		return "&&"
	case "OR":
		return "||"
	case "NOT":
		return "!"
	case "\\":
		return "%"
	case "AFTER":
		return ">"
	case "/\\":
		return "&"
	case "\\/":
		return "|"
	case "><":
		return "^"
	case "~":
		return "^"
	case "<<":
		return "<<"
	case ">>":
		return ">>"
	case "PLUS":
		return "+"
	case "MINUS":
		return "-"
	case "TIMES":
		return "*"
	default:
		return op // +, -, *, /, <, >, <=, >= are the same
	}
}

// generateArrayLiteral emits a Go slice literal: []int{e1, e2, ...}
func (g *Generator) generateArrayLiteral(al *ast.ArrayLiteral) {
	g.write("[]int{")
	for i, elem := range al.Elements {
		if i > 0 {
			g.write(", ")
		}
		g.generateExpression(elem)
	}
	g.write("}")
}

// generateRetypesDecl emits code for a RETYPES declaration.
// VAL INT X RETYPES X : — reinterpret float32/64 bits as int(s)
// When source and target share the same name (shadowing a parameter), the parameter
// has been renamed in the signature (e.g. X → _rp_X) so we can use := with the
// original name to create a new variable.
func (g *Generator) generateRetypesDecl(r *ast.RetypesDecl) {
	gName := goIdent(r.Name)
	gSource := goIdent(r.Source)
	// If the parameter was renamed for RETYPES shadowing, use the renamed source
	if renamed, ok := g.retypesRenames[r.Source]; ok {
		gSource = renamed
	}
	if r.IsArray {
		// VAL [2]INT X RETYPES X : — split float64 into two int32 words
		tmpVar := fmt.Sprintf("_retmp%d", g.tmpCounter)
		g.tmpCounter++
		g.writeLine(fmt.Sprintf("%s := math.Float64bits(float64(%s))", tmpVar, gSource))
		g.writeLine(fmt.Sprintf("%s := []int{int(int32(uint32(%s))), int(int32(uint32(%s >> 32)))}", gName, tmpVar, tmpVar))
	} else {
		// VAL INT X RETYPES X : — reinterpret float32 as int
		g.writeLine(fmt.Sprintf("%s := int(int32(math.Float32bits(float32(%s))))", gName, gSource))
	}
}

// containsIntrinsics checks if a statement tree contains transputer intrinsic calls.
func (g *Generator) containsIntrinsics(stmt ast.Statement) bool {
	return g.walkStatements(stmt, func(e ast.Expression) bool {
		if fc, ok := e.(*ast.FuncCall); ok {
			return transpIntrinsics[fc.Name]
		}
		return false
	})
}

// containsBoolConversion checks if a statement tree contains a bool-to-numeric type conversion.
func (g *Generator) containsBoolConversion(stmt ast.Statement) bool {
	return g.walkStatements(stmt, func(e ast.Expression) bool {
		tc, ok := e.(*ast.TypeConversion)
		if !ok {
			return false
		}
		// Only need the helper for bool→numeric (not numeric→bool)
		return tc.TargetType != "BOOL" && g.isBoolExpression(tc.Expr)
	})
}

// isBoolExpression returns true if the expression is known to produce a bool value.
func (g *Generator) isBoolExpression(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.BooleanLiteral:
		return true
	case *ast.Identifier:
		return g.boolVars[e.Value]
	case *ast.UnaryExpr:
		return e.Operator == "NOT"
	case *ast.BinaryExpr:
		switch e.Operator {
		case "=", "<>", "<", ">", "<=", ">=", "AND", "OR", "AFTER":
			return true
		}
	case *ast.TypeConversion:
		return e.TargetType == "BOOL"
	case *ast.ParenExpr:
		return g.isBoolExpression(e.Expr)
	}
	return false
}

// emitBoolHelper writes the _boolToInt helper function.
func (g *Generator) emitBoolHelper() {
	g.writeLine("func _boolToInt(b bool) int {")
	g.indent++
	g.writeLine("if b {")
	g.indent++
	g.writeLine("return 1")
	g.indent--
	g.writeLine("}")
	g.writeLine("return 0")
	g.indent--
	g.writeLine("}")
	g.writeLine("")
}

// containsRetypes checks if a statement tree contains RETYPES declarations.
func (g *Generator) containsRetypes(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.RetypesDecl:
		return true
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			if g.containsRetypes(inner) {
				return true
			}
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			if g.containsRetypes(inner) {
				return true
			}
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			if g.containsRetypes(inner) {
				return true
			}
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			if g.containsRetypes(inner) {
				return true
			}
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			if g.containsRetypes(inner) {
				return true
			}
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil && g.containsRetypes(choice.NestedIf) {
				return true
			}
			for _, inner := range choice.Body {
				if g.containsRetypes(inner) {
					return true
				}
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				if g.containsRetypes(inner) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Generator) containsAltReplicator(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.AltBlock:
		if s.Replicator != nil {
			return true
		}
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.containsAltReplicator(inner) {
					return true
				}
			}
		}
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			if g.containsAltReplicator(inner) {
				return true
			}
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			if g.containsAltReplicator(inner) {
				return true
			}
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			if g.containsAltReplicator(inner) {
				return true
			}
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			if g.containsAltReplicator(inner) {
				return true
			}
		}
	case *ast.WhileLoop:
		for _, inner := range s.Body {
			if g.containsAltReplicator(inner) {
				return true
			}
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil && g.containsAltReplicator(choice.NestedIf) {
				return true
			}
			for _, inner := range choice.Body {
				if g.containsAltReplicator(inner) {
					return true
				}
			}
		}
	case *ast.CaseStatement:
		for _, choice := range s.Choices {
			for _, inner := range choice.Body {
				if g.containsAltReplicator(inner) {
					return true
				}
			}
		}
	}
	return false
}

// walkStatements recursively walks a statement tree, applying fn to all expressions.
// Returns true if fn returns true for any expression.
func (g *Generator) walkStatements(stmt ast.Statement, fn func(ast.Expression) bool) bool {
	switch s := stmt.(type) {
	case *ast.Assignment:
		result := g.walkExpr(s.Value, fn)
		for _, idx := range s.Indices {
			result = result || g.walkExpr(idx, fn)
		}
		return result
	case *ast.MultiAssignment:
		for _, v := range s.Values {
			if g.walkExpr(v, fn) {
				return true
			}
		}
	case *ast.Abbreviation:
		return g.walkExpr(s.Value, fn)
	case *ast.SeqBlock:
		for _, inner := range s.Statements {
			if g.walkStatements(inner, fn) {
				return true
			}
		}
	case *ast.ParBlock:
		for _, inner := range s.Statements {
			if g.walkStatements(inner, fn) {
				return true
			}
		}
	case *ast.ProcDecl:
		for _, inner := range s.Body {
			if g.walkStatements(inner, fn) {
				return true
			}
		}
	case *ast.FuncDecl:
		for _, inner := range s.Body {
			if g.walkStatements(inner, fn) {
				return true
			}
		}
	case *ast.WhileLoop:
		if g.walkExpr(s.Condition, fn) {
			return true
		}
		for _, inner := range s.Body {
			if g.walkStatements(inner, fn) {
				return true
			}
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.NestedIf != nil && g.walkStatements(choice.NestedIf, fn) {
				return true
			}
			if g.walkExpr(choice.Condition, fn) {
				return true
			}
			for _, inner := range choice.Body {
				if g.walkStatements(inner, fn) {
					return true
				}
			}
		}
	case *ast.CaseStatement:
		if g.walkExpr(s.Selector, fn) {
			return true
		}
		for _, choice := range s.Choices {
			for _, v := range choice.Values {
				if g.walkExpr(v, fn) {
					return true
				}
			}
			for _, inner := range choice.Body {
				if g.walkStatements(inner, fn) {
					return true
				}
			}
		}
	case *ast.Send:
		if g.walkExpr(s.Value, fn) {
			return true
		}
		for _, v := range s.Values {
			if g.walkExpr(v, fn) {
				return true
			}
		}
	case *ast.ProcCall:
		for _, arg := range s.Args {
			if g.walkExpr(arg, fn) {
				return true
			}
		}
	case *ast.AltBlock:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.walkStatements(inner, fn) {
					return true
				}
			}
		}
	case *ast.VariantReceive:
		for _, c := range s.Cases {
			for _, inner := range c.Body {
				if g.walkStatements(inner, fn) {
					return true
				}
			}
		}
	}
	return false
}

// walkExpr recursively walks an expression tree, applying fn.
func (g *Generator) walkExpr(expr ast.Expression, fn func(ast.Expression) bool) bool {
	if expr == nil {
		return false
	}
	if fn(expr) {
		return true
	}
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		return g.walkExpr(e.Left, fn) || g.walkExpr(e.Right, fn)
	case *ast.UnaryExpr:
		return g.walkExpr(e.Right, fn)
	case *ast.ParenExpr:
		return g.walkExpr(e.Expr, fn)
	case *ast.TypeConversion:
		return g.walkExpr(e.Expr, fn)
	case *ast.SizeExpr:
		return g.walkExpr(e.Expr, fn)
	case *ast.IndexExpr:
		return g.walkExpr(e.Left, fn) || g.walkExpr(e.Index, fn)
	case *ast.FuncCall:
		for _, arg := range e.Args {
			if g.walkExpr(arg, fn) {
				return true
			}
		}
	case *ast.SliceExpr:
		return g.walkExpr(e.Array, fn) || g.walkExpr(e.Start, fn) || g.walkExpr(e.Length, fn)
	case *ast.ArrayLiteral:
		for _, elem := range e.Elements {
			if g.walkExpr(elem, fn) {
				return true
			}
		}
	}
	return false
}

// emitIntrinsicHelpers writes the Go helper functions for transputer intrinsics.
// These implement 32-bit transputer semantics using uint32/uint64 arithmetic.
func (g *Generator) emitIntrinsicHelpers() {
	g.writeLine("// Transputer intrinsic helper functions")
	g.writeLine("func _LONGPROD(a, b, c int) (int, int) {")
	g.writeLine("\tr := uint64(uint32(a))*uint64(uint32(b)) + uint64(uint32(c))")
	g.writeLine("\treturn int(int32(uint32(r >> 32))), int(int32(uint32(r)))")
	g.writeLine("}")
	g.writeLine("")
	g.writeLine("func _LONGDIV(hi, lo, divisor int) (int, int) {")
	g.writeLine("\tn := (uint64(uint32(hi)) << 32) | uint64(uint32(lo))")
	g.writeLine("\td := uint64(uint32(divisor))")
	g.writeLine("\tif d == 0 { panic(\"LONGDIV: division by zero\") }")
	g.writeLine("\treturn int(int32(uint32(n / d))), int(int32(uint32(n % d)))")
	g.writeLine("}")
	g.writeLine("")
	g.writeLine("func _LONGSUM(a, b, carry int) (int, int) {")
	g.writeLine("\tr := uint64(uint32(a)) + uint64(uint32(b)) + uint64(uint32(carry))")
	g.writeLine("\treturn int(int32(uint32(r >> 32))), int(int32(uint32(r)))")
	g.writeLine("}")
	g.writeLine("")
	g.writeLine("func _LONGDIFF(a, b, borrow int) (int, int) {")
	g.writeLine("\tr := uint64(uint32(a)) - uint64(uint32(b)) - uint64(uint32(borrow))")
	g.writeLine("\tif uint32(a) >= uint32(b)+uint32(borrow) { return 0, int(int32(uint32(r))) }")
	g.writeLine("\treturn 1, int(int32(uint32(r)))")
	g.writeLine("}")
	g.writeLine("")
	g.writeLine("func _NORMALISE(hi, lo int) (int, int, int) {")
	g.writeLine("\tv := (uint64(uint32(hi)) << 32) | uint64(uint32(lo))")
	g.writeLine("\tif v == 0 { return 64, 0, 0 }")
	g.writeLine("\tn := bits.LeadingZeros64(v)")
	g.writeLine("\tv <<= uint(n)")
	g.writeLine("\treturn n, int(int32(uint32(v >> 32))), int(int32(uint32(v)))")
	g.writeLine("}")
	g.writeLine("")
	g.writeLine("func _SHIFTRIGHT(hi, lo, n int) (int, int) {")
	g.writeLine("\tv := (uint64(uint32(hi)) << 32) | uint64(uint32(lo))")
	g.writeLine("\tv >>= uint(uint32(n))")
	g.writeLine("\treturn int(int32(uint32(v >> 32))), int(int32(uint32(v)))")
	g.writeLine("}")
	g.writeLine("")
	g.writeLine("func _SHIFTLEFT(hi, lo, n int) (int, int) {")
	g.writeLine("\tv := (uint64(uint32(hi)) << 32) | uint64(uint32(lo))")
	g.writeLine("\tv <<= uint(uint32(n))")
	g.writeLine("\treturn int(int32(uint32(v >> 32))), int(int32(uint32(v)))")
	g.writeLine("}")
	g.writeLine("")
}
