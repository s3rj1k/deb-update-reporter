package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Email struct {
		Header struct {
			From    string   `yaml:"From"`
			ReplyTo string   `yaml:"ReplyTo"`
			Subject string   `yaml:"Subject"`
			To      []string `yaml:"To"`
		} `yaml:"Header"`
		SMTP struct {
			Address  string `yaml:"Address"`
			Password string `yaml:"Password"`
			Port     int    `yaml:"Port"`
			Server   string `yaml:"Server"`
		} `yaml:"SMTP"`
	} `yaml:"Email"`
	Repo []struct {
		Name     string   `yaml:"Name"`
		URL      []string `yaml:"URL"`
		Packages []*struct {
			Name             string `yaml:"Name"`
			VersionNewerThan string `yaml:"VersionNewerThan"`
		} `yaml:"Packages"`
	} `yaml:"Repo"`
}

func getConfig(path string) (config, error) {
	var c config

	// read file from disk
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	// decode yaml
	if err = yaml.Unmarshal(f, &c); err != nil {
		return config{}, err
	}

	return c, nil
}

func saveConfig(c config, path string) error {
	// encode config
	newConfig, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// write file to disk
	return ioutil.WriteFile(path, newConfig, 0644)
}
