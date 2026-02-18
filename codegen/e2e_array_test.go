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
