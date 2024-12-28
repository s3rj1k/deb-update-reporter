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

	"golang.org/x/net/context/ctxhttp"
	"pault.ag/go/debian/control"
)

func getPackagesBinaryIndexURL(urls []string) ([]control.BinaryIndex, error) {
	out := []control.BinaryIndex{}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	var client = &http.Client{
		Transport: tr,
	}

	for _, url := range urls {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		response, err := ctxhttp.Get(ctx, client, url)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode == 404 {
			return nil, fmt.Errorf("remote URL not found: %s", url)
		}

		var reader io.ReadCloser
		var index []control.BinaryIndex

		switch response.Header.Get("Content-Type") {
		case "gzip", "application/x-gzip", "application/gzip":
			reader, err = gzip.NewReader(response.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to parse URL=%q: %w", url, err)
			}
			defer reader.Close()

			index, err = control.ParseBinaryIndex(bufio.NewReader(reader))
			if err != nil {
				return nil, fmt.Errorf("failed to parse URL=%q: %w", url, err)
			}

			out = append(out, index...)
		case "text/plain":
			index, err = control.ParseBinaryIndex(bufio.NewReader(response.Body))
			if err != nil {
				return nil, fmt.Errorf("failed to parse URL=%q: %w", url, err)
			}

			out = append(out, index...)
		case "application/octet-stream":
			switch {
			case strings.HasSuffix(url, ".gz"):
				reader, err = gzip.NewReader(response.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to parse URL=%q: %w", url, err)
				}
				defer reader.Close()

				index, err := control.ParseBinaryIndex(bufio.NewReader(reader))
				if err != nil {
					return nil, fmt.Errorf("failed to parse URL=%q: %w", url, err)
				}

				out = append(out, index...)
			default:
				index, err := control.ParseBinaryIndex(bufio.NewReader(response.Body))
				if err != nil {
					return nil, fmt.Errorf("failed to parse URL=%q: %w", url, err)
				}

				out = append(out, index...)
			}
		default:
			index, err := control.ParseBinaryIndex(bufio.NewReader(response.Body))
			if err != nil {
				return nil, fmt.Errorf("failed to parse URL=%q: %w", url, err)
			}

			out = append(out, index...)
		}
	}

	return out, nil
}
