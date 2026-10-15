// the window: header, tabs, and the processes tab. the overview tab's panels
// live in analytics.ts
import {getSnapshot} from './store';
import type {Packet, Snapshot} from './store';
import {main} from '../wailsjs/go/models';
import {SelectProcess} from '../wailsjs/go/main/App';
import {renderAnalytics} from './analytics';
import {el, empty, keyed} from './dom';
import {fmtBytes, fmtMs, fmtRate, fmtTime} from './format';

type View = 'overview' | 'processes';

const errorEl = document.getElementById('error')!;
const statusEl = document.getElementById('status')!;
const ratesEl = document.getElementById('rates')!;
const listEl = document.getElementById('proc-list')!;
const detailEl = document.getElementById('proc-detail')!;
const tabs = Array.from(document.querySelectorAll<HTMLButtonElement>('.tab'));
const views: Record<View, HTMLElement> = {
    overview: document.getElementById('view-overview')!,
    processes: document.getElementById('view-processes')!,
};

// --- tabs ---

let view: View = 'overview';
try {
    const saved = localStorage.getItem('ntwatch.view');
    if (saved === 'overview' || saved === 'processes') view = saved;
} catch {
    // no storage - just start on the overview
}

function showView(v: View) {
    view = v;
    for (const t of tabs) t.setAttribute('aria-selected', String(t.dataset.view === v));
    for (const [name, section] of Object.entries(views)) section.hidden = name !== v;
    try {
        localStorage.setItem('ntwatch.view', v);
    } catch {
        // fine, it just won't be remembered
    }
}

for (const t of tabs) t.addEventListener('click', () => showView(t.dataset.view as View));
showView(view);

function openProcess(pid: string, name: string) {
    SelectProcess(pid, name);
    showView('processes');
}

function isSelected(snap: Snapshot, sock: main.SocketView): boolean {
    return sock.pid === snap.selected?.pid && sock.processName === snap.selected?.processName;
}

// --- header ---

function renderHeader(snap: Snapshot) {
    errorEl.textContent = snap.err;

    const status = el('span', 'status');
    if (snap.capturing) {
        status.append(el('i', 'dot good'), document.createTextNode(`Capturing on ${snap.interface}`));
    } else {
        status.append(el('i', 'dot off'), document.createTextNode('Not capturing'));
    }
    statusEl.replaceChildren(status);

    const points = snap.analytics?.bandwidth ?? [];
    const latest = points[points.length - 1];
    ratesEl.replaceChildren(
        keyed('s1', `↓ ${fmtRate(latest?.down ?? 0)}`),
        keyed('s2', `↑ ${fmtRate(latest?.up ?? 0)}`),
    );
}

// --- process list ---

// the process being viewed is pinned above the rest, which are sorted by how
// recently they were active. rows are kept and moved rather than rebuilt, so a
// click isn't lost to a redraw, and they slide to new spots instead of jumping
const pinnedEl = document.getElementById('proc-pinned')!;
const rows = new Map<string, HTMLLIElement>();
const rowKey = (sock: {pid: string; processName: string}) => `${sock.pid}\u0000${sock.processName}`;

// the unpinned order only changes once a second (or when a process comes or
// goes), so busy processes don't swap places on every snapshot
const resortEvery = 1000;
let order: string[] = [];
let sortedAt = 0;

const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
const noOthers = el('li');
noOthers.append(empty('No other processes yet.'));

function createRow(sock: main.SocketView): HTMLLIElement {
    const li = el('li');
    const button = el('button', 'proc-item');
    button.append(
        el('span', 'proc-name', sock.processName),
        el('span', 'proc-pid', sock.pid),
        el('span', 'proc-data num'),
        el('span', 'proc-meta'),
    );
    button.addEventListener('click', () => openProcess(sock.pid, sock.processName));
    li.append(button);
    return li;
}

function updateRow(li: HTMLLIElement, sock: main.SocketView, selected: boolean) {
    const button = li.firstElementChild as HTMLButtonElement;
    button.classList.toggle('selected', selected);
    button.setAttribute('aria-current', String(selected));
    button.querySelector('.proc-data')!.textContent = fmtBytes(sock.up + sock.down);
    const conns = sock.connections === 1 ? '1 connection' : `${sock.connections} connections`;
    button.querySelector('.proc-meta')!.textContent = `↓ ${fmtBytes(sock.down)} · ↑ ${fmtBytes(sock.up)} · ${conns}`;
}

// puts each container's rows in the given order, sliding rows that moved
function arrange(layout: [HTMLElement, HTMLLIElement[]][]) {
    const changed = layout.some(([container, want]) =>
        container.children.length !== want.length || want.some((li, i) => container.children[i] !== li));
    if (!changed) return;

    const before = new Map<HTMLLIElement, number>();
    for (const li of rows.values()) {
        if (li.isConnected) before.set(li, li.getBoundingClientRect().top);
    }
    for (const [container, want] of layout) container.replaceChildren(...want);
    if (reduceMotion.matches) return;

    for (const [li, top] of before) {
        const delta = top - li.getBoundingClientRect().top;
        if (li.isConnected && Math.abs(delta) > 1) {
            li.animate([{transform: `translateY(${delta}px)`}, {transform: 'none'}], {duration: 250, easing: 'ease-out'});
        }
    }
}

