package sputnik_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/g41797/sputnik"
)

type TestConf struct {
	FIRSTSTRING string `mapstructure:"FIRSTSTRING"`
	SECONDINT   int    `mapstructure:"SECONDINT"`
}

func TestJSON(t *testing.T) {

	confFolderPath := "./_conf_test/"

	cfact := sputnik.ConfigFactory(confFolderPath)

	expected := defaults()

	var tf TestConf

	err := cfact("example", &tf)

	compare(t, err, tf, expected)
}

func TestENV(t *testing.T) {

	confFolderPath := "./_conf_test/"

	cfact := sputnik.ConfigFactory(confFolderPath)

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
