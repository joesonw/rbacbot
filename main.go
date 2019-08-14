package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joesonw/rbacbot/github"
)

func main() {
	provider := os.Args[1]

	if provider != "github" {
		println("Currently, only github is supported.")
		os.Exit(1)
	}

	addr := os.Getenv("RBACBOT_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	println("listening on: ", addr)

	ttl := os.Getenv("RBACBOT_TTL")
	if ttl == "" {
		ttl = "1h"
		println("env RBACBOT_TTL is not set, using default cache ttl: 1 hour.")
	}

	duration, err := time.ParseDuration(ttl)
	if err != nil {
		panic(err)
	}

	server, err := github.New(os.Getenv("RBACBOT_CONFIG_NAME"), os.Getenv("RBACBOT_WEBHOOK_SECRET"), os.Getenv("RBACBOT_ACCESS_TOKEN"), duration)
	if err != nil {
		panic(err)
	}

	if err := http.Serve(listener, server); err != nil {
		panic(err)
	}
}
