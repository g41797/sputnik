package sidecar

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/g41797/gonfig"
	"github.com/g41797/sputnik"
)

// ConfigFactory returns implementation of sputnik.ConfFactory based on github.com/tkanos/gonfig
// - JSON format of config files
// - Automatic matching of environment variables
// - Env. variable for configuration "example" and key "kname" should be set in environment as "EXAMPLE_KNAME"
// - Value in environment automatically overrides value from the file
// - Temporary used github.com/g41797/gonfig  (till merge of PR)
func ConfigFactory(cfPath string) sputnik.ConfFactory {
	cnf := newConfig(cfPath)
	return cnf.unmarshal
}

type config struct {
	confPath string
}

func newConfig(cfPath string) *config {
	return &config{confPath: cfPath}
}

func (conf *config) unmarshal(confName string, result any) error {
	fPath := filepath.Join(conf.confPath, strings.ToLower(confName))
	fPath += ".json"
	_, err := os.Open(fPath)
	if err != nil {
		return err
	}
	err = gonfig.GetConf(fPath, result)
	return err
}

// Command line of sidecar : <exe name> --cf <path of config folder>
// ConfFolder() return the value of 'cf' flag
func ConfFolder() (confFolder string, err error) {
	fName := "cf"
	fVal := flag.Lookup(fName)
	if fVal == nil {
		flag.StringVar(&confFolder, fName, "", "Path of folder with config files")
		flag.Parse()
	} else {
		confFolder = fVal.Value.String()
	}

	if len(confFolder) == 0 {
		err = fmt.Errorf("-cf <path of config folder> - was not set")
		return "", err
	}

	info, err := os.Stat(confFolder)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not the folder", confFolder)
	}

	return confFolder, nil
}

// UseEmbeddedConfiguration creates temporary directory and copies configuration files
// embedded within package.
// For success returns cleanUp function for removing created directory
// Use this function in main function if you don't supply path to config directory in
// command line.
// Example:
//
// import (
//
//		"embed"
//	 .......
//
// )
//
// go:embed configdir
// var cnffiles embed.FS
//
//	func main() {
//		cleanup, err := sidecar.UseEmbeddedConfiguration(&cnffiles)
//		if err != nil {
//			return err
//		}
//		defer cleanup()
//		sidecar.Start(new(adapter.BrokerConnector))
//	}
func UseEmbeddedConfiguration(efs *embed.FS) (cleanUp func(), err error) {

	tmpDir, err := os.MkdirTemp("", "config")
	if err != nil {
		return nil, err
	}

	cleanUp = func() {
		os.RemoveAll(tmpDir)
	}

	files, err := getAllFilenames(efs)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		dstfp := path.Join(tmpDir, filepath.Base(file))
		if err = copyFile(file, dstfp); err != nil {
			cleanUp()
			return nil, err
		}
	}
	flag.String("cf", tmpDir, "path to configuration folder")
	return cleanUp, nil
}

// https://gist.github.com/clarkmcc/1fdab4472283bb68464d066d6b4169bc?permalink_comment_id=4405804#gistcomment-4405804
func getAllFilenames(efs *embed.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

// https://blog.depa.do/post/copy-files-and-directories-in-go
func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}
