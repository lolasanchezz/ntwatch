package main

import (
	"cmp"
	"maps"
	"net"
	"slices"
	"time"

	"github.com/google/gopacket/pcap"
)

const (
	uiInterval        = 100 * time.Millisecond
	maxPacketsPerProc = 50                     // only the newest few get shown, no need to keep the rest forever
	maxUnmatched      = 100                    // same idea
	maxNewPackets     = 1000                   // per ui tick, so a burst of traffic can't flood the views
	scanGap           = 250 * time.Millisecond // min time between socket scans when they're wanted back to back
	matchRetryDelay   = 150 * time.Millisecond // how long an unmatched packet waits for its flow to show up
)

type TickMsg time.Time
type uiTickMsg time.Time
type scanMsg struct{}
type errMsg struct{ err error }

// asks the engine for the current state, answered on reply
type snapshotRequestMsg struct{ reply chan Snapshot }

// packets are kept per process, across all of its connections
type processKey struct{ Pid, ProcessName string }

func processOf(sock socketKey) processKey {
	return processKey{Pid: sock.Pid, ProcessName: sock.ProcessName}
}

// picks which process gets the detailed view
type selectProcessMsg processKey

func doTick() Cmd {
	return Tick(time.Second*2, func(t time.Time) Msg { return TickMsg(t) })
}

func doUITick() Cmd {
	return Tick(uiInterval, func(t time.Time) Msg { return uiTickMsg(t) })
}

// all of the app's state. Update never changes a model in place, it hands back a new one
type model struct {
	handle            *pcap.Handle
	mac               net.HardwareAddr
	err               error
	unmatchedPacket   []PacketInfo
	timer             int
	tableNum          int
	socketNum         int
	selectedNum       int
	selected          processKey // zero value means "whichever is newest"
	socketTable       socketMap
	flows             map[uint64]flow // from network statistics, by flow id
	flowIndex         flowIndex       // the same flows, keyed for matching
	nstatOK           bool            // flows are coming in, so unmatched packets don't need rescans
	nstatErr          error
	scanning          bool // a socket scan is running (or waiting out scanGap)
	scanWanted        bool // something asked for a scan while one was running
	matchedPacketsTbl map[processKey][]PacketInfo
	recentProcesses   []processKey
	newPackets        []PacketView      // everything seen since the last ui tick
	now               time.Time         // as of the last ui tick
	traffic           []trafficSecond   // the last minute, a bucket per second, oldest first
	hostnames         map[string]string // ip -> name, from dns answers + tls server names
	hostOrder         []string          // hostnames' ips, oldest first, for forgetting
	plaintext         []plaintextAlert  // newest last
	publish           func(Snapshot)    // where snapshots go - the tui or the wails window
}

func initialModel(publish func(Snapshot)) model {
	tableNum := 20
	socketNum := 3
	selectedNum := 30
	return model{
		matchedPacketsTbl: make(map[processKey][]PacketInfo),
		flows:             make(map[uint64]flow),
		flowIndex:         make(flowIndex),
		tableNum:          tableNum,
		socketNum:         socketNum,
		selectedNum:       selectedNum,
		publish:           publish,
	}
}

func (m model) Init() []Cmd {
	return []Cmd{func() Msg { return wireInit() }, func() Msg { return nstatStart() }, doTick(), doUITick()}
}

func (m model) Update(msg Msg) (model, []Cmd) {
	var cmds []Cmd

	switch msg := msg.(type) {
	case wireInitMsg:
		m.handle = msg.handle
		m.mac = msg.mac
		cmds = append(cmds, sendPacketCmd(m.handle))
	case WireDataMsg:
		cmds = append(cmds, readPacketCmd(msg.data, m.mac), sendPacketCmd(m.handle))
	case packetInfoMsg:
		if !msg.retried {
			m = m.learnHostnames(msg.data.hostnames)
		}
		cmds = append(cmds, matchPacketsCmd(m.flowIndex, m.socketTable, msg))
	case TickMsg:
		m.timer++
		m = m.pruneFlows()
		//refresh socket table
		var scan Cmd
		m, scan = m.requestScan()
		cmds = append(cmds, doTick(), scan)
	case socketMapMsg:
		m.socketTable = msg.data
		m.scanning = false
		if m.scanWanted {
			// someone wanted a fresher table while that scan ran - do one more, after a gap
			m.scanWanted = false
			m.scanning = true
			cmds = append(cmds, Tick(scanGap, func(time.Time) Msg { return scanMsg{} }))
		}
	case scanMsg:
		cmds = append(cmds, getCStructCmd())
	case nstatStartedMsg:
		m.nstatOK = true
		cmds = append(cmds, nstatNextCmd())
	case nstatFailedMsg:
		m.nstatErr = msg.err
	case flowMsg:
		m = m.applyFlow(msg)
		cmds = append(cmds, nstatNextCmd())
	case flowStatsMsg:
		m = m.applyFlowStats(msg)
		cmds = append(cmds, nstatNextCmd())
	case matchedPacketMsg:
		if msg.data.process == (socketKey{}) && !msg.data.system && !msg.retried {
			// a new connection's first packets can beat its flow's description -
			// give it a moment, then try once more before calling it unmatched
			retry := packetInfoMsg{data: msg.data.packet, retried: true}
			return m, append(cmds, Tick(matchRetryDelay, func(time.Time) Msg { return retry }))
		}
		proc := processOf(msg.data.process)
		view := m.toPacketView(proc, msg.data.packet)
		view.System = msg.data.system
		m.newPackets = push(m.newPackets, view, maxNewPackets)
		m = m.countTraffic(proc, msg.data.packet, msg.data.system)
		switch {
		case msg.data.system:
			// no socket could own it, so a fresh scan wouldn't help
		case msg.data.process == (socketKey{}): //unmatched
			m.unmatchedPacket = push(m.unmatchedPacket, msg.data.packet, maxUnmatched)
			// flows arrive as they open, so a rescan only helps without them
			if !m.nstatOK {
				var scan Cmd
				m, scan = m.requestScan()
				cmds = append(cmds, scan)
			}
		default:
			m = m.addMatched(proc, msg.data.packet)
			if msg.data.packet.readable != "" {
				m = m.notePlaintext(proc, msg.data.packet)
			}
		}
	case uiTickMsg:
		m.now = time.Time(msg)
		snap := m.snapshot()
		snap.NewPackets = m.newPackets
		m.newPackets = nil // the snapshot owns the old slice now
		publish := m.publish
		cmds = append(cmds, doUITick(), func() Msg { publish(snap); return nil })
	case snapshotRequestMsg:
		snap := m.snapshot()
		cmds = append(cmds, func() Msg { msg.reply <- snap; return nil })
	case selectProcessMsg:
		m.selected = processKey(msg)
	case errMsg:
		m.err = msg.err
	}
	return m, cmds
}

