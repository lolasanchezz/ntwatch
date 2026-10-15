package main

import "maps"

// how long a closed flow sticks around, in TickMsgs, so its last few packets still match
const flowGraceTicks = 2

type flow struct {
	id        uint64
	sock      socketKey
	localIP   string
	listening bool
	rttMs     float64 // 0 until the first stats come in
	removed   bool
	removedAt int // m.timer when it was removed
}

// what a packet gets matched on. unconnected udp flows have remoteIP + remotePort "0"
type flowKey struct {
	proto      int32
	localPort  string
	remoteIP   string
	remotePort string
}

// flows keyed for matching. built fresh from the flow table, never changed after
type flowIndex map[flowKey]flow

func keyOf(sock socketKey) flowKey {
	return flowKey{proto: sock.Proto, localPort: sock.SrcPort, remoteIP: sock.DestIP, remotePort: sock.DestPort}
}

func buildFlowIndex(flows map[uint64]flow) flowIndex {
	idx := make(flowIndex, len(flows))
	for _, f := range flows {
		if old, ok := idx[keyOf(f.sock)]; !ok || betterFlow(f, old) {
			idx[keyOf(f.sock)] = f
		}
	}
	return idx
}

// when two flows share a key (lots of processes listen on udp 5353), open beats
// closed, then the oldest wins - so the pick doesn't flip around between rebuilds
func betterFlow(a, b flow) bool {
	if a.removed != b.removed {
		return !a.removed
	}
	return a.id < b.id
}

func (m model) applyFlow(msg flowMsg) model {
	flows := maps.Clone(m.flows)
	old, existed := flows[msg.id]
	if msg.removed {
		if !existed {
			return m
		}
		old.removed = true
		old.removedAt = m.timer
		flows[msg.id] = old
	} else {
		flows[msg.id] = flow{id: msg.id, sock: msg.sock, localIP: msg.localIP, listening: msg.listening, rttMs: old.rttMs}
	}
	m.flows = flows
	// the index only cares about addresses, so skip the rebuild if those didn't move
	if msg.removed || !existed || old.sock != msg.sock {
		m.flowIndex = buildFlowIndex(flows)
	}
	return m
}

func (m model) applyFlowStats(msg flowStatsMsg) model {
	f, ok := m.flows[msg.id]
	if !ok {
		return m
	}
	f.rttMs = msg.rttMs
	flows := maps.Clone(m.flows)
	flows[msg.id] = f
	m.flows = flows
	return m
}

func (m model) pruneFlows() model {
	expired := func(_ uint64, f flow) bool { return f.removed && m.timer-f.removedAt >= flowGraceTicks }
	flows := maps.Clone(m.flows)
	maps.DeleteFunc(flows, expired)
	if len(flows) == len(m.flows) {
		return m
	}
	m.flows = flows
	m.flowIndex = buildFlowIndex(flows)
	return m
}
