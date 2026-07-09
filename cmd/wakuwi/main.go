package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	wakuwi "github.com/stut/wakuwi"
	"github.com/stut/wakuwi/internal/process"
	"github.com/stut/wakuwi/internal/server"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var version = "dev"

func main() {
	port := flag.Int("port", 9753, "port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()

	log.Printf("starting wakuwi on %s", addr)

	pm, err := process.NewManager(ctx)
	if err != nil {
		log.Fatalf("init process manager: %v", err)
	}
	log.Printf("process manager initialised (log dir: %s/wakuwi)", os.TempDir())

	srv := server.New(wakuwi.StaticFiles, pm, version)

	// Bind the socket before starting the probe so the probe only
	// succeeds when OUR server is listening, not any pre-existing process.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	log.Printf("listening on %s", addr)

	log.Printf("open %s", url)

	if err := http.Serve(ln, srv); err != nil {
		log.Fatal(err)
	}
}
