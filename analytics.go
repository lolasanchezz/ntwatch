package main

// the numbers behind the analytics panels. like the rest of the model, nothing
// here is changed in place - every update hands back new maps/slices

import (
	"cmp"
	"maps"
	"net"
	"slices"
	"strconv"
)

const (
	trafficWindow = 60   // seconds of history for the bandwidth chart + top lists
	maxHostnames  = 4096 // remembered ip -> name pairs
	maxAlerts     = 20   // remembered plaintext alerts
	topProcesses  = 5
	topDests      = 6
	topLatency    = 6
	maxOpenPorts  = 40
	shownAlerts   = 8
)

// one second of traffic
type trafficSecond struct {
	sec   int64
	up    int64 // bytes sent from this machine
	down  int64 // bytes received
	procs map[processKey]usage
	dests map[string]destUsage // by remote ip
}

type usage struct{ up, down int64 }

type destUsage struct {
	bytes, packets int64
	procs          []processKey // who talked to it - a few at most
}

type plaintextAlert struct {
	proc       processKey
	remoteIP   string
	remotePort string
	protocol   string
	outgoing   bool
	snippet    string
	count      int
	time       int64 // unix millis of the latest one
}

func cloneOrNew[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return make(map[K]V)
	}
	return maps.Clone(m)
}

// the other end of a packet
func remoteOf(p PacketInfo) (ip, port string) {
	if p.outgoing {
		return p.destIP, p.destPort
	}
	return p.sourceIP, p.sourcePort
}

func (m model) countTraffic(proc processKey, p PacketInfo, system bool) model {
	sec := p.Timestamp.Unix()
	if n := len(m.traffic); n > 0 && sec <= m.traffic[n-1].sec-trafficWindow {
		return m // older than anything we keep
	}

	traffic := slices.Clone(m.traffic)
	find := func() int { return slices.IndexFunc(traffic, func(t trafficSecond) bool { return t.sec == sec }) }
	i := find()
	if i < 0 {
		traffic = append(traffic, trafficSecond{sec: sec})
		slices.SortFunc(traffic, func(a, b trafficSecond) int { return cmp.Compare(a.sec, b.sec) })
		newest := traffic[len(traffic)-1].sec
		traffic = slices.DeleteFunc(traffic, func(t trafficSecond) bool { return t.sec <= newest-trafficWindow })
		i = find()
	}

	t := traffic[i]
	size := int64(p.packetSize)
	if p.outgoing {
		t.up += size
	} else {
		t.down += size
	}
	if proc != (processKey{}) {
		t.procs = cloneOrNew(t.procs)
		u := t.procs[proc]
		if p.outgoing {
			u.up += size
		} else {
			u.down += size
		}
		t.procs[proc] = u
	}
	if !system {
		remote, _ := remoteOf(p)
		t.dests = cloneOrNew(t.dests)
		d := t.dests[remote]
		d.bytes += size
		d.packets++
		if proc != (processKey{}) && len(d.procs) < 3 && !slices.Contains(d.procs, proc) {
			d.procs = append(slices.Clip(d.procs), proc)
		}
		t.dests[remote] = d
	}
	traffic[i] = t
	m.traffic = traffic
	return m
}

func (m model) learnHostnames(names []hostname) model {
	if len(names) == 0 {
		return m
	}
	hosts := cloneOrNew(m.hostnames)
	order := m.hostOrder // appending only writes past older models' ends
	for _, h := range names {
		if _, known := hosts[h.ip]; !known {
			order = append(order, h.ip)
		}
		hosts[h.ip] = h.name
	}
	for len(order) > maxHostnames {
		delete(hosts, order[0])
		order = order[1:]
	}
	m.hostnames, m.hostOrder = hosts, order
	return m
}

func (m model) notePlaintext(proc processKey, p PacketInfo) model {
	remoteIP, remotePort := remoteOf(p)
	alerts := slices.Clone(m.plaintext)
	alert := plaintextAlert{
		proc:       proc,
		remoteIP:   remoteIP,
		remotePort: remotePort,
		protocol:   p.packetProtocol.String(),
		outgoing:   p.outgoing,
		snippet:    p.readable,
		count:      1,
		time:       p.Timestamp.UnixMilli(),
	}
	// one alert per process + destination, bumped to newest when it happens again
	same := func(a plaintextAlert) bool {
		return a.proc == proc && a.remoteIP == remoteIP && a.remotePort == remotePort
	}
	if i := slices.IndexFunc(alerts, same); i >= 0 {
		alert.count += alerts[i].count
		alerts = slices.Delete(alerts, i, i+1)
	}
	alerts = append(alerts, alert)
	if len(alerts) > maxAlerts {
		alerts = alerts[len(alerts)-maxAlerts:]
	}
	m.plaintext = alerts
	return m
}

