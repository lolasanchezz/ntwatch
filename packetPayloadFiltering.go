package main

import (
	"unicode"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/gopacket"
)

func lookThroughBody(body []byte) string {
	var message string
	var buf []byte
	//slice up the body byte until its empty
	for len(body) > 0 {
		r, size := utf8.DecodeRune(body)
		if r == utf8.RuneError && size == 1 {
			//rune is invalid - take what we have and put it in the message
			flushBuf(&message, &buf)
			//slice up body
			body = body[1:]
			continue
		}

		// get valid rune
		buf = append(buf, body[:size]...)
		body = body[size:]
	}

	// flush last buffer
	flushBuf(&message, &buf)
	return message
}

// checks if a buffer has an appropiate length and amount of characters
func flushBuf(message *string, buf *[]byte) {
	if len(*buf) == 0 {
		return
	}

	runeCount := utf8.RuneCount(*buf)
	if runeCount > 12 && countLetters(*buf) >= 8 {
		*message += string(*buf) + "\n"
	}
	*buf = nil
}

// counts how many runes are letters (
func countLetters(b []byte) int {
	count := 0
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if unicode.IsLetter(r) {
			count++
		}
		b = b[size:]
	}
	return count
}

/// bubble tea viewing

type payloadViewer struct {
	currentData string
	packets     chan gopacket.Packet
}

func (m *payloadViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	select {
	case pkt := <-m.packets:
		if pkt.ApplicationLayer() != nil {
			m.currentData = lookThroughBody(pkt.ApplicationLayer().Payload())
		}
	default:
		m.currentData = "no packets yet"
	}
	return m, nil
}

func (m *payloadViewer) Init() tea.Cmd {
	return nil
}

func (m *payloadViewer) View() string {
	return m.currentData
}
