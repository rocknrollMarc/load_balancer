// Copyright 2015, rocknrollmarc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// balancer.go [created: Thu, 19 Feb 2015]

package main

import (
	"flag"
	"io"
	"log"
	"net"
	"strings"
)

type Backends struct {
	servers []string
	n       int
}

func (b *Backends) Choose() string {
	idx := b.n % len(b.servers)
	b.n++
	return b.servers[idx]
}

func (b *Backends) String() string {
	return strings.Join(b.servers, ", ")
}

var (
	bind     = flag.String("bind", "", "The address to bind on")
	balance  = flag.String("balance", "", "The backend servers to balance connections across, seperated by commas")
	backends *Backends
)

func init() {
	flag.Parse()

	if *bind == "" {
		log.Fatalln("Specify the address to listen on with -bind")
	}

	servers := strings.Split(*balance, ",")
	if len(servers) == 1 && servers[0] == "" {
		log.Fatalln("Please specify backend servers with -backends")
	}

	backends = &Backends{servers: servers}
}

func copy(wc io.WriteCloser, r io.Reader) {
	defer wc.Close()
	io.Copy(wc, r)
}

func handleConnection(us net.Conn, server string) {
	ds, err := net.Dial("tcp", server)
	if err != nil {
		us.Close()
		log.Printf("failed to dail %s: %s", server, err)
		return
	}

	go copy(ds, us)
	go copy(us, ds)
}

func main() {
	ln, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatalf("failed to bind; %s", err)
	}

	log.Printf("listening on %s, balancing %s", *bind, backends)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept: %s", err)
			continue
		}
		go handleConnection(conn, backends.Choose())
	}
}
