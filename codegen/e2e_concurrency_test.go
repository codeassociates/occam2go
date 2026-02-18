package codegen

import "testing"

func TestE2E_PAR(t *testing.T) {
	// Test that PAR executes both branches
	// We can't guarantee order, but both should run
	occam := `SEQ
  INT x, y:
  PAR
    x := 10
    y := 20
  print.int(x + y)
`
	output := transpileCompileRun(t, occam)
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Channel(t *testing.T) {
	// Test basic channel communication between parallel processes
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  PAR
    c ! 42
    c ? result
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChannelExpression(t *testing.T) {
	// Test sending an expression over a channel
	occam := `SEQ
  CHAN OF INT c:
  INT x, result:
  x := 10
  PAR
    c ! x * 2
    c ? result
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChannelPingPong(t *testing.T) {
	// Test two-way communication: send a value, double it, send back
	occam := `SEQ
  CHAN OF INT request:
  CHAN OF INT response:
  INT result:
  PAR
    SEQ
      request ! 21
      response ? result
    SEQ
      INT x:
      request ? x
      response ! x * 2
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltBasic(t *testing.T) {
	// Test basic ALT: select from first ready channel
	occam := `SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  PAR
    c1 ! 42
    ALT
      c1 ? result
        print.int(result)
      c2 ? result
        print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltSecondChannel(t *testing.T) {
	// Test ALT selecting from second channel
	occam := `SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  PAR
    c2 ! 99
    ALT
      c1 ? result
        print.int(result)
      c2 ? result
        print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltWithBody(t *testing.T) {
	// Test ALT with computation in body
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  PAR
    c ! 10
    ALT
      c ? result
        SEQ
          result := result * 2
          print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_TimerRead(t *testing.T) {
	// Test reading a timer: value should be positive (microseconds since epoch)
	occam := `SEQ
  TIMER tim:
  INT t:
  tim ? t
  IF
    t > 0
      print.int(1)
    TRUE
      print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_TimerAltTimeout(t *testing.T) {
	// Test ALT with timer timeout: no channel is ready, so timer fires
	occam := `SEQ
  TIMER tim:
  INT t:
  tim ? t
  CHAN OF INT c:
  INT result:
  result := 0
  ALT
    c ? result
      result := 1
    tim ? AFTER (t + 1000)
      result := 2
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChanParam(t *testing.T) {
	occam := `PROC sender(CHAN OF INT output)
  output ! 42

SEQ
  CHAN OF INT c:
  PAR
    sender(c)
    SEQ
      INT x:
      c ? x
      print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChanParamTwoWay(t *testing.T) {
	occam := `PROC doubler(CHAN OF INT input, CHAN OF INT output)
  SEQ
    INT x:
    input ? x
    output ! x * 2

SEQ
  CHAN OF INT inCh:
  CHAN OF INT outCh:
  PAR
    doubler(inCh, outCh)
    SEQ
      inCh ! 21
      INT result:
      outCh ? result
      print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChanDirParam(t *testing.T) {
	occam := `PROC producer(CHAN OF INT output!)
  output ! 42

PROC consumer(CHAN OF INT input?)
  SEQ
    INT x:
    input ? x
    print.int(x)

SEQ
  CHAN OF INT c:
  PAR
    producer(c)
    consumer(c)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2EChanShorthand(t *testing.T) {
	occam := `SEQ
  CHAN INT c:
  INT result:
  PAR
    c ! 42
    c ? result
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
