package sputnik

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Start(cntr ServerConnector) {

	var confFolder string
	flag.StringVar(&confFolder, "cf", "", "Path of folder with config files")
	flag.Parse()

	if len(confFolder) == 0 {
		fmt.Fprintln(os.Stderr, fmt.Errorf("-cf <path of config folder> - was not set!!!"))
		os.Exit(1)
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
	kill ShootDown
	// Signalling channel
	done chan struct{}
}

func StartRunner(confFolder string, cntr ServerConnector) (*Runner, error) {
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
	cfact     ConfFactory
	cnt       ServerConnector
	appBlocks []BlockDescriptor
}

func prepare(confFolder string, cntr ServerConnector) (*runnerInfo, error) {
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

	sp, err := NewSputnik(
		WithAppBlocks(ri.appBlocks),
		WithConfFactory(ri.cfact),
		WithConnector(ri.cnt, brokerCheckTimeOut),
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

	go func(l Launch, done chan struct{}) {
		l()
		close(done)
	}(launch, rnr.done)

	return nil
}

func ReadAppBlocks(confFolder string) ([]BlockDescriptor, error) {
	fPath := filepath.Join(confFolder, "blocks.json")

	blocksRaw, err := os.ReadFile(fPath)
	if err != nil {
		return nil, err
	}

	var result []BlockDescriptor

	json.Unmarshal([]byte(blocksRaw), &result)

	return result, nil
}
