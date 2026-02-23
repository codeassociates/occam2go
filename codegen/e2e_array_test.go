package codegen

import "testing"

func TestE2E_ArrayBasic(t *testing.T) {
	// Test basic array: declare, store, load
	occam := `SEQ
  [5]INT arr:
  arr[0] := 42
  print.int(arr[0])
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArrayWithLoop(t *testing.T) {
	// Test filling array with replicated SEQ and printing all elements
	occam := `SEQ
  [5]INT arr:
  SEQ i = 0 FOR 5
    arr[i] := i * 10
  SEQ i = 0 FOR 5
    print.int(arr[i])
`
	output := transpileCompileRun(t, occam)
	expected := "0\n10\n20\n30\n40\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArraySum(t *testing.T) {
	// Test computing sum of array elements
	occam := `SEQ
  [4]INT arr:
  arr[0] := 10
  arr[1] := 20
  arr[2] := 30
  arr[3] := 40
  INT sum:
  sum := 0
  SEQ i = 0 FOR 4
    sum := sum + arr[i]
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "100\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArrayExpressionIndex(t *testing.T) {
	// Test using variable and expression as array index
	occam := `SEQ
  [3]INT arr:
  INT idx:
  arr[0] := 100
  arr[1] := 200
  arr[2] := 300
  idx := 1
  print.int(arr[idx])
  print.int(arr[idx + 1])
`
	output := transpileCompileRun(t, occam)
	expected := "200\n300\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChanArrayBasic(t *testing.T) {
	// Declare channel array, use in replicated PAR to send/receive
	occam := `SEQ
  [3]CHAN OF INT cs:
  INT sum:
  sum := 0
  PAR
    PAR i = 0 FOR 3
      cs[i] ! (i + 1) * 10
    SEQ i = 0 FOR 3
      INT x:
      cs[i] ? x
      sum := sum + x
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "60\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChanArrayWithProc(t *testing.T) {
	// Pass channel array to a PROC
	occam := `PROC sender([]CHAN OF INT cs, VAL INT n)
  SEQ i = 0 FOR n
    cs[i] ! (i + 1) * 100

SEQ
  [3]CHAN OF INT cs:
  INT sum:
  sum := 0
  PAR
    sender(cs, 3)
    SEQ i = 0 FOR 3
      INT x:
      cs[i] ? x
      sum := sum + x
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "600\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChanArrayAlt(t *testing.T) {
	// Use channel array in ALT
	occam := `SEQ
  [2]CHAN OF INT cs:
  INT result:
  result := 0
  PAR
    cs[0] ! 42
    ALT
      cs[0] ? result
        print.int(result)
      cs[1] ? result
        print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SizeArray(t *testing.T) {
	occam := `SEQ
  [5]INT arr:
  INT n:
  n := SIZE arr
  print.int(n)
`
	output := transpileCompileRun(t, occam)
	expected := "5\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SizeString(t *testing.T) {
	occam := `SEQ
  INT n:
  n := SIZE "hello"
  print.int(n)
`
	output := transpileCompileRun(t, occam)
	expected := "5\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SliceAsArg(t *testing.T) {
	// Pass an array slice to a PROC expecting an open array param
	occam := `PROC printarray(VAL []INT arr)
  SEQ i = 0 FOR SIZE arr
    print.int(arr[i])
SEQ
  [5]INT nums:
  SEQ i = 0 FOR 5
    nums[i] := (i + 1) * 10
  printarray([nums FROM 1 FOR 3])
`
	output := transpileCompileRun(t, occam)
	expected := "20\n30\n40\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SliceAssignment(t *testing.T) {
	// Copy elements within an array using slice assignment
	occam := `SEQ
  [6]INT arr:
  SEQ i = 0 FOR 6
    arr[i] := i + 1
  [arr FROM 3 FOR 3] := [arr FROM 0 FOR 3]
  SEQ i = 0 FOR 6
    print.int(arr[i])
`
	output := transpileCompileRun(t, occam)
	expected := "1\n2\n3\n1\n2\n3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SliceSize(t *testing.T) {
	// SIZE of a slice expression
	occam := `SEQ
  [10]INT arr:
  INT n:
  n := SIZE [arr FROM 2 FOR 5]
  print.int(n)
`
	output := transpileCompileRun(t, occam)
	expected := "5\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SliceFromZero(t *testing.T) {
	// Slice starting from index 0 passed to a VAL open array proc
	occam := `PROC printsum(VAL []INT arr)
  SEQ
    INT total:
    total := 0
    SEQ i = 0 FOR SIZE arr
      total := total + arr[i]
    print.int(total)
SEQ
  [5]INT arr:
  SEQ i = 0 FOR 5
    arr[i] := i + 1
  printsum([arr FROM 0 FOR 3])
`
	output := transpileCompileRun(t, occam)
	expected := "6\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_OpenArrayParam(t *testing.T) {
	occam := `PROC printarray(VAL []INT arr)
  SEQ i = 0 FOR SIZE arr
    print.int(arr[i])
SEQ
  [3]INT nums:
  SEQ
    nums[0] := 10
    nums[1] := 20
    nums[2] := 30
  printarray(nums)
`
	output := transpileCompileRun(t, occam)
	expected := "10\n20\n30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiAssignmentSwap(t *testing.T) {
	occam := `SEQ
  [2]INT arr:
  SEQ
    arr[0] := 10
    arr[1] := 20
    arr[0], arr[1] := arr[1], arr[0]
    print.int(arr[0])
    print.int(arr[1])
`
	output := transpileCompileRun(t, occam)
	expected := "20\n10\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiAssignmentMixed(t *testing.T) {
	occam := `SEQ
  INT a:
  [3]INT arr:
  SEQ
    arr[0] := 99
    a, arr[1] := arr[0], 42
    print.int(a)
    print.int(arr[1])
`
	output := transpileCompileRun(t, occam)
	expected := "99\n42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiAssignmentValues(t *testing.T) {
	occam := `SEQ
  INT a, b, c:
  a, b, c := 10, 20, 30
  print.int(a)
  print.int(b)
  print.int(c)
`
	output := transpileCompileRun(t, occam)
	expected := "10\n20\n30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChanArrayDirParam(t *testing.T) {
	// Channel arrays passed to direction-annotated params must compile
	// (Go slices are not covariant, so direction is dropped for array params)
	occam := `PROC sender([]CHAN OF INT out!)
  SEQ i = 0 FOR SIZE out
    out[i] ! i
:
PROC receiver([]CHAN OF INT in?)
  SEQ i = 0 FOR SIZE in
    INT v:
    SEQ
      in[i] ? v
      print.int(v)
:
SEQ
  [3]CHAN OF INT cs:
  PAR
    sender(cs)
    receiver(cs)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n1\n2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiDimArray(t *testing.T) {
	// 2D array: declare, fill with SEQ loops, read back
	occam := `SEQ
  [3][4]INT grid:
  SEQ i = 0 FOR 3
    SEQ j = 0 FOR 4
      grid[i][j] := (i * 10) + j
  SEQ i = 0 FOR 3
    SEQ j = 0 FOR 4
      print.int(grid[i][j])
`
	output := transpileCompileRun(t, occam)
	expected := "0\n1\n2\n3\n10\n11\n12\n13\n20\n21\n22\n23\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiDimChanArray(t *testing.T) {
	// 2D channel array: send/receive
	occam := `SEQ
  [2][3]CHAN OF INT cs:
  INT sum:
  sum := 0
  PAR
    SEQ i = 0 FOR 2
      SEQ j = 0 FOR 3
        cs[i][j] ! (i * 10) + j
    SEQ
      SEQ i = 0 FOR 2
        SEQ j = 0 FOR 3
          INT x:
          cs[i][j] ? x
          sum := sum + x
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	// sum = 0+1+2+10+11+12 = 36
	expected := "36\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiDimChanArrayWithProc(t *testing.T) {
	// Pass 2D channel array to a PROC
	occam := `PROC fill([][]CHAN OF INT grid, VAL INT rows, VAL INT cols)
  SEQ i = 0 FOR rows
    SEQ j = 0 FOR cols
      grid[i][j] ! (i * 100) + j
:
SEQ
  [2][3]CHAN OF INT cs:
  INT sum:
  sum := 0
  PAR
    fill(cs, 2, 3)
    SEQ i = 0 FOR 2
      SEQ j = 0 FOR 3
        INT v:
        cs[i][j] ? v
        sum := sum + v
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	// sum = 0+1+2+100+101+102 = 306
	expected := "306\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
