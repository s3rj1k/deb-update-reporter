package main

import (
	"os"

	yaml "gopkg.in/yaml.v3"
)

type config struct {
	Email struct {
		Headers HeadersConfig `json:"Headers" yaml:"Headers"`
		SMTP    SMTPConfig    `json:"SMTP" yaml:"SMTP"`
		To      []string      `json:"To" yaml:"To"`
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

	f, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	if err = yaml.Unmarshal(f, &c); err != nil {
		return config{}, err
	}

	return c, nil
}

func saveConfig(c config, path string) error {
	newConfig, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, newConfig, 0644)
}
