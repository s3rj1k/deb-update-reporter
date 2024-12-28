package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/version"
)

func main() {
	var (
		cmdDryRun       bool
		cmdUpdateConfig bool
		cmdConfigPath   string
	)

	log.SetFlags(0)

	flag.StringVar(&cmdConfigPath, "config-path", "config.yaml", "path to config file")
	flag.BoolVar(&cmdUpdateConfig, "update-config", false, "save config before exit")
	flag.BoolVar(&cmdDryRun, "dry-run", false, "print to console instead of sending email")
	flag.Parse()

	m := make(map[string]map[string]version.Version)

	c, err := getConfig(cmdConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range c.Repo {
		var bIndex []control.BinaryIndex

		bIndex, err = getPackagesBinaryIndexURL(repo.URL)
		if err != nil {
			log.Fatal(err)
		}

		for _, pkgIndex := range bIndex {
			for _, pkg := range repo.Packages {
				if strings.EqualFold(pkg.Name, pkgIndex.Package) {
					var ver version.Version

					ver, err = version.Parse(pkg.VersionNewerThan)
					if err != nil {
						log.Fatal(err)
					}

					if version.Compare(pkgIndex.Version, ver) > 0 {
						pkg.VersionNewerThan = pkgIndex.Version.String()

						if m[repo.Name] == nil {
							m[repo.Name] = make(map[string]version.Version)
						}

						m[repo.Name][pkgIndex.Package] = pkgIndex.Version
					}
				}
			}
		}
	}

	out := []string{}
	for name, pkgs := range m {
		out = append(out, fmt.Sprintf("%s:\n", name))
		for pkg, ver := range pkgs {
			out = append(out, fmt.Sprintf("\t%s: %s\n", pkg, ver.String()))
		}
	}

	if len(out) == 0 {
		return
	}

	if !cmdDryRun {
		for _, addr := range c.Email.To {
			mail := MailConfig{
				Headers: c.Email.Headers,
				SMTP:    c.Email.SMTP,
				Body:    strings.Join(out, ""),
			}

			mail.Headers.To = addr

			if err = mail.SendText(); err != nil {
				log.Println(err)
			}
		}
	} else {
		fmt.Printf("%s", strings.Join(out, ""))
	}

	if cmdUpdateConfig {
		if err = saveConfig(c, cmdConfigPath); err != nil {
			log.Fatal(err)
		}
	}
}
