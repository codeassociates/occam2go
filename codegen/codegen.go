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

	// Nesting level: 0 = package level, >0 = inside a function
	nestingLevel int
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
func goIdent(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

// Generate produces Go code from the AST
func (g *Generator) Generate(program *ast.Program) string {
	g.builder.Reset()
	g.needSync = false
	g.needFmt = false
	g.needTime = false
	g.needOs = false
	g.needMath = false
	g.procSigs = make(map[string][]ast.ProcParam)
	g.refParams = make(map[string]bool)
	g.protocolDefs = make(map[string]*ast.ProtocolDecl)
	g.chanProtocols = make(map[string]string)
	g.tmpCounter = 0
	g.recordDefs = make(map[string]*ast.RecordDecl)
	g.recordVars = make(map[string]string)

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

	// Write package declaration
	g.writeLine("package main")
	g.writeLine("")

	// Write imports
	if g.needSync || g.needFmt || g.needTime || g.needOs || g.needMath {
		g.writeLine("import (")
		g.indent++
		if g.needFmt {
			g.writeLine(`"fmt"`)
		}
		if g.needMath {
			g.writeLine(`"math"`)
		}
		if g.needOs {
			g.writeLine(`"os"`)
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
		switch stmt.(type) {
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
		default:
			mainStatements = append(mainStatements, stmt)
		}
	}

	// Generate type definitions first (at package level)
	for _, stmt := range typeDecls {
		g.generateStatement(stmt)
	}

	// Generate package-level abbreviations (constants)
	for _, stmt := range abbrDecls {
		abbr := stmt.(*ast.Abbreviation)
		goType := g.occamTypeToGo(abbr.Type)
		if abbr.IsOpenArray {
			goType = "[]" + goType
		}
		g.builder.WriteString("var ")
		g.write(fmt.Sprintf("%s %s = ", goIdent(abbr.Name), goType))
		g.generateExpression(abbr.Value)
		g.write("\n")
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
	}

	return g.builder.String()
}

// collectNestedProcSigs recursively collects procedure/function signatures
// from nested declarations inside PROC bodies.
func (g *Generator) collectNestedProcSigs(stmts []ast.Statement) {
	for _, stmt := range stmts {
		if proc, ok := stmt.(*ast.ProcDecl); ok {
			g.procSigs[proc.Name] = proc.Params
			g.collectNestedProcSigs(proc.Body)
		}
		if fn, ok := stmt.(*ast.FuncDecl); ok {
			g.procSigs[fn.Name] = fn.Params
		}
	}
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
			if c.Body != nil && g.containsPar(c.Body) {
				return true
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
			if c.Body != nil && g.containsPrint(c.Body) {
				return true
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
			if c.Body != nil && g.containsTimer(c.Body) {
				return true
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
			if c.Body != nil && g.containsStop(c.Body) {
				return true
			}
		}
	}
	return false
}

func (g *Generator) containsMostExpr(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.Assignment:
		return g.exprNeedsMath(s.Value) || g.exprNeedsMath(s.Index)
	case *ast.MultiAssignment:
		for _, t := range s.Targets {
			if g.exprNeedsMath(t.Index) {
				return true
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
			if c.Body != nil && g.containsMostExpr(c.Body) {
				return true
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
}

func (g *Generator) generateAbbreviation(abbr *ast.Abbreviation) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write(fmt.Sprintf("%s := ", goIdent(abbr.Name)))
	g.generateExpression(abbr.Value)
	g.write("\n")
}

func (g *Generator) generateChanDecl(decl *ast.ChanDecl) {
	goType := g.occamTypeToGo(decl.ElemType)
	if decl.IsArray {
		for _, name := range decl.Names {
			n := goIdent(name)
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("%s := make([]chan %s, ", n, goType))
			g.generateExpression(decl.Size)
			g.write(")\n")
			g.builder.WriteString(strings.Repeat("\t", g.indent))
			g.write(fmt.Sprintf("for _i := range %s { %s[_i] = make(chan %s) }\n", n, n, goType))
		}
	} else {
		for _, name := range decl.Names {
			g.writeLine(fmt.Sprintf("%s := make(chan %s)", goIdent(name), goType))
		}
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
		g.builder.WriteString(strings.Repeat("\t", g.indent))
		g.write(fmt.Sprintf("%s := make([]%s, ", n, goType))
		g.generateExpression(decl.Size)
		g.write(")\n")
	}
}

func (g *Generator) generateSend(send *ast.Send) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write(goIdent(send.Channel))
	if send.ChannelIndex != nil {
		g.write("[")
		g.generateExpression(send.ChannelIndex)
		g.write("]")
	}
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
	if recv.ChannelIndex != nil {
		var buf strings.Builder
		buf.WriteString(goIdent(recv.Channel))
		buf.WriteString("[")
		// Generate the index expression into a temporary buffer
		oldBuilder := g.builder
		g.builder = strings.Builder{}
		g.generateExpression(recv.ChannelIndex)
		buf.WriteString(g.builder.String())
		g.builder = oldBuilder
		buf.WriteString("]")
		chanRef = buf.String()
	}

	if len(recv.Variables) > 0 {
		// Sequential receive: _tmpN := <-c; x = _tmpN._0; y = _tmpN._1
		tmpName := fmt.Sprintf("_tmp%d", g.tmpCounter)
		g.tmpCounter++
		g.writeLine(fmt.Sprintf("%s := <-%s", tmpName, chanRef))
		varRef := goIdent(recv.Variable)
		if g.refParams[recv.Variable] {
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
		if g.refParams[recv.Variable] {
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
	if vr.ChannelIndex != nil {
		var buf strings.Builder
		buf.WriteString(goIdent(vr.Channel))
		buf.WriteString("[")
		oldBuilder := g.builder
		g.builder = strings.Builder{}
		g.generateExpression(vr.ChannelIndex)
		buf.WriteString(g.builder.String())
		g.builder = oldBuilder
		buf.WriteString("]")
		chanRef = buf.String()
	}
	g.writeLine(fmt.Sprintf("switch _v := (<-%s).(type) {", chanRef))
	for _, vc := range vr.Cases {
		g.writeLine(fmt.Sprintf("case _proto_%s_%s:", gProtoName, goIdent(vc.Tag)))
		g.indent++
		for i, v := range vc.Variables {
			g.writeLine(fmt.Sprintf("%s = _v._%d", goIdent(v), i))
		}
		if vc.Body != nil {
			g.generateStatement(vc.Body)
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
			if p.IsChan || p.IsChanArray {
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
			return "_proto_" + occamType
		}
		// Check if it's a record type name
		if _, ok := g.recordDefs[occamType]; ok {
			return occamType
		}
		return occamType // pass through unknown types
	}
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

	if assign.Index != nil {
		// Check if this is a record field access
		if _, ok := g.recordVars[assign.Name]; ok {
			if ident, ok := assign.Index.(*ast.Identifier); ok {
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
		// Array index: dereference if ref param
		if g.refParams[assign.Name] {
			g.write("*")
		}
		g.write(goIdent(assign.Name))
		g.write("[")
		g.generateExpression(assign.Index)
		g.write("]")
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
			if c.Guard != nil {
				g.builder.WriteString(strings.Repeat("\t", g.indent))
				g.write(fmt.Sprintf("var _alt%d chan ", i))
				// We don't know the channel type here, so use interface{}
				// Actually, we should use the same type as the original channel
				// For now, let's just reference the original channel conditionally
				g.write(fmt.Sprintf("int = nil\n")) // Assuming int for now
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
		if c.IsTimer {
			g.write("case <-time.After(time.Duration(")
			g.generateExpression(c.Deadline)
			g.write(" - int(time.Now().UnixMicro())) * time.Microsecond):\n")
		} else if c.Guard != nil {
			g.write(fmt.Sprintf("case %s = <-_alt%d:\n", goIdent(c.Variable), i))
		} else if c.ChannelIndex != nil {
			g.write(fmt.Sprintf("case %s = <-%s[", goIdent(c.Variable), goIdent(c.Channel)))
			g.generateExpression(c.ChannelIndex)
			g.write("]:\n")
		} else {
			g.write(fmt.Sprintf("case %s = <-%s:\n", goIdent(c.Variable), goIdent(c.Channel)))
		}
		g.indent++
		for _, s := range c.Body {
			g.generateStatement(s)
		}
		g.indent--
	}
	g.writeLine("}")
}

func (g *Generator) generateProcDecl(proc *ast.ProcDecl) {
	// Track reference parameters for this procedure
	oldRefParams := g.refParams
	newRefParams := make(map[string]bool)
	// Inherit parent's ref params for closure captures when nested
	if g.nestingLevel > 0 {
		for k, v := range oldRefParams {
			newRefParams[k] = v
		}
	}
	for _, p := range proc.Params {
		if !p.IsVal && !p.IsChan && !p.IsChanArray && !p.IsOpenArray {
			newRefParams[p.Name] = true
		} else {
			// Own param shadows any inherited ref param with same name
			delete(newRefParams, p.Name)
		}
		// Register chan params with protocol mappings
		if p.IsChan || p.IsChanArray {
			if _, ok := g.protocolDefs[p.ChanElemType]; ok {
				g.chanProtocols[p.Name] = p.ChanElemType
			}
		}
		// Register record-typed params
		if !p.IsChan {
			if _, ok := g.recordDefs[p.Type]; ok {
				g.recordVars[p.Name] = p.Type
			}
		}
	}
	g.refParams = newRefParams

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

	for _, stmt := range proc.Body {
		g.generateStatement(stmt)
	}

	g.nestingLevel--
	g.indent--
	g.writeLine("}")
	g.writeLine("")

	// Restore previous context
	g.refParams = oldRefParams
}

func (g *Generator) generateProcParams(params []ast.ProcParam) string {
	var parts []string
	for _, p := range params {
		var goType string
		if p.IsChanArray {
			goType = "[]" + chanDirPrefix(p.ChanDir) + g.occamTypeToGo(p.ChanElemType)
		} else if p.IsChan {
			goType = chanDirPrefix(p.ChanDir) + g.occamTypeToGo(p.ChanElemType)
		} else if p.IsOpenArray {
			goType = "[]" + g.occamTypeToGo(p.Type)
		} else if p.ArraySize != "" {
			// Fixed-size array parameter: [n]TYPE
			goType = "[" + p.ArraySize + "]" + g.occamTypeToGo(p.Type)
			if !p.IsVal {
				goType = "*" + goType
			}
		} else {
			goType = g.occamTypeToGo(p.Type)
			if !p.IsVal {
				// Non-VAL parameters are pass by reference in Occam
				goType = "*" + goType
			}
		}
		parts = append(parts, fmt.Sprintf("%s %s", goIdent(p.Name), goType))
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
		// Channels and channel arrays are already reference types, so no & needed
		if i < len(params) && !params[i].IsVal && !params[i].IsChan && !params[i].IsChanArray && !params[i].IsOpenArray && params[i].ArraySize == "" {
			g.write("&")
		}
		// Wrap string literals with []byte() when passed to []BYTE parameters
		if _, isStr := arg.(*ast.StringLiteral); isStr && i < len(params) && params[i].IsOpenArray && params[i].Type == "BYTE" {
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
}

func (g *Generator) generateFuncCallExpr(call *ast.FuncCall) {
	g.write(goIdent(call.Name))
	g.write("(")
	params := g.procSigs[call.Name]
	for i, arg := range call.Args {
		if i > 0 {
			g.write(", ")
		}
		// Wrap string literals with []byte() when passed to []BYTE parameters
		if _, isStr := arg.(*ast.StringLiteral); isStr && i < len(params) && params[i].IsOpenArray && params[i].Type == "BYTE" {
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
		if target.Index != nil {
			// Check if this is a record field access
			if _, ok := g.recordVars[target.Name]; ok {
				if ident, ok := target.Index.(*ast.Identifier); ok {
					g.write(goIdent(target.Name))
					g.write(".")
					g.write(goIdent(ident.Value))
					continue
				}
			}
			if g.refParams[target.Name] {
				g.write("(*")
				g.write(goIdent(target.Name))
				g.write(")")
			} else {
				g.write(goIdent(target.Name))
			}
			g.write("[")
			g.generateExpression(target.Index)
			g.write("]")
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
		g.write(g.occamTypeToGo(e.TargetType))
		g.write("(")
		g.generateExpression(e.Expr)
		g.write(")")
	case *ast.MostExpr:
		g.generateMostExpr(e)
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
