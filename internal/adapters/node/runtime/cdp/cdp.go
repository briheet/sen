// Package cdp implements a minimal Chrome DevTools Protocol client over a websocket.
package cdp

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
)

const magicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// Client is a minimal CDP websocket client.
type Client struct {
	conn    net.Conn
	reader  *bufio.Reader
	nextID  uint64
	pending map[uint64]chan []byte
	mu      sync.Mutex
}

// Dial connects to a CDP websocket endpoint.
func Dial(ctx context.Context, rawURL string) (*Client, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	address := parsed.Host
	if !strings.Contains(address, ":") {
		address += ":80"
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	key := make([]byte, 16)
	_, _ = rand.Read(key)
	key64 := base64.StdEncoding.EncodeToString(key)
	path := parsed.Path
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", path, address, key64)
	if _, err := conn.Write([]byte(request)); err != nil {
		conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !strings.Contains(status, "101") {
		conn.Close()
		return nil, fmt.Errorf("websocket upgrade failed: %s", strings.TrimSpace(status))
	}
	accept := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		header, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(header), "Sec-WebSocket-Accept") {
			accept = strings.TrimSpace(value)
		}
	}
	sum := sha1.Sum([]byte(key64 + magicGUID))
	if accept != base64.StdEncoding.EncodeToString(sum[:]) {
		conn.Close()
		return nil, errors.New("websocket accept mismatch")
	}

	client := &Client{conn: conn, reader: reader, pending: make(map[uint64]chan []byte)}
	go client.readLoop()
	return client, nil
}

// Call sends a CDP command and decodes its result into out.
func (c *Client) Call(ctx context.Context, method string, params any, out any) error {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	ch := make(chan []byte, 1)
	c.pending[id] = ch
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	message := map[string]any{"id": id, "method": method}
	if params != nil {
		message["params"] = params
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if err := c.writeText(payload); err != nil {
		return err
	}

	select {
	case response, ok := <-ch:
		if !ok {
			return errors.New("cdp: connection closed")
		}
		var envelope struct {
			ID    uint64 `json:"id"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal(response, &envelope); err != nil {
			return err
		}
		if envelope.Error != nil {
			return fmt.Errorf("cdp %s: %s", method, envelope.Error.Message)
		}
		if out != nil {
			return json.Unmarshal(envelope.Result, out)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close closes the websocket connection.
func (c *Client) Close() error {
	_ = c.writeFrame(0x8, nil)
	return c.conn.Close()
}

func (c *Client) readLoop() {
	defer c.conn.Close()
	for {
		payload, err := c.readFrame()
		if err != nil {
			c.mu.Lock()
			for _, ch := range c.pending {
				close(ch)
			}
			c.pending = make(map[uint64]chan []byte)
			c.mu.Unlock()
			return
		}
		var header struct {
			ID uint64 `json:"id"`
		}
		if json.Unmarshal(payload, &header) != nil {
			continue
		}
		c.mu.Lock()
		ch := c.pending[header.ID]
		c.mu.Unlock()
		if ch != nil {
			select {
			case ch <- payload:
			default:
			}
		}
	}
}

func (c *Client) readFrame() ([]byte, error) {
	for {
		header := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, header); err != nil {
			return nil, err
		}
		opcode := header[0] & 0x0f
		masked := header[1]&0x80 != 0
		length := uint64(header[1] & 0x7f)
		switch length {
		case 126:
			var extended [2]byte
			if _, err := io.ReadFull(c.reader, extended[:]); err != nil {
				return nil, err
			}
			length = uint64(binary.BigEndian.Uint16(extended[:]))
		case 127:
			var extended [8]byte
			if _, err := io.ReadFull(c.reader, extended[:]); err != nil {
				return nil, err
			}
			length = binary.BigEndian.Uint64(extended[:])
		}
		var maskKey [4]byte
		if masked {
			if _, err := io.ReadFull(c.reader, maskKey[:]); err != nil {
				return nil, err
			}
		}
		payload := make([]byte, length)
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return nil, err
		}
		if masked {
			for index := range payload {
				payload[index] ^= maskKey[index%4]
			}
		}
		switch opcode {
		case 0x1, 0x2:
			return payload, nil
		case 0x8:
			return nil, io.EOF
		case 0x9:
			_ = c.writeFrame(0xa, payload)
		}
	}
}

func (c *Client) writeText(payload []byte) error {
	return c.writeFrame(0x1, payload)
}

func (c *Client) writeFrame(opcode byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	maskKey := make([]byte, 4)
	_, _ = rand.Read(maskKey)
	header := []byte{0x80 | opcode}
	switch {
	case len(payload) < 126:
		header = append(header, 0x80|byte(len(payload)))
	case len(payload) <= 0xffff:
		header = append(header, 0x80|126)
		header = binary.BigEndian.AppendUint16(header, uint16(len(payload)))
	default:
		header = append(header, 0x80|127)
		header = binary.BigEndian.AppendUint64(header, uint64(len(payload)))
	}
	header = append(header, maskKey...)
	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	masked := make([]byte, len(payload))
	for index := range payload {
		masked[index] = payload[index] ^ maskKey[index%4]
	}
	_, err := c.conn.Write(masked)
	return err
}
