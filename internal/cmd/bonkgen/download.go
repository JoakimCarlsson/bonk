// Package main implements the bonkgen code generator.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joakimcarlsson/bonk/internal/pdl"
)

const (
	pdlBaseURL         = "https://raw.githubusercontent.com/ChromeDevTools/devtools-protocol/master/pdl/"
	browserProtocolURL = pdlBaseURL + "browser_protocol.pdl"
	jsProtocolURL      = pdlBaseURL + "js_protocol.pdl"
)

func downloadAndParse(protoDir string) (*pdl.Protocol, error) {
	browserPDL := filepath.Join(protoDir, "browser_protocol.pdl")
	jsPDL := filepath.Join(protoDir, "js_protocol.pdl")

	if err := downloadFile(browserProtocolURL, browserPDL); err != nil {
		return nil, fmt.Errorf("download browser protocol: %w", err)
	}

	if err := downloadFile(jsProtocolURL, jsPDL); err != nil {
		return nil, fmt.Errorf("download js protocol: %w", err)
	}

	if err := resolveIncludes(protoDir, browserPDL); err != nil {
		return nil, fmt.Errorf("resolve browser includes: %w", err)
	}

	if err := resolveIncludes(protoDir, jsPDL); err != nil {
		return nil, fmt.Errorf("resolve js includes: %w", err)
	}

	return parseWithIncludes(protoDir, browserPDL, jsPDL)
}

func resolveIncludes(protoDir, mainFile string) error {
	data, err := os.ReadFile(mainFile)
	if err != nil {
		return err
	}

	proto, err := pdl.Parse(strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	for _, inc := range proto.Includes {
		destPath := filepath.Join(protoDir, inc)
		if _, err := os.Stat(destPath); err == nil {
			continue
		}

		url := pdlBaseURL + inc
		if err := downloadFile(url, destPath); err != nil {
			return fmt.Errorf("download include %s: %w", inc, err)
		}
	}

	return nil
}

func parseWithIncludes(
	protoDir string,
	mainFiles ...string,
) (*pdl.Protocol, error) {
	merged := &pdl.Protocol{}

	for _, mainFile := range mainFiles {
		data, err := os.ReadFile(mainFile)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", mainFile, err)
		}

		main, err := pdl.Parse(strings.NewReader(string(data)))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", mainFile, err)
		}

		if main.Version.Major != "" && merged.Version.Major == "" {
			merged.Version = main.Version
		}

		merged.Domains = append(merged.Domains, main.Domains...)

		for _, inc := range main.Includes {
			incPath := filepath.Join(protoDir, inc)
			incData, err := os.ReadFile(incPath)
			if err != nil {
				return nil, fmt.Errorf("read include %s: %w", inc, err)
			}

			incProto, err := pdl.Parse(strings.NewReader(string(incData)))
			if err != nil {
				return nil, fmt.Errorf("parse include %s: %w", inc, err)
			}

			merged.Domains = append(merged.Domains, incProto.Domains...)
		}
	}

	return merged, nil
}

func parseLocalFiles(files ...string) (*pdl.Protocol, error) {
	merged := &pdl.Protocol{}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}

		proto, err := pdl.Parse(strings.NewReader(string(data)))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}

		if proto.Version.Major != "" && merged.Version.Major == "" {
			merged.Version = proto.Version
		}

		merged.Domains = append(merged.Domains, proto.Domains...)
	}

	return merged, nil
}

func downloadFile(url, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
