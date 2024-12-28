package main

import (
	"fmt"
	"os"

	"pault.ag/go/debian/version"
)

func mustParseVersion(input string) version.Version {
	ver, err := version.Parse(input)
	if err != nil {
		panic(err)
	}

	return ver
}

// Sane replacement for: `dpkg --compare-versions 0.1.0 gt 0.2.0 && echo $?`
func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: <dpkg-compare-versions> <version1> <version2>")

		os.Exit(1)
	}

	versionLeft := os.Args[1]
	versionRight := os.Args[2]

	cmp := version.Compare(mustParseVersion(versionLeft), mustParseVersion(versionRight))

	if cmp == 0 {
		fmt.Printf("%s == %s\n", versionLeft, versionRight)
	} else if cmp < 0 {
		fmt.Printf("%s < %s\n", versionLeft, versionRight)
	} else if cmp > 0 {
		fmt.Printf("%s > %s\n", versionLeft, versionRight)
	}
}
