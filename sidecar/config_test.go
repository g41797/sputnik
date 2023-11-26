package sidecar_test

import (
	"embed"
	"os"
	"strconv"
	"testing"

	"github.com/g41797/sputnik/sidecar"
)

//go:embed _conf_test
var cnffiles embed.FS

type TestConf struct {
	FIRSTSTRING string `mapstructure:"FIRSTSTRING"`
	SECONDINT   int    `mapstructure:"SECONDINT"`
}

func TestJSON(t *testing.T) {

	confFolderPath := "./_conf_test/"

	cfact := sidecar.ConfigFactory(confFolderPath)

	expected := defaults()

	var tf TestConf

	err := cfact("example", &tf)

	compare(t, err, tf, expected)
}

func TestENV(t *testing.T) {

	confFolderPath := "./_conf_test/"

	cfact := sidecar.ConfigFactory(confFolderPath)

	expected := defaults()
	expected.SECONDINT = 12345

	os.Setenv("EXAMPLE_FIRSTSTRING", expected.FIRSTSTRING)

	os.Setenv("EXAMPLE_SECONDINT", strconv.Itoa((expected.SECONDINT)))

	secIntString, ok := os.LookupEnv("EXAMPLE_SECONDINT")

	if !ok {
		t.Errorf("Wrong working with environment")
	}

	if secIntString != "12345" {
		t.Errorf("Wrong test")
	}

	var tf TestConf

	err := cfact("example", &tf)

	compare(t, err, tf, expected)
}

func defaults() TestConf {
	return TestConf{"1.0.3", 300}
}

func compare(t *testing.T, getErr error, actual TestConf, expected TestConf) {

	if getErr != nil {
		t.Errorf("unmarshal error %v", getErr)
	}

	if actual.FIRSTSTRING != expected.FIRSTSTRING {
		t.Errorf("Expected %s Actual %s", expected.FIRSTSTRING, actual.FIRSTSTRING)
	}

	if actual.SECONDINT != expected.SECONDINT {
		t.Errorf("Expected %d Actual %d", expected.SECONDINT, actual.SECONDINT)
	}
}

func TestUseTemporaryFolderForConfiguration(t *testing.T) {
	cleanup, err := sidecar.UseEmbeddedConfiguration(&cnffiles)
	if err != nil {
		t.Errorf("UseEmbeddedConfiguration error:%v", err)
		return
	}
	defer cleanup()

	dir, err := sidecar.ConfFolder()

	if err != nil {
		t.Errorf("UseEmbeddedConfiguration error:%v", err)
		return
	}

	fileInfo, err := os.Stat(dir)
	if err != nil {
		t.Errorf("UseEmbeddedConfiguration error:%v", err)
	}

	if !fileInfo.IsDir() {
		t.Errorf("%s is not directory", dir)
	}

	url := "https://memphisdev.github.io/memphis-docker/docker-compose-dev.yml"

	err = sidecar.LoadDockerComposeFile("https://memphisdev.github.io/memphis-docker/docker-compose-dev.yml")
	if err != nil {
		t.Errorf("LoadDockerComposeFile from %s error:%v", url, err)
	}

}
