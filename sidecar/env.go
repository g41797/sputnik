package sidecar

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/g41797/godotenv"
)

// Loads all *.env file from configuration folder of the process.
// Overwrites non-existing environment variables.
// Supports "prefixed" environment variables
// (https://github.com/g41797/gonfig#using-prefixes-for-environment-variables-name)
func LoadEnv() error {
	cf, err := ConfFolder()
	if err != nil {
		return err
	}

	envf, err := WalkDir(cf, []string{"env"})
	if err != nil {
		return err
	}

	if len(envf) == 0 {
		return nil
	}

	for _, ef := range envf {
		if err := godotenv.Load(ef); err != nil {
			return err
		}
	}

	return nil
}

// Finds all the files matching a particular suffix in all the directories
// https://stackoverflow.com/questions/70537979/how-to-efficiently-find-all-the-files-matching-a-particular-suffix-in-all-the-di
func WalkDir(root string, exts []string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		for _, s := range exts {
			if strings.HasSuffix(path, "."+s) {
				files = append(files, path)
				return nil
			}
		}

		return nil
	})
	return files, err
}
