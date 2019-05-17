package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/version"

	sendmail "github.com/s3rj1k/go-smtp-html-helper"
)

func main() {
	var (
		cmdDryRun       bool
		cmdUpdateConfig bool
		cmdConfigPath   string
	)

	// set simple log output
	log.SetFlags(0)

	// cmd flags
	flag.StringVar(&cmdConfigPath, "config-path", "config.yaml", "path to config file")
	flag.BoolVar(&cmdUpdateConfig, "update-config", false, "save config before exit")
	flag.BoolVar(&cmdDryRun, "dry-run", false, "print to console instead of sending email")
	flag.Parse()

	// output map
	m := make(map[string]map[string]version.Version)

	// read config
	c, err := getConfig(cmdConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	// loop-over repos inside config
	for _, repo := range c.Repo {
		var bIndex []control.BinaryIndex

		// parse repo db
		bIndex, err = getPackagesBinaryIndexURL(repo.URL)
		if err != nil {
			log.Fatal(err)
		}

		// loop-over packages in repo db
		for _, pkgIndex := range bIndex {
			// loop-over config packages
			for _, pkg := range repo.Packages {
				// check if package in config
				if strings.EqualFold(pkg.Name, pkgIndex.Package) {
					var ver version.Version
					// parse version string
					ver, err = version.Parse(pkg.VersionNewerThan)
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
	out := []string{}
	for name, pkgs := range m {
		out = append(out, fmt.Sprintf("%s:\n", name))
		for pkg, ver := range pkgs {
			out = append(out, fmt.Sprintf("\t%s: %s\n", pkg, ver.String()))
		}
	}

	// nothing to do
	if len(out) == 0 {
		return
	}

	// print or send email
	if !cmdDryRun {
		for _, addr := range c.Email.To {
			// SMTP config
			mail := sendmail.Config{
				Headers: c.Email.Headers,
				SMTP:    c.Email.SMTP,
				Body:    strings.Join(out, ""),
			}

			// set TO header
			mail.Headers.To = addr

			// send email
			if err = mail.SendText(); err != nil {
				log.Println(err)
			}
		}
	} else {
		fmt.Printf("%s", strings.Join(out, ""))
	}

	// save config
	if cmdUpdateConfig {
		if err = saveConfig(c, cmdConfigPath); err != nil {
			log.Fatal(err)
		}
	}
}
