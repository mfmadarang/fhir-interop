package hl7v2

import (
	"fmt"
	"strings"
)

type Segment struct {
	ID     string
	Fields []Field
}

func (s Segment) Field(i int) Field {
	idx := i - 1
	if idx < 0 || idx >= len(s.Fields) {
		return ""
	}
	return s.Fields[idx]
}

type Message struct {
	Segments []Segment
}

func (m *Message) Segment(id string) *Segment {
	for i := range m.Segments {
		if m.Segments[i].ID == id {
			return &m.Segments[i]
		}
	}
	return nil
}

func (m *Message) SegmentsOf(id string) []Segment {
	var out []Segment
	for _, seg := range m.Segments {
		if seg.ID == id {
			out = append(out, seg)
		}
	}
	return out
}

func Parse(data []byte) (*Message, error) {
	text := string(data)
	text = strings.ReplaceAll(text, "\r\n", "\r")
	text = strings.ReplaceAll(text, "\n", "\r")

	lines := strings.Split(text, "\r")

	var msg Message
	var fieldSep byte

	for lineNum, line := range lines {
		line = strings.TrimRight(line, " \t")
		if line == "" {
			continue
		}
		if len(line) < 4 {
			return nil, fmt.Errorf("hl7v2: line %d: segment too short: %q", lineNum+1, line)
		}

		segID := line[0:3]

		if segID == "MSH" {
			fieldSep = line[3]
			rest := line[4:]
			parts := strings.Split(rest, string(fieldSep))

			fields := make([]Field, 0, len(parts)+1)
			fields = append(fields, Field(string(fieldSep))) // MSH-1
			for _, p := range parts {
				fields = append(fields, Field(p))
			}
			msg.Segments = append(msg.Segments, Segment{ID: segID, Fields: fields})
			continue
		}

		if fieldSep == 0 {
			return nil, fmt.Errorf("hl7v2: line %d: %q appears before MSH; cannot determine field separator", lineNum+1, segID)
		}
		if line[3] != fieldSep {
			return nil, fmt.Errorf("hl7v2: line %d: segment %q does not use the message field separator %q", lineNum+1, segID, string(fieldSep))
		}

		rest := line[4:]
		parts := strings.Split(rest, string(fieldSep))
		fields := make([]Field, len(parts))
		for i, p := range parts {
			fields[i] = Field(p)
		}
		msg.Segments = append(msg.Segments, Segment{ID: segID, Fields: fields})
	}

	if len(msg.Segments) == 0 {
		return nil, fmt.Errorf("hl7v2: no segments found")
	}
	if msg.Segments[0].ID != "MSH" {
		return nil, fmt.Errorf("hl7v2: message must start with MSH, got %q", msg.Segments[0].ID)
	}

	return &msg, nil
}
