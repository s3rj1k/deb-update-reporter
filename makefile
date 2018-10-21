GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
NAME=deb-update-reporter

all: clean build pack
build:
	$(GOBUILD) -ldflags="-s -w" -o $(NAME) -v
clean:
	$(GOCLEAN)
	rm -f $(NAME)
deps:
	$(GOGET) -u "github.com/pkg/errors"
	$(GOGET) -u "github.com/s3rj1k/go-smtp-html-helper"
	$(GOGET) -u "gopkg.in/yaml.v2"
	$(GOGET) -u "pault.ag/go/debian/control"
	$(GOGET) -u "pault.ag/go/debian/version"
