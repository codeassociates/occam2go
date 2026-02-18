package codegen

import "testing"

func TestE2E_ReplicatedSeq(t *testing.T) {
	// Test replicated SEQ: SEQ i = 0 FOR 5 prints 0, 1, 2, 3, 4
	occam := `SEQ i = 0 FOR 5
  print.int(i)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n1\n2\n3\n4\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeqWithExpression(t *testing.T) {
	// Test replicated SEQ with expression for count
	occam := `SEQ
  INT n:
  n := 3
  SEQ i = 0 FOR n
    print.int(i)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n1\n2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeqWithStartOffset(t *testing.T) {
	// Test replicated SEQ with non-zero start
	occam := `SEQ i = 5 FOR 3
  print.int(i)
`
	output := transpileCompileRun(t, occam)
	expected := "5\n6\n7\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeqSum(t *testing.T) {
	// Test replicated SEQ computing sum: 1+2+3+4+5 = 15
	occam := `SEQ
  INT sum:
  sum := 0
  SEQ i = 1 FOR 5
    sum := sum + i
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "15\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedPar(t *testing.T) {
	// Test replicated PAR: PAR i = 0 FOR n spawns n goroutines
	// Since PAR is concurrent, we use channels to verify all goroutines ran
	occam := `SEQ
  CHAN OF INT c:
  INT sum:
  sum := 0
  PAR
    PAR i = 0 FOR 5
      c ! i
    SEQ j = 0 FOR 5
      INT x:
      c ? x
      sum := sum + x
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	// sum should be 0+1+2+3+4 = 10
	expected := "10\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedIf(t *testing.T) {
	// Test replicated IF: find first matching element
	occam := `SEQ
  INT result:
  result := -1
  [5]INT arr:
  arr[0] := 10
  arr[1] := 20
  arr[2] := 30
  arr[3] := 40
  arr[4] := 50
  IF i = 0 FOR 5
    arr[i] = 30
      result := i
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeqStep(t *testing.T) {
	// Test replicated SEQ with STEP: SEQ i = 0 FOR 5 STEP 2 prints 0, 2, 4, 6, 8
	occam := `SEQ i = 0 FOR 5 STEP 2
  print.int(i)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n2\n4\n6\n8\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeqNegativeStep(t *testing.T) {
	// Test replicated SEQ with negative STEP: counts down
	occam := `SEQ i = 9 FOR 5 STEP -1
  print.int(i)
`
	output := transpileCompileRun(t, occam)
	expected := "9\n8\n7\n6\n5\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedParStep(t *testing.T) {
	// Test replicated PAR with STEP: verify all goroutines run with correct values
	occam := `SEQ
  CHAN OF INT c:
  INT sum:
  sum := 0
  PAR
    PAR i = 0 FOR 3 STEP 10
      c ! i
    SEQ j = 0 FOR 3
      INT x:
      c ? x
      sum := sum + x
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	// sum should be 0+10+20 = 30
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