func (m model) hostOrIP(ip string) string {
	if name, ok := m.hostnames[ip]; ok {
		return name
	}
	return ip
}

func isLoopback(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.IsLoopback()
}

// everything the analytics panels show, as of m.now
func (m model) analytics() Analytics {
	a := Analytics{FlowsAvailable: m.nstatOK}
	now := m.now.Unix()

	bySec := make(map[int64]trafficSecond)
	procs := make(map[processKey]usage)
	dests := make(map[string]destUsage)
	for _, t := range m.traffic {
		if t.sec < now-trafficWindow || t.sec > now {
			continue
		}
		bySec[t.sec] = t
		for k, u := range t.procs {
			total := procs[k]
			procs[k] = usage{up: total.up + u.up, down: total.down + u.down}
		}
		for ip, d := range t.dests {
			total := dests[ip]
			total.bytes += d.bytes
			total.packets += d.packets
			for _, p := range d.procs {
				if len(total.procs) < 3 && !slices.Contains(total.procs, p) {
					total.procs = append(total.procs, p)
				}
			}
			dests[ip] = total
		}
	}

	// the current second isn't over yet, so the chart stops at the last full one
	for s := now - trafficWindow; s < now; s++ {
		t := bySec[s]
		a.Bandwidth = append(a.Bandwidth, BandwidthPoint{Time: s * 1000, Up: t.up, Down: t.down})
	}

	for k, u := range procs {
		a.TopProcesses = append(a.TopProcesses, ProcessUsage{ProcessName: k.ProcessName, Pid: k.Pid, Up: u.up, Down: u.down})
	}
	slices.SortFunc(a.TopProcesses, func(x, y ProcessUsage) int {
		return cmp.Or(cmp.Compare(y.Up+y.Down, x.Up+x.Down), cmp.Compare(x.ProcessName, y.ProcessName))
	})
	a.TopProcesses = a.TopProcesses[:min(len(a.TopProcesses), topProcesses)]

	for ip, d := range dests {
		dest := Destination{IP: ip, Host: m.hostnames[ip], Bytes: d.bytes, Packets: d.packets}
		for _, p := range d.procs {
			dest.Processes = append(dest.Processes, p.ProcessName)
		}
		a.TopDestinations = append(a.TopDestinations, dest)
	}
	slices.SortFunc(a.TopDestinations, func(x, y Destination) int {
		return cmp.Or(cmp.Compare(y.Bytes, x.Bytes), cmp.Compare(x.IP, y.IP))
	})
	a.TopDestinations = a.TopDestinations[:min(len(a.TopDestinations), topDests)]

	for i := len(m.plaintext) - 1; i >= 0 && len(a.Plaintext) < shownAlerts; i-- {
		al := m.plaintext[i]
		a.Plaintext = append(a.Plaintext, PlaintextAlert{
			ProcessName: al.proc.ProcessName,
			Pid:         al.proc.Pid,
			Remote:      m.hostOrIP(al.remoteIP),
			RemotePort:  al.remotePort,
			Protocol:    al.protocol,
			Outgoing:    al.outgoing,
			Snippet:     al.snippet,
			Count:       al.count,
			Time:        al.time,
		})
	}

	a.Latency = m.latency()
	a.OpenPorts = m.openPorts()
	return a
}

// round trip times of each process's open tcp connections
func (m model) latency() []ProcessLatency {
	byProc := make(map[processKey][]flow)
	for _, f := range m.flows {
		if f.removed || f.listening || f.sock.Proto != TCP || f.rttMs <= 0 || f.sock.DestPort == "0" {
			continue
		}
		k := processOf(f.sock)
		byProc[k] = append(byProc[k], f)
	}

	var out []ProcessLatency
	for k, fs := range byProc {
		slices.SortFunc(fs, func(x, y flow) int { return cmp.Or(cmp.Compare(x.rttMs, y.rttMs), cmp.Compare(x.id, y.id)) })
		median := fs[len(fs)/2].rttMs
		if len(fs)%2 == 0 {
			median = (fs[len(fs)/2-1].rttMs + fs[len(fs)/2].rttMs) / 2
		}
		slowest := fs[len(fs)-1]
		out = append(out, ProcessLatency{
			ProcessName: k.ProcessName,
			Pid:         k.Pid,
			MedianMs:    median,
			Connections: len(fs),
			Slowest:     m.hostOrIP(slowest.sock.DestIP),
			SlowestMs:   slowest.rttMs,
		})
	}
	slices.SortFunc(out, func(x, y ProcessLatency) int {
		return cmp.Or(cmp.Compare(y.MedianMs, x.MedianMs), cmp.Compare(x.ProcessName, y.ProcessName))
	})
	return out[:min(len(out), topLatency)]
}