function renderList(snap: Snapshot) {
    const sockets = snap.recentSockets ?? [];
    if (sockets.length === 0) {
        rows.clear();
        order = [];
        pinnedEl.replaceChildren();
        const li = el('li');
        li.append(empty('Waiting for traffic…'));
        listEl.replaceChildren(li);
        return;
    }

    const selected = snap.selected ? rowKey(snap.selected) : '';
    const live = new Map(sockets.map((sock) => [rowKey(sock), sock]));
    for (const [k, sock] of live) {
        let li = rows.get(k);
        if (!li) {
            li = createRow(sock);
            rows.set(k, li);
        }
        updateRow(li, sock, k === selected);
    }
    for (const [k, li] of rows) {
        if (!live.has(k)) {
            li.remove();
            rows.delete(k);
        }
    }

    const wanted = sockets.map(rowKey).filter((k) => k !== selected);
    const sameMembers = wanted.length === order.length && wanted.every((k) => order.includes(k));
    const now = performance.now();
    if (!sameMembers || now - sortedAt >= resortEvery) {
        order = wanted;
        sortedAt = now;
    }

    const others = order.map((k) => rows.get(k)!);
    if (others.length === 0) others.push(noOthers);
    const pinned = rows.get(selected);
    arrange([[pinnedEl, pinned ? [pinned] : []], [listEl, others]]);
}

// --- selected process ---

function stat(label: string, value: string): HTMLDivElement {
    const s = el('div', 'stat');
    s.append(el('div', 'label', label), el('div', 'value', value));
    return s;
}

function packetRow(p: Packet): HTMLTableRowElement {
    const remoteIP = p.outgoing ? p.destIP : p.sourceIP;
    const remotePort = p.outgoing ? p.destPort : p.sourcePort;
    const localPort = p.outgoing ? p.sourcePort : p.destPort;

    const remote = el('td');
    if (p.remoteHost) {
        remote.append(el('span', '', `${p.remoteHost}:${remotePort}`), el('span', 'mono muted ip-after', remoteIP));
    } else {
        remote.append(el('span', 'mono', `${remoteIP}:${remotePort}`));
    }
    const direction = el('td');
    direction.append(p.outgoing ? keyed('s2', 'Sent') : keyed('s1', 'Received'));

    const tr = el('tr');
    tr.append(
        el('td', 'num', fmtTime(p.time)),
        direction,
        remote,
        el('td', 'mono', localPort),
        el('td', '', p.protocol),
        el('td', 'right num', fmtBytes(p.size)),
    );
    return tr;
}

function renderDetail(snap: Snapshot) {
    const sel = snap.selected;
    if (!sel) {
        detailEl.replaceChildren(empty('Waiting for traffic…'));
        return;
    }
    const flows = snap.analytics?.flowsAvailable ?? false;

    const title = el('div', 'detail-title');
    title.append(el('h1', '', sel.processName), el('span', 'muted', `pid ${sel.pid}`));

    const stats = el('div', 'stats');
    stats.append(
        stat('Received, last minute', fmtBytes(sel.down)),
        stat('Sent, last minute', fmtBytes(sel.up)),
        stat('Open connections', flows ? String(sel.connections) : '—'),
        stat('Median latency', sel.medianMs > 0 ? fmtMs(sel.medianMs) : '—'),
    );

    const connected = el('div', 'connected');
    connected.append(el('span', 'label', 'Connected to'));
    const hosts = sel.connectedTo ?? [];
    if (hosts.length === 0) {
        connected.append(el('span', 'muted', flows ? 'nothing right now' : '—'));
    } else {
        for (const h of hosts) connected.append(el('span', 'chip', h));
    }

    const packetsHead = el('div', 'card-head');
    packetsHead.append(el('h2', '', 'Recent packets'), el('span', 'sub', 'newest first'));
    const packets = sel.packets ?? [];
    let table: HTMLElement;
    if (packets.length === 0) {
        table = empty('No packets yet.');
    } else {
        table = el('table', 'data-table');
        const head = el('tr');
        head.append(
            el('th', '', 'Time'), el('th', '', 'Direction'), el('th', '', 'Remote'),
            el('th', '', 'Local port'), el('th', '', 'Protocol'), el('th', 'right', 'Size'),
        );
        table.append(head, ...packets.map(packetRow));
    }

    detailEl.replaceChildren(title, stats, connected, packetsHead, table);
}

// --- keyboard: up/down moves through the process list ---

document.addEventListener('keydown', (e) => {
    if (view !== 'processes' || (e.key !== 'ArrowUp' && e.key !== 'ArrowDown')) return;
    const snap = getSnapshot();
    const sockets = snap?.recentSockets ?? [];
    if (!snap || sockets.length === 0) return;
    e.preventDefault();

    const current = Math.max(0, sockets.findIndex((sock) => isSelected(snap, sock)));
    const next = e.key === 'ArrowUp' ? current - 1 : current + 1;
    const sock = sockets[Math.min(Math.max(next, 0), sockets.length - 1)];
    SelectProcess(sock.pid, sock.processName);
});

export function render(snap: Snapshot) {
    renderHeader(snap);
    renderList(snap);
    renderDetail(snap);

    const live = new Set((snap.recentSockets ?? []).map(rowKey));
    if (snap.analytics) {
        renderAnalytics(
            snap.analytics,
            (pid, name) => live.has(`${pid}\u0000${name}`),
            openProcess,
        );
    }
}
