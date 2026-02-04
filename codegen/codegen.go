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

	// Track procedure signatures for proper pointer handling
	procSigs map[string][]ast.ProcParam
	// Track current procedure's reference parameters
	refParams map[string]bool
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

// Generate produces Go code from the AST
func (g *Generator) Generate(program *ast.Program) string {
	g.builder.Reset()
	g.needSync = false
	g.needFmt = false
	g.procSigs = make(map[string][]ast.ProcParam)
	g.refParams = make(map[string]bool)

	// First pass: collect procedure signatures and check for PAR/print
	for _, stmt := range program.Statements {
		if g.containsPar(stmt) {
			g.needSync = true
		}
		if g.containsPrint(stmt) {
			g.needFmt = true
		}
		if proc, ok := stmt.(*ast.ProcDecl); ok {
			g.procSigs[proc.Name] = proc.Params
		}
	}

	// Write package declaration
	g.writeLine("package main")
	g.writeLine("")

	// Write imports
	if g.needSync || g.needFmt {
		g.writeLine("import (")
		g.indent++
		if g.needFmt {
			g.writeLine(`"fmt"`)
		}
		if g.needSync {
			g.writeLine(`"sync"`)
		}
		g.indent--
		g.writeLine(")")
		g.writeLine("")
	}

	// Separate procedure declarations from other statements
	var procDecls []ast.Statement
	var mainStatements []ast.Statement

	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.ProcDecl); ok {
			procDecls = append(procDecls, stmt)
		} else {
			mainStatements = append(mainStatements, stmt)
		}
	}

	// Generate procedure declarations first (at package level)
	for _, stmt := range procDecls {
		g.generateStatement(stmt)
	}

	// Generate main function with other statements
	if len(mainStatements) > 0 {
		g.writeLine("func main() {")
		g.indent++
		for _, stmt := range mainStatements {
			g.generateStatement(stmt)
		}
		g.indent--
		g.writeLine("}")
	}

	return g.builder.String()
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
	case *ast.ProcDecl:
		if s.Body != nil && g.containsPar(s.Body) {
			return true
		}
	case *ast.WhileLoop:
		if s.Body != nil && g.containsPar(s.Body) {
			return true
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.Body != nil && g.containsPar(choice.Body) {
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
	case *ast.ProcDecl:
		if s.Body != nil && g.containsPrint(s.Body) {
			return true
		}
	case *ast.WhileLoop:
		if s.Body != nil && g.containsPrint(s.Body) {
			return true
		}
	case *ast.IfStatement:
		for _, choice := range s.Choices {
			if choice.Body != nil && g.containsPrint(choice.Body) {
				return true
			}
		}
	}
	return false
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
	case *ast.Skip:
		g.writeLine("// SKIP")
	case *ast.ProcDecl:
		g.generateProcDecl(s)
	case *ast.ProcCall:
		g.generateProcCall(s)
	case *ast.WhileLoop:
		g.generateWhileLoop(s)
	case *ast.IfStatement:
		g.generateIfStatement(s)
	}
}

func (g *Generator) generateVarDecl(decl *ast.VarDecl) {
	goType := g.occamTypeToGo(decl.Type)
	g.writeLine(fmt.Sprintf("var %s %s", strings.Join(decl.Names, ", "), goType))
}

func (g *Generator) generateChanDecl(decl *ast.ChanDecl) {
	goType := g.occamTypeToGo(decl.ElemType)
	for _, name := range decl.Names {
		g.writeLine(fmt.Sprintf("%s := make(chan %s)", name, goType))
	}
}

func (g *Generator) generateSend(send *ast.Send) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write(send.Channel)
	g.write(" <- ")
	g.generateExpression(send.Value)
	g.write("\n")
}

func (g *Generator) generateReceive(recv *ast.Receive) {
	g.writeLine(fmt.Sprintf("%s = <-%s", recv.Variable, recv.Channel))
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
	default:
		return occamType // pass through unknown types
	}
}

func (g *Generator) generateAssignment(assign *ast.Assignment) {
	g.builder.WriteString(strings.Repeat("\t", g.indent))
	// Dereference if assigning to a reference parameter
	if g.refParams[assign.Name] {
		g.write("*")
	}
	g.write(assign.Name)
	g.write(" = ")
	g.generateExpression(assign.Value)
	g.write("\n")
}

func (g *Generator) generateSeqBlock(seq *ast.SeqBlock) {
	// SEQ just becomes sequential Go code (Go's default)
	for _, stmt := range seq.Statements {
		g.generateStatement(stmt)
	}
}

func (g *Generator) generateParBlock(par *ast.ParBlock) {
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

func (g *Generator) generateProcDecl(proc *ast.ProcDecl) {
	// Track reference parameters for this procedure
	oldRefParams := g.refParams
	g.refParams = make(map[string]bool)
	for _, p := range proc.Params {
		if !p.IsVal {
			g.refParams[p.Name] = true
		}
	}

	// Generate function signature
	params := g.generateProcParams(proc.Params)
	g.writeLine(fmt.Sprintf("func %s(%s) {", proc.Name, params))
	g.indent++

	if proc.Body != nil {
		g.generateStatement(proc.Body)
	}

	g.indent--
	g.writeLine("}")
	g.writeLine("")

	// Restore previous context
	g.refParams = oldRefParams
}

func (g *Generator) generateProcParams(params []ast.ProcParam) string {
	var parts []string
	for _, p := range params {
		goType := g.occamTypeToGo(p.Type)
		if !p.IsVal {
			// Non-VAL parameters are pass by reference in Occam
			goType = "*" + goType
		}
		parts = append(parts, fmt.Sprintf("%s %s", p.Name, goType))
	}
	return strings.Join(parts, ", ")
}

func (g *Generator) generateProcCall(call *ast.ProcCall) {
	// Handle built-in print procedures
	if printBuiltins[call.Name] {
		g.generatePrintCall(call)
		return
	}

	g.builder.WriteString(strings.Repeat("\t", g.indent))
	g.write(call.Name)
	g.write("(")

	// Look up procedure signature to determine which args need address-of
	params := g.procSigs[call.Name]

	for i, arg := range call.Args {
		if i > 0 {
			g.write(", ")
		}
		// If this parameter is not VAL (i.e., pass by reference), take address
		if i < len(params) && !params[i].IsVal {
			g.write("&")
		}
		g.generateExpression(arg)
	}
	g.write(")")
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

	if loop.Body != nil {
		g.generateStatement(loop.Body)
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) generateIfStatement(stmt *ast.IfStatement) {
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

		if choice.Body != nil {
			g.generateStatement(choice.Body)
		}

		g.indent--
	}
	g.writeLine("}")
}

func (g *Generator) generateExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		g.write(e.Value)
	case *ast.IntegerLiteral:
		g.write(fmt.Sprintf("%d", e.Value))
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
	case *ast.ParenExpr:
		g.write("(")
		g.generateExpression(e.Expr)
		g.write(")")
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
	if op == "!" {
		// Go's NOT doesn't need space
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
	default:
		return op // +, -, *, /, <, >, <=, >= are the same
	}
}
