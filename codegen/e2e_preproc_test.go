package codegen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestE2E_IncludeConstants(t *testing.T) {
	// Test #INCLUDE of a constants file and using the constant in a program
	tmpDir := t.TempDir()

	// Create a constants file with a function
	constsContent := "INT FUNCTION magic(VAL INT n)\n  IS n * 2\n"
	os.WriteFile(filepath.Join(tmpDir, "consts.inc"), []byte(constsContent), 0644)

	// Create main file that includes the constants
	mainContent := `#INCLUDE "consts.inc"
SEQ
  print.int(magic(21))
`
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	output := transpileCompileRunFromFile(t, mainFile, nil)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IfFalseExcludes(t *testing.T) {
	// Test that #IF FALSE excludes code from compilation
	tmpDir := t.TempDir()

	mainContent := `SEQ
  print.int(1)
#IF FALSE
  THIS IS INVALID OCCAM AND SHOULD NOT BE PARSED
#ENDIF
`
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	output := transpileCompileRunFromFile(t, mainFile, nil)
	expected := "1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IncludeGuardPreventsDouble(t *testing.T) {
	// Test that include guards prevent double-inclusion of declarations
	tmpDir := t.TempDir()

	// Create a guarded module with a function
	modContent := "#IF NOT (DEFINED (TEST.MODULE))\n#DEFINE TEST.MODULE\nINT FUNCTION doubled(VAL INT x)\n  IS x * 2\n#ENDIF\n"
	os.WriteFile(filepath.Join(tmpDir, "test.module"), []byte(modContent), 0644)

	// Include it twice — should work thanks to guards
	mainContent := `#INCLUDE "test.module"
#INCLUDE "test.module"
SEQ
  print.int(doubled(21))
`
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	output := transpileCompileRunFromFile(t, mainFile, nil)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
