package main

import (
	"io/ioutil"

	sendmail "github.com/s3rj1k/go-smtp-html-helper"
	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Email struct {
		Headers sendmail.HeadersConfig `json:"Headers" yaml:"Headers"`
		SMTP    sendmail.SMTPConfig    `json:"SMTP" yaml:"SMTP"`
		To      []string               `json:"To" yaml:"To"`
	} `json:"Email" yaml:"Email"`
	Repo []struct {
		Name string   `json:"Name" yaml:"Name"`
		URL  []string `json:"URL" yaml:"URL"`

		Packages []*struct {
			Name             string `json:"Name" yaml:"Name"`
			VersionNewerThan string `json:"VersionNewerThan" yaml:"VersionNewerThan"`
		} `json:"Packages" yaml:"Packages"`
	} `json:"Repo" yaml:"Repo"`
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
