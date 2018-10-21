package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/s3rj1k/go-smtp-html-helper"
	yaml "gopkg.in/yaml.v2"
	"pault.ag/go/debian/control"
	"pault.ag/go/debian/version"
)

var (
	cmdDryRun       bool
	cmdUpdateConfig bool
	cmdConfigPath   string
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

func init() {
	// set simple log output
	log.SetFlags(0)
	// cmd flags
	flag.StringVar(&cmdConfigPath, "config-path", "config.yaml", "path to config file")
	flag.BoolVar(&cmdUpdateConfig, "update-config", true, "save config before exit")
	flag.BoolVar(&cmdDryRun, "dry-run", false, "print to console instead of sending email")
	flag.Parse()
}

func getPackagesBinaryIndexURL(urls []string) ([]control.BinaryIndex, error) {

	out := make([]control.BinaryIndex, 0)

	// set http client config
	var client = &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 30 * time.Second,
		},
	}

	for _, url := range urls {

		// get data from remote URL
		response, err := client.Get(url)
		if err != nil {
			return []control.BinaryIndex{}, err
		}
		defer response.Body.Close()

		// error 404
		if response.StatusCode == 404 {
			return []control.BinaryIndex{}, fmt.Errorf("remote URL not found: %s", url)
		}

		// set default reader
		var reader io.ReadCloser

		// Check that the server actually sent compressed data
		switch response.Header.Get("Content-Type") {

		case "gzip", "application/x-gzip", "application/gzip":
			// decode gzip
			reader, err = gzip.NewReader(response.Body)
			if err != nil {
				return []control.BinaryIndex{}, errors.Wrapf(err, "failed to parse URL=%s", url)
			}
			defer reader.Close()

			// parse binary index
			index, err := control.ParseBinaryIndex(bufio.NewReader(reader))
			if err != nil {
				return []control.BinaryIndex{}, errors.Wrapf(err, "failed to parse URL=%s", url)
			}

			// append to output
			out = append(out, index...)

		case "text/plain":
			// parse binary index
			index, err := control.ParseBinaryIndex(bufio.NewReader(response.Body))
			if err != nil {
				return []control.BinaryIndex{}, errors.Wrapf(err, "failed to parse URL=%s", url)
			}

			// append to output
			out = append(out, index...)

		case "application/octet-stream":
			switch {
			// decode gzip by URL suffix
			case strings.HasSuffix(url, ".gz"):
				// decode gzip
				reader, err = gzip.NewReader(response.Body)
				if err != nil {
					return []control.BinaryIndex{}, errors.Wrapf(err, "failed to parse URL=%s", url)
				}
				defer reader.Close()

				// parse binary index
				index, err := control.ParseBinaryIndex(bufio.NewReader(reader))
				if err != nil {
					return []control.BinaryIndex{}, errors.Wrapf(err, "failed to parse URL=%s", url)
				}

				// append to output
				out = append(out, index...)

			default:
				// parse binary index
				index, err := control.ParseBinaryIndex(bufio.NewReader(response.Body))
				if err != nil {
					return []control.BinaryIndex{}, errors.Wrapf(err, "failed to parse URL=%s", url)
				}

				// append to output
				out = append(out, index...)
			}

		default:
			// parse binary index
			index, err := control.ParseBinaryIndex(bufio.NewReader(response.Body))
			if err != nil {
				return []control.BinaryIndex{}, errors.Wrapf(err, "failed to parse URL=%s", url)
			}

			// append to output
			out = append(out, index...)
		}
	}

	return out, nil
}

func getConfig(path string) (config, error) {

	var c config

	// read file from disk
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	// decode yaml
	err = yaml.Unmarshal(f, &c)
	if err != nil {
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
	err = ioutil.WriteFile(path, newConfig, 644)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	// output map
	m := make(map[string]map[string]version.Version)

	// read config
	c, err := getConfig(cmdConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	// loop-over repos inside config
	for _, repo := range c.Repo {

		// parse repo db
		bIndex, err := getPackagesBinaryIndexURL(repo.URL)
		if err != nil {
			log.Fatal(err)
		}

		// loop-over packages in repo db
		for _, pkgIndex := range bIndex {
			// loop-over config packages
			for _, pkg := range repo.Packages {
				// check if package in config
				if strings.EqualFold(pkg.Name, pkgIndex.Package) {
					// parse version string
					ver, err := version.Parse(pkg.VersionNewerThan)
					if err != nil {
						log.Fatal(err)
					}
					// compare version
					if version.Compare(pkgIndex.Version, ver) > 0 {
						// update config
						pkg.VersionNewerThan = pkgIndex.Version.String()
						// update output map
						if m[repo.Name] == nil {
							m[repo.Name] = make(map[string]version.Version)
						}
						m[repo.Name][pkgIndex.Package] = pkgIndex.Version
					}
				}
			}
		}
	}

	// prepare output string
	out := make([]string, 0)
	for name, pkgs := range m {
		out = append(out, fmt.Sprintf("%s:\n", name))
		for pkg, version := range pkgs {
			out = append(out, fmt.Sprintf("\t%s: %s\n", pkg, version.String()))
		}
	}

	// only do if output non-empty
	if len(out) > 0 {
		if !cmdDryRun {
			// SMTP config
			var mail sendmail.Config
			mail.Headers.From = c.Email.Header.From
			mail.Headers.ReplyTo = c.Email.Header.ReplyTo
			mail.Headers.Subject = c.Email.Header.Subject
			mail.Headers.IsText = true
			mail.Body.Message = strings.Join(out, "")
			mail.SMTP.Server = c.Email.SMTP.Server
			mail.SMTP.Port = c.Email.SMTP.Port
			mail.SMTP.Email = c.Email.SMTP.Address
			mail.SMTP.Password = c.Email.SMTP.Password
			// send email to multiple addresses
			for _, to := range c.Email.Header.To {
				// set TO header
				mail.Headers.To = to
				// send actual mail
				err = mail.Send()
				if err != nil {
					log.Println(err)
				}
			}
		} else {
			fmt.Printf("%s", strings.Join(out, ""))
		}
	}

	// save config
	if cmdUpdateConfig {
		err = saveConfig(c, cmdConfigPath)
		if err != nil {
			log.Fatal(err)
		}
	}
}
