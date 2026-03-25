// Package main implements the bonkmcp CLI tool.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	bonkmcp "github.com/joakimcarlsson/bonk/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	transport := flag.String(
		"transport", "stdio",
		"Transport mode: stdio or sse",
	)
	port := flag.Int(
		"port", 8080,
		"Port for SSE transport",
	)
	headless := flag.Bool(
		"headless", true,
		"Run Chrome in headless mode",
	)
	stealth := flag.Bool(
		"stealth", true,
		"Enable stealth mode",
	)
	chrome := flag.String(
		"chrome", "",
		"Path to Chrome binary",
	)
	flag.Parse()

	sess := bonkmcp.NewSession(
		bonkmcp.WithHeadless(*headless),
		bonkmcp.WithStealth(*stealth),
		bonkmcp.WithChromePath(*chrome),
	)

	s := bonkmcp.NewServer(sess)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		sess.Close()
		os.Exit(0)
	}()

	switch *transport {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			log.Fatal(err)
		}
	case "sse":
		addr := fmt.Sprintf(":%d", *port)
		sseServer := server.NewSSEServer(s,
			server.WithBaseURL(
				fmt.Sprintf("http://localhost:%d", *port),
			),
		)
		log.Printf("SSE server listening on %s", addr)
		if err := sseServer.Start(addr); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown transport: %s", *transport)
	}
}
