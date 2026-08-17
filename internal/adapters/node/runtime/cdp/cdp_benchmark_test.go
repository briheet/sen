package cdp

import (
	"bufio"
	"net"
	"testing"
)

const benchPayloadSize = 1024

func BenchmarkWriteFrame(b *testing.B) {
	server, client := net.Pipe()
	defer func() { _ = server.Close() }()
	defer func() { _ = client.Close() }()
	writer := &Client{conn: client}
	reader := &Client{conn: server, reader: bufio.NewReader(server)}
	go func() {
		for {
			if _, err := reader.readFrame(); err != nil {
				return
			}
		}
	}()

	payload := make([]byte, benchPayloadSize)
	b.ReportAllocs()
	for b.Loop() {
		if err := writer.writeText(payload); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadFrame(b *testing.B) {
	server, client := net.Pipe()
	defer func() { _ = server.Close() }()
	defer func() { _ = client.Close() }()
	writer := &Client{conn: server}
	reader := &Client{conn: client, reader: bufio.NewReader(client)}
	go func() {
		payload := make([]byte, benchPayloadSize)
		for {
			if err := writer.writeText(payload); err != nil {
				return
			}
		}
	}()

	b.ReportAllocs()
	for b.Loop() {
		if _, err := reader.readFrame(); err != nil {
			b.Fatal(err)
		}
	}
}
