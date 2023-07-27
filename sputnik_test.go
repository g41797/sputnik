package sputnik_test

import (
	"testing"
	"time"

	"github.com/g41797/sputnik"
)

func TestPrepare(t *testing.T) {

	tb := NewTestBlocks()

	dsp := dumbSputnik(tb)

	_, kill, err := dsp.Prepare()

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	tb.attachQueue()

	time.Sleep(time.Second)

	kill()

	return
}

func TestFinisher(t *testing.T) {

	tb := NewTestBlocks()

	dsp := dumbSputnik(tb)

	launch, kill, err := dsp.Prepare()

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	tb.attachQueue()
	tb.launch = launch
	tb.kill = kill

	tb.run()

	time.Sleep(1 * time.Second)

	// Simulate SIGINT
	ok := tb.sendTo("finisher", make(sputnik.Msg))
	if !ok {
		t.Errorf("send to finisher failed")
	}

	tb.kill()

	<-tb.done

	return
}

func TestRun(t *testing.T) {

	tb := NewTestBlocks()

	dsp := dumbSputnik(tb)

	launch, kill, err := dsp.Prepare()

	if err != nil {
		t.Errorf("Prepare error %v", err)
	}

	tb.attachQueue()
	tb.launch = launch
	tb.kill = kill

	tb.run()

	time.Sleep(1 * time.Second)

	// Simulate ServerConnect
	tb.conntr.SetState(true)
	if !tb.expect(3, "serverConnected") {
		t.Errorf("Wrong processing of serverconnected")
	}

	// Simulate ServerDisconnect
	tb.conntr.SetState(false)
	if !tb.expect(3, "serverDisconnected") {
		t.Errorf("Wrong processing of serverDisconnected")
	}

	tb.kill()

	<-tb.done

	return
}