// sockets waiting for someone else to connect: tcp listeners, plus udp bound
// to a fixed (non-ephemeral) port
func (m model) openPorts() []OpenPort {
	type portKey struct {
		proc  processKey
		proto int32
		port  string
	}
	ports := make(map[portKey]OpenPort)
	for _, f := range m.flows {
		if f.removed {
			continue
		}
		port, _ := strconv.Atoi(f.sock.SrcPort)
		udpBound := f.sock.Proto == UDP && f.sock.DestPort == "0" && port > 0 && port < 49152
		if !udpBound && !(f.sock.Proto == TCP && f.listening) {
			continue
		}

		address := f.localIP
		if address == "" || address == "0.0.0.0" || address == "::" {
			address = "all interfaces"
		}
		k := portKey{proc: processOf(f.sock), proto: f.sock.Proto, port: f.sock.SrcPort}
		op, seen := ports[k]
		if !seen {
			proto := "UDP"
			if f.sock.Proto == TCP {
				proto = "TCP"
			}
			op = OpenPort{ProcessName: k.proc.ProcessName, Pid: k.proc.Pid, Protocol: proto, Port: f.sock.SrcPort, Address: address}
		} else if address == "all interfaces" {
			op.Address = address // ipv4 + ipv6 sockets on one port - show the widest
		}
		op.Network = op.Network || !isLoopback(f.localIP)
		ports[k] = op
	}

	out := slices.Collect(maps.Values(ports))
	slices.SortFunc(out, func(x, y OpenPort) int {
		xp, _ := strconv.Atoi(x.Port)
		yp, _ := strconv.Atoi(y.Port)
		// reachable from other machines first, then by port
		if x.Network != y.Network {
			if x.Network {
				return -1
			}
			return 1
		}
		return cmp.Or(cmp.Compare(xp, yp), cmp.Compare(x.Protocol, y.Protocol), cmp.Compare(x.ProcessName, y.ProcessName))
	})
	return out[:min(len(out), maxOpenPorts)]
}

// what the process list + detail show about each process, besides its packets
type processSummary struct {
	up, down    int64
	connections int
	rtts        []float64
	hosts       []string
}

const maxConnectedTo = 8

func (s processSummary) fill(v *SocketView) {
	v.Up, v.Down, v.Connections, v.ConnectedTo = s.up, s.down, s.connections, s.hosts
	if n := len(s.rtts); n > 0 {
		rtts := slices.Sorted(slices.Values(s.rtts))
		v.MedianMs = rtts[n/2]
		if n%2 == 0 {
			v.MedianMs = (rtts[n/2-1] + rtts[n/2]) / 2
		}
	}
}

func (m model) processSummaries() map[processKey]processSummary {
	out := make(map[processKey]processSummary)
	now := m.now.Unix()
	for _, t := range m.traffic {
		if t.sec < now-trafficWindow || t.sec > now {
			continue
		}
		for k, u := range t.procs {
			s := out[k]
			s.up += u.up
			s.down += u.down
			out[k] = s
		}
	}

	for _, f := range m.flows {
		if f.removed || f.listening || f.sock.DestPort == "0" {
			continue
		}
		k := processOf(f.sock)
		s := out[k]
		s.connections++
		if f.sock.Proto == TCP && f.rttMs > 0 {
			s.rtts = append(s.rtts, f.rttMs)
		}
		if host := m.hostOrIP(f.sock.DestIP); !slices.Contains(s.hosts, host) {
			s.hosts = append(s.hosts, host)
		}
		out[k] = s
	}

	// named hosts read better than ips, so they go first
	named := func(h string) bool { return net.ParseIP(h) == nil }
	for k, s := range out {
		slices.SortFunc(s.hosts, func(a, b string) int {
			if named(a) != named(b) {
				if named(a) {
					return -1
				}
				return 1
			}
			return cmp.Compare(a, b)
		})
		s.hosts = s.hosts[:min(len(s.hosts), maxConnectedTo)]
		out[k] = s
	}
	return out
}
