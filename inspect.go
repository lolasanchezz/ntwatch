package main

// looking inside packets for things the analytics want: hostnames and plaintext

import (
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// a name learned from the traffic itself - a dns answer, or the server a tls
// connection asked for
type hostname struct{ ip, name string }

func dnsHostnames(p gopacket.Packet) []hostname {
	dns, _ := p.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if dns == nil {
		// mdns isn't decoded by default - it's how devices on the network name themselves
		if udp, ok := p.Layer(layers.LayerTypeUDP).(*layers.UDP); ok && (udp.SrcPort == 5353 || udp.DstPort == 5353) {
			var d layers.DNS
			if d.DecodeFromBytes(udp.Payload, gopacket.NilDecodeFeedback) == nil {
				dns = &d
			}
		}
	}
	if dns == nil || !dns.QR {
		return nil
	}

	asked := ""
	if len(dns.Questions) > 0 {
		asked = string(dns.Questions[0].Name)
	}
	var names []hostname
	for _, a := range dns.Answers {
		if (a.Type != layers.DNSTypeA && a.Type != layers.DNSTypeAAAA) || a.IP == nil {
			continue
		}
		// the name the app asked for reads better than the end of a cname chain
		name := asked
		if name == "" {
			name = string(a.Name)
		}
		names = append(names, hostname{ip: a.IP.String(), name: name})
	}
	return names
}

// the server name from a tls ClientHello, if the payload starts with one. a
// hello can span a couple of tcp segments and this only sees the first, so the
// name is sometimes out of reach
func clientHelloSNI(b []byte) string {
	// record header: type 22 (handshake), version, length
	if len(b) < 9 || b[0] != 0x16 || b[1] != 0x03 || b[5] != 0x01 { // 0x01 = ClientHello
		return ""
	}
	b = b[9:] // record header + handshake header

	skip := func(n int) bool {
		if n > len(b) {
			return false
		}
		b = b[n:]
		return true
	}
	u8 := func() int {
		if len(b) < 1 {
			return 1 << 30
		}
		return int(b[0])
	}
	u16 := func() int {
		if len(b) < 2 {
			return 1 << 30
		}
		return int(b[0])<<8 | int(b[1])
	}

	// version + random, session id, cipher suites, compression methods
	if !skip(34) || !skip(1+u8()) || !skip(2+u16()) || !skip(1+u8()) || !skip(2) {
		return ""
	}
	// extensions - possibly cut off by the end of the segment
	for len(b) >= 4 {
		typ := int(b[0])<<8 | int(b[1])
		n := int(b[2])<<8 | int(b[3])
		b = b[4:]
		if n > len(b) {
			return ""
		}
		if typ == 0 { // server_name: list length, name type (0 = host), name length, name
			ext := b[:n]
			if len(ext) < 5 || ext[2] != 0 {
				return ""
			}
			nameLen := int(ext[3])<<8 | int(ext[4])
			if 5+nameLen > len(ext) {
				return ""
			}
			return string(ext[5 : 5+nameLen])
		}
		b = b[n:]
	}
	return ""
}

// ports where payloads are encrypted, or where plaintext is normal and not interesting
var quietPorts = map[string]bool{
	"22": true, "53": true, "443": true, "465": true, "853": true, "993": true,
	"995": true, "1900": true, "5223": true, "5353": true, "8443": true,
}

const maxSnippet = 160

// readable text in a payload, for spotting unencrypted traffic. empty if none
func readableText(payload []byte, srcPort, dstPort string) string {
	if len(payload) == 0 || quietPorts[srcPort] || quietPorts[dstPort] {
		return ""
	}
	// tls records (handshake, alert, app data) are binary apart from the odd hostname
	if len(payload) >= 3 && payload[0] >= 0x14 && payload[0] <= 0x17 && payload[1] == 0x03 {
		return ""
	}
	text := strings.Join(strings.Fields(lookThroughBody(payload)), " ")
	if runes := []rune(text); len(runes) > maxSnippet {
		text = string(runes[:maxSnippet]) + "…"
	}
	return text
}