// at most one socket scan runs at a time - asking while one is running just
// marks that another is wanted once it finishes
func (m model) requestScan() (model, Cmd) {
	if m.scanning {
		m.scanWanted = true
		return m, nil
	}
	m.scanning = true
	return m, getCStructCmd()
}

func (m model) addMatched(proc processKey, packet PacketInfo) model {
	// maps are shared between copies of the model, so write to a fresh one
	tbl := maps.Clone(m.matchedPacketsTbl)
	tbl[proc] = push(tbl[proc], packet, maxPacketsPerProc)

	// recent processes are kept least -> most recently active, so this one moves
	// to the end. (it's usually there already - nothing to do then)
	recent := m.recentProcesses
	if n := len(recent); n == 0 || recent[n-1] != proc {
		if i := slices.Index(recent, proc); i >= 0 {
			recent = slices.Delete(slices.Clone(recent), i, i+1)
		} else if len(recent) >= m.tableNum {
			// full - drop the least recently active process and its packets,
			// but never the one being looked at
			j := max(0, slices.IndexFunc(recent, func(p processKey) bool { return p != m.selected }))
			delete(tbl, recent[j])
			recent = slices.Delete(slices.Clone(recent), j, j+1)
		}
		m.recentProcesses = append(slices.Clip(recent), proc)
	}

	// the first process seen becomes the selection, so the detail always shows
	// one process instead of whichever was most recently active
	if m.selected == (processKey{}) {
		m.selected = proc
	}

	m.matchedPacketsTbl = tbl
	return m
}

// push appends v, dropping the oldest values past max. appending only ever writes past
// the end of the old slice, so older models never see their contents change
func push[T any](s []T, v T, max int) []T {
	if len(s) >= max {
		s = s[len(s)-max+1:]
	}
	return append(s, v)
}

// snapshot copies the parts of the model the views care about, most recently active first
func (m model) snapshot() Snapshot {
	snap := Snapshot{Interface: captureInterface, Capturing: m.handle != nil && m.err == nil}
	if m.err != nil {
		snap.Err = m.err.Error()
	} else if m.nstatErr != nil {
		snap.Err = m.nstatErr.Error() + " - matching with socket scans only"
	}
	summaries := m.processSummaries()
	for i := len(m.recentProcesses) - 1; i >= 0; i-- {
		view := m.processView(m.recentProcesses[i], m.socketNum)
		summaries[m.recentProcesses[i]].fill(&view)
		snap.RecentSockets = append(snap.RecentSockets, view)
	}

	// the selected process (or, if it's somehow gone, the most recently active)
	if len(m.recentProcesses) > 0 {
		shown := m.recentProcesses[len(m.recentProcesses)-1]
		if slices.Contains(m.recentProcesses, m.selected) {
			shown = m.selected
		}
		view := m.processView(shown, m.selectedNum)
		summaries[shown].fill(&view)
		snap.Selected = &view
	}
	snap.Analytics = m.analytics()
	return snap
}

// the newest n packets for a process, across all of its connections
func (m model) processView(proc processKey, n int) SocketView {
	packets := m.matchedPacketsTbl[proc]
	view := SocketView{ProcessName: proc.ProcessName, Pid: proc.Pid}
	for j := len(packets) - 1; j >= max(0, len(packets)-n); j-- {
		view.Packets = append(view.Packets, m.toPacketView(proc, packets[j]))
	}
	// a retried packet gets added late, so order by when it was captured
	slices.SortStableFunc(view.Packets, func(a, b PacketView) int { return cmp.Compare(b.Time, a.Time) })
	return view
}

func (m model) toPacketView(proc processKey, p PacketInfo) PacketView {
	remoteIP, _ := remoteOf(p)
	return PacketView{
		Outgoing:    p.outgoing,
		RemoteHost:  m.hostnames[remoteIP],
		Protocol:    p.packetProtocol.String(),
		SourceIP:    p.sourceIP,
		DestIP:      p.destIP,
		SourcePort:  p.sourcePort,
		DestPort:    p.destPort,
		Size:        p.packetSize,
		Time:        p.Timestamp.UnixMilli(),
		ProcessName: proc.ProcessName,
		Pid:         proc.Pid,
	}
}
