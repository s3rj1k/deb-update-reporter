package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"pault.ag/go/debian/control"
)

func getPackagesBinaryIndexURL(urls []string) ([]control.BinaryIndex, error) {
	out := []control.BinaryIndex{}

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
			return nil, err
		}
		defer response.Body.Close()

		// error 404
		if response.StatusCode == 404 {
			return nil, fmt.Errorf("remote URL not found: %s", url)
		}

		// set default reader
		var reader io.ReadCloser
		// declare control index
		var index []control.BinaryIndex

		// Check that the server actually sent compressed data
		switch response.Header.Get("Content-Type") {
		case "gzip", "application/x-gzip", "application/gzip":
			// decode gzip
			reader, err = gzip.NewReader(response.Body)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse URL=%s", url)
			}
			defer reader.Close()

			// parse binary index
			index, err = control.ParseBinaryIndex(bufio.NewReader(reader))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse URL=%s", url)
			}

			// append to output
			out = append(out, index...)

		case "text/plain":
			// parse binary index
			index, err = control.ParseBinaryIndex(bufio.NewReader(response.Body))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse URL=%s", url)
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
					return nil, errors.Wrapf(err, "failed to parse URL=%s", url)
				}
				defer reader.Close()

				// parse binary index
				index, err := control.ParseBinaryIndex(bufio.NewReader(reader))
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse URL=%s", url)
				}

				// append to output
				out = append(out, index...)

			default:
				// parse binary index
				index, err := control.ParseBinaryIndex(bufio.NewReader(response.Body))
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse URL=%s", url)
				}

				// append to output
				out = append(out, index...)
			}

		default:
			// parse binary index
			index, err := control.ParseBinaryIndex(bufio.NewReader(response.Body))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse URL=%s", url)
			}

			// append to output
			out = append(out, index...)
		}
	}

	return out, nil
}
