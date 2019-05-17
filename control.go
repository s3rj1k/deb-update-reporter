package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context/ctxhttp"
	"pault.ag/go/debian/control"
)

func getPackagesBinaryIndexURL(urls []string) ([]control.BinaryIndex, error) {
	out := []control.BinaryIndex{}

	// custom transport config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // nolint: gosec
		},
	}

	// custom http client config
	var client = &http.Client{
		Transport: tr,
	}

	for _, url := range urls {
		// set timeout
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		// get data from remote URL
		response, err := ctxhttp.Get(ctx, client, url)
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
