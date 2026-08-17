// Package runtime manages an OCaml target and collects its runtime data.
package runtime

import (
	"encoding/binary"
	"errors"
)

// metadataHeader is the on-disk OCaml runtime_events header
// (caml/runtime_events.h: struct runtime_events_metadata_header).
type metadataHeader struct {
	Version            uint64
	MaxDomains         uint64
	RingHeaderSizeByte uint64
	RingSizeBytes      uint64
	RingSizeElements   uint64
	HeadersOffset      uint64
	DataOffset         uint64
	CustomEventsOffset uint64
}

// Packed item header bits (caml/runtime_events.h RUNTIME_EVENTS_* macros).
const (
	itemLengthShift = 54
	itemLengthMask  = (1 << 10) - 1
	itemIsUserBit   = 1 << 53
	itemTypeShift   = 49
	itemTypeMask    = (1 << 4) - 1
	itemIDShift     = 36
	itemIDMask      = (1 << 13) - 1

	// ev_runtime_message_type values.
	msgExit = 3 // EV_EXIT: span end

	// GC span ids (ev_* enum in runtime_events.h).
	gcMinor = 17 // EV_MINOR
	gcMajor = 27 // EV_MAJOR

	// User span payload layout (verified against the ring):
	// payload[0] = timestamp, payload[1] = function id.
	userSpanIDWord = 2 // payload word index of the function id
)

// eventCounts accumulated from the OCaml runtime events ring.
type eventCounts struct {
	MinorCollections uint64
	MajorCollections uint64
	// FunctionSpans counts, keyed by instrumented function id.
	FunctionSpans map[uint64]uint64
	// Spans holds each observed custom span (timestamp + function id).
	Spans []Span
}

// Span is one custom user-span observation.
type Span struct {
	Timestamp uint64
	Function  uint64
}

// decodeEvents parses an OCaml .events ring buffer into GC collection counts
// and per-function custom span counts.
func decodeEvents(data []byte) (eventCounts, error) {
	counts := eventCounts{FunctionSpans: make(map[uint64]uint64)}
	if len(data) < 64 {
		return counts, errors.New("runtime events: file too small")
	}
	var meta metadataHeader
	if err := decodeHeader(data[:64], &meta); err != nil {
		return counts, err
	}
	if meta.Version != 1 || meta.DataOffset == 0 {
		return counts, errors.New("runtime events: unsupported header")
	}

	ring := data[meta.DataOffset:]
	for offset := 0; offset+8 <= len(ring); {
		word := binary.LittleEndian.Uint64(ring[offset:])
		// The ring is sparse: most of it is zero padding that must be skipped
		// word-by-word, so fast-forward over the longest zero run.
		if word == 0 {
			next := nextNonZero(ring, offset+8)
			if next < 0 {
				break
			}
			offset = next
			continue
		}
		length := (word >> itemLengthShift) & itemLengthMask
		if length > 63 {
			offset += 8
			continue
		}
		msgType := (word >> itemTypeShift) & itemTypeMask
		id := (word >> itemIDShift) & itemIDMask
		if word&itemIsUserBit != 0 {
			// custom span: payload[0]=timestamp, payload[1]=function id
			if offset+userSpanIDWord*8+8 <= len(ring) {
				ts := binary.LittleEndian.Uint64(ring[offset+8:])
				fn := binary.LittleEndian.Uint64(ring[offset+userSpanIDWord*8:])
				counts.FunctionSpans[fn]++
				counts.Spans = append(counts.Spans, Span{Timestamp: ts, Function: fn})
			}
			offset += (int(length) + 1) * 8
			continue
		}
		if msgType == msgExit {
			switch id {
			case gcMinor:
				counts.MinorCollections++
			case gcMajor:
				counts.MajorCollections++
			}
		}
		offset += (int(length) + 1) * 8
	}
	return counts, nil
}

// nextNonZero returns the byte offset of the next non-zero 8-byte word at or
// after start, or -1 if none exist in the remaining buffer.
func nextNonZero(ring []byte, start int) int {
	for i := start; i+8 <= len(ring); i += 8 {
		if ring[i] != 0 || ring[i+1] != 0 || ring[i+2] != 0 || ring[i+3] != 0 ||
			ring[i+4] != 0 || ring[i+5] != 0 || ring[i+6] != 0 || ring[i+7] != 0 {
			return i
		}
	}
	return -1
}

// decodeHeader reads the fixed-width metadata header as eight little-endian words.
func decodeHeader(b []byte, m *metadataHeader) error {
	if len(b) < 8*8 {
		return errors.New("runtime events: header too small")
	}
	m.Version = binary.LittleEndian.Uint64(b[0:])
	m.MaxDomains = binary.LittleEndian.Uint64(b[8:])
	m.RingHeaderSizeByte = binary.LittleEndian.Uint64(b[16:])
	m.RingSizeBytes = binary.LittleEndian.Uint64(b[24:])
	m.RingSizeElements = binary.LittleEndian.Uint64(b[32:])
	m.HeadersOffset = binary.LittleEndian.Uint64(b[40:])
	m.DataOffset = binary.LittleEndian.Uint64(b[48:])
	m.CustomEventsOffset = binary.LittleEndian.Uint64(b[56:])
	return nil
}
