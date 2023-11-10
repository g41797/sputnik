package sidecar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/g41797/sputnik"
)

func Start(cntr sputnik.ServerConnector) {

	confFolder, err := ConfFolder()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	rnr, err := StartRunner(confFolder, cntr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rnr.Wait()
	return

}

const brokerCheckTimeOut = time.Second

type Runner struct {
	// ShootDown
	kill sputnik.ShootDown
	// Signalling channel
	done chan struct{}
}

func StartRunner(confFolder string, cntr sputnik.ServerConnector) (*Runner, error) {
	info, err := prepare(confFolder, cntr)
	if err != nil {
		return nil, err
	}
	rnr := new(Runner)
	err = rnr.Start(info)
	if err != nil {
		return nil, err
	}
	return rnr, nil
}

func (rnr *Runner) Stop() {
	if rnr == nil {
		return
	}
	if rnr.kill == nil {
		return
	}

	select {
	case <-rnr.done:
		return
	default:
	}

	rnr.kill()

	return
}

func (rnr *Runner) Wait() {
	if rnr == nil {
		return
	}

	<-rnr.done

	return
}

type runnerInfo struct {
	cfact     sputnik.ConfFactory
	cnt       sputnik.ServerConnector
	appBlocks []sputnik.BlockDescriptor
}

func prepare(confFolder string, cntr sputnik.ServerConnector) (*runnerInfo, error) {
	info, err := os.Stat(confFolder)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not the folder", confFolder)
	}

	ri := runnerInfo{}

	ri.cfact = ConfigFactory(confFolder)

	ri.cnt = cntr

	ri.appBlocks, err = ReadAppBlocks(confFolder)

	return &ri, err
}

func (rnr *Runner) Start(ri *runnerInfo) error {

	sp, err := sputnik.NewSputnik(
		sputnik.WithAppBlocks(ri.appBlocks),
		sputnik.WithConfFactory(ri.cfact),
		sputnik.WithConnector(ri.cnt, brokerCheckTimeOut),
	)

	if err != nil {
		return err
	}

	launch, kill, err := sp.Prepare()

	if err != nil {
		return err
	}
	rnr.kill = kill
	rnr.done = make(chan struct{})

	go func(l sputnik.Launch, done chan struct{}) {
		l()
		close(done)
	}(launch, rnr.done)

	return nil
}

func ReadAppBlocks(confFolder string) ([]sputnik.BlockDescriptor, error) {
	fPath := filepath.Join(confFolder, "blocks.json")

	blocksRaw, err := os.ReadFile(fPath)
	if err != nil {
		return nil, err
	}

	var result []sputnik.BlockDescriptor

	json.Unmarshal([]byte(blocksRaw), &result)

	return result, nil
}
