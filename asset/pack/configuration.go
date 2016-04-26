package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type InputConfig struct {
	Path      string
	Recursive bool
}

type Config struct {
	Package     string
	Tags        string
	Input       []InputConfig
	Output      string
	FSPrefix    string
	FSName      string
	StripPrefix string
	NoMemCopy   bool
	NoCompress  bool
	Debug       bool
	Dev         bool
	NoMetadata  bool
	Mode        uint
	ModTime     int64
	Ignore      []*regexp.Regexp
}

func NewConfig() *Config {
	c := new(Config)
	c.Package = "resources"
	c.FSName = "ResourceFS"
	c.FSPrefix = ""
	c.NoMemCopy = false
	c.NoCompress = false
	c.Debug = false
	c.Output = "./packed.go"
	c.Ignore = make([]*regexp.Regexp, 0)
	return c
}

func (c *Config) validate() error {
	if len(c.Package) == 0 {
		return fmt.Errorf("Missing package name")
	}

	for _, input := range c.Input {
		_, err := os.Lstat(input.Path)
		if err != nil {
			return fmt.Errorf("Failed to stat input path '%s': %v", input.Path, err)
		}
	}

	if len(c.Output) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("Unable to determine current working directory.")
		}

		c.Output = filepath.Join(cwd, "packed.go")
	}

	stat, err := os.Lstat(c.Output)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Output path: %v", err)
		}

		// File does not exist. This is fine, just make
		// sure the directory it is to be in exists.
		dir, _ := filepath.Split(c.Output)
		if dir != "" {
			err = os.MkdirAll(dir, 0744)

			if err != nil {
				return fmt.Errorf("Create output directory: %v", err)
			}
		}
	}

	if stat != nil && stat.IsDir() {
		return fmt.Errorf("Output path is a directory.")
	}

	return nil
}
