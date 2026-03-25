package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joakimcarlsson/bonk/internal/gen"
	"github.com/joakimcarlsson/bonk/internal/pdl"
)

func main() {
	var (
		protoDir = flag.String(
			"proto",
			"",
			"directory containing .pdl files (downloads from Chrome if empty)",
		)
		outDir = flag.String(
			"out",
			"proto",
			"output directory for generated packages",
		)
		modulePath = flag.String(
			"module",
			"github.com/joakimcarlsson/bonk",
			"Go module path",
		)
	)
	flag.Parse()

	tmpDir := *protoDir
	if tmpDir == "" {
		var err error
		tmpDir, err = os.MkdirTemp("", "bonkgen-*")
		if err != nil {
			log.Fatalf("create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)
	}

	var proto *pdl.Protocol

	if *protoDir == "" {
		fmt.Println("Downloading protocol definitions...")
		p, err := downloadAndParse(tmpDir)
		if err != nil {
			log.Printf("download and parse: %v\n", err)
			return
		}
		proto = p
	} else {
		fmt.Printf("Parsing protocol definitions from %s...\n", *protoDir)
		files, err := findPDLFiles(*protoDir)
		if err != nil {
			log.Fatalf("find pdl files: %v", err)
		}
		p, err := parseLocalFiles(files...)
		if err != nil {
			log.Fatalf("parse: %v", err)
		}
		proto = p
	}

	fmt.Printf("Found %d domains\n", len(proto.Domains))

	g := &gen.Generator{
		Proto:      proto,
		OutputDir:  *outDir,
		ModulePath: *modulePath,
	}

	if err := g.Generate(); err != nil {
		log.Printf("generate: %v\n", err)
		return
	}

	fmt.Printf("Generated code in %s/\n", *outDir)
}

func findPDLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 4 &&
			e.Name()[len(e.Name())-4:] == ".pdl" {
			files = append(files, dir+"/"+e.Name())
		}
	}
	return files, nil
}
