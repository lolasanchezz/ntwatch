// the analytics panels at the bottom. they redraw once a second (when a new
// second of bandwidth is in), not on every snapshot, so numbers don't flicker.
// names come off the network, so text always goes in through textContent
import {main} from '../wailsjs/go/models';
import {el, empty, svg} from './dom';
import {fmtBytes, fmtMs, fmtRate, fmtTime} from './format';

const bwStats = document.getElementById('bw-stats')!;
const bwChart = document.getElementById('bw-chart')!;
const bwTable = document.getElementById('bw-table') as HTMLTableElement;
const bwToggle = document.getElementById('bw-toggle')!;
const topProcs = document.getElementById('top-procs')!;
const topDests = document.getElementById('top-dests')!;
const plaintext = document.getElementById('plaintext')!;
const latency = document.getElementById('latency')!;
const ports = document.getElementById('ports')!;
const portsToggle = document.getElementById('ports-toggle')!;

let current: main.Analytics | null = null;
let canOpen: (pid: string, name: string) => boolean = () => false;
let open: (pid: string, name: string) => void = () => {};
let renderedSecond: number | null = null;
let hover: number | null = null; // index into bandwidth, while the crosshair is out
let showTable = false;
let showAllPorts = false;
const portsShown = 12;

// a round number at or above v, for the top of an axis
function niceMax(v: number): number {
    if (v <= 0) return 1000;
    const pow = Math.pow(10, Math.floor(Math.log10(v)));
    for (const step of [1, 2, 2.5, 5, 10]) {
        if (step * pow >= v) return step * pow;
    }
    return 10 * pow;
}

function warningIcon(): SVGSVGElement {
    const icon = svg('svg', {width: 12, height: 12, viewBox: '0 0 12 12', 'aria-hidden': 'true'});
    icon.append(
        svg('path', {d: 'M6 1 L11.2 10.5 H0.8 Z', fill: 'var(--status-warning)'}),
        svg('rect', {x: 5.3, y: 4.2, width: 1.4, height: 3.4, rx: 0.7, fill: '#1b2636'}),
        svg('circle', {cx: 6, cy: 9, r: 0.8, fill: '#1b2636'}),
    );
    return icon;
}

function warning(label: string): HTMLSpanElement {
    const span = el('span', 'status');
    span.append(warningIcon(), document.createTextNode(label));
    return span;
}

function nameCell(name: string, pid: string, detail?: string): HTMLDivElement {
    const cell = el('div', 'bar-name');
    if (canOpen(pid, name)) {
        // pointerdown, not click - the row may be redrawn between press and release
        cell.classList.add('link');
        cell.title = 'Show in Processes';
        cell.addEventListener('pointerdown', () => open(pid, name));
    }
    cell.append(document.createTextNode(name + ' '), el('span', 'muted', pid));
    if (detail) cell.append(el('div', 'bar-detail', detail));
    return cell;
}

// a horizontal bar: segments share its length by value, 2px gaps between them
function bar(segments: {value: number; cls: string; title: string}[], fraction: number, label: string): HTMLDivElement {
    const track = el('div', 'bar-track');
    const b = el('div', 'bar');
    b.style.width = `${(fraction * 72).toFixed(2)}%`; // leaves room for the label at the tip
    for (const s of segments) {
        if (s.value <= 0) continue;
        const seg = el('div', `seg ${s.cls}`);
        seg.style.flex = `${s.value} 1 0`;
        seg.title = s.title;
        b.append(seg);
    }
    track.append(b, el('span', 'bar-label', label));
    return track;
}

// --- bandwidth ---

const PLOT_H = 150;
const AXIS_H = 20;
const LEFT = 64;
const RIGHT = 84;
const TOP = 8;

function drawBandwidth() {
    const points = current?.bandwidth ?? [];
    const latest = points[points.length - 1];

    const minuteDown = points.reduce((sum, p) => sum + p.down, 0);
    const minuteUp = points.reduce((sum, p) => sum + p.up, 0);
    const stat = (label: string, value: string) => {
        const s = el('div', 'stat');
        s.append(el('div', 'label', label), el('div', 'value', value));
        return s;
    };
    bwStats.replaceChildren(
        stat('Download now', fmtRate(latest?.down ?? 0)),
        stat('Upload now', fmtRate(latest?.up ?? 0)),
        stat('Downloaded, last minute', fmtBytes(minuteDown)),
        stat('Uploaded, last minute', fmtBytes(minuteUp)),
    );

    bwChart.hidden = showTable;
    bwTable.hidden = !showTable;
    bwToggle.textContent = showTable ? 'Show chart' : 'Show table';
    if (showTable) {
        drawBandwidthTable(points);
    } else {
        drawBandwidthChart(points);
    }
}

function drawBandwidthTable(points: main.BandwidthPoint[]) {
    const head = el('tr');
    head.append(el('th', '', 'When'), el('th', 'right', 'Download'), el('th', 'right', 'Upload'));
    const rows = [...points].reverse().map((p, i) => {
        const tr = el('tr');
        tr.append(el('td', 'num', `${i + 1}s ago`), el('td', 'right num', fmtRate(p.down)), el('td', 'right num', fmtRate(p.up)));
        return tr;
    });
    bwTable.replaceChildren(head, ...rows);
}

function drawBandwidthChart(points: main.BandwidthPoint[]) {
    const width = Math.max(bwChart.clientWidth, 320);
    const plotW = width - LEFT - RIGHT;
    const n = points.length;
    const top = niceMax(Math.max(1, ...points.map((p) => Math.max(p.up, p.down))));
    const x = (i: number) => LEFT + (n <= 1 ? plotW : (i * plotW) / (n - 1));
    const y = (v: number) => TOP + PLOT_H - (v / top) * PLOT_H;

    const chart = svg('svg', {viewBox: `0 0 ${width} ${TOP + PLOT_H + AXIS_H}`, height: TOP + PLOT_H + AXIS_H, role: 'img'});
    chart.setAttribute('aria-label', 'Bandwidth chart');

    // gridlines + y ticks
    for (const v of [0, top / 2, top]) {
        chart.append(svg('line', {
            x1: LEFT, x2: LEFT + plotW, y1: y(v), y2: y(v),
            stroke: v === 0 ? 'var(--baseline)' : 'var(--grid)', 'stroke-width': 1,
        }));
        const tick = svg('text', {x: LEFT - 8, y: y(v) + 4, 'text-anchor': 'end', fill: 'var(--text-muted)', 'font-size': 11});
        tick.textContent = v === 0 ? '0' : fmtRate(v);
        chart.append(tick);
    }
    // x labels
    for (const [i, label, anchor] of [[0, '60s ago', 'start'], [Math.floor((n - 1) / 2), '30s', 'middle'], [n - 1, 'now', 'end']] as const) {
        if (n === 0) break;
        const t = svg('text', {x: x(i), y: TOP + PLOT_H + 15, 'text-anchor': anchor, fill: 'var(--text-muted)', 'font-size': 11});
        t.textContent = label;
        chart.append(t);
    }

    const series = [
        {key: 'down' as const, color: 'var(--series-1)', name: 'Download'},
        {key: 'up' as const, color: 'var(--series-2)', name: 'Upload'},
    ];
    for (const s of series) {
        if (n === 0) break;
        const d = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i).toFixed(1)},${y(p[s.key]).toFixed(1)}`).join(' ');
        chart.append(svg('path', {d, fill: 'none', stroke: s.color, 'stroke-width': 2, 'stroke-linejoin': 'round', 'stroke-linecap': 'round'}));
    }

    // end markers, and direct labels if they don't collide (otherwise the legend + tooltip carry it)
    if (n > 0) {
        const last = points[n - 1];
        const ends = series.map((s) => ({...s, cy: y(last[s.key]), value: last[s.key]}));
        const apart = Math.abs(ends[0].cy - ends[1].cy) >= 14;
        for (const e of ends) {
            chart.append(svg('circle', {cx: x(n - 1), cy: e.cy, r: 4, fill: e.color, stroke: 'var(--surface)', 'stroke-width': 2}));
            if (apart) {
                const t = svg('text', {x: x(n - 1) + 10, y: e.cy + 4, fill: 'var(--text-secondary)', 'font-size': 11});
                t.textContent = fmtRate(e.value);
                chart.append(t);
            }
        }
    }

    // crosshair
    let tip: HTMLDivElement | null = null;
    if (hover !== null && n > 0) {
        const i = Math.min(hover, n - 1);
        const p = points[i];
        chart.append(svg('line', {x1: x(i), x2: x(i), y1: TOP, y2: TOP + PLOT_H, stroke: 'var(--text-muted)', 'stroke-width': 1}));
        for (const s of series) {
            chart.append(svg('circle', {cx: x(i), cy: y(p[s.key]), r: 4, fill: s.color, stroke: 'var(--surface)', 'stroke-width': 2}));
        }
        tip = el('div', 'tooltip');
        tip.append(el('div', 'when', n - i === 1 ? '1 second ago' : `${n - i} seconds ago`));
        for (const s of series) {
            const row = el('div', 'row');
            const key = el('i', 'key-line');
            key.style.background = s.color;
            row.append(key, el('strong', '', fmtRate(p[s.key])), el('span', 'name', s.name));
            tip.append(row);
        }
        const left = x(i) + 12;
        tip.style.top = `${TOP}px`;
        if (left > width - 150) {
            tip.style.right = `${width - x(i) + 12}px`;
        } else {
            tip.style.left = `${left}px`;
        }
    }
    bwChart.replaceChildren(chart, ...(tip ? [tip] : []));
}

function hoverAt(clientX: number) {
    const n = current?.bandwidth?.length ?? 0;
    if (n === 0) return;
    const rect = bwChart.getBoundingClientRect();
    const plotW = Math.max(rect.width, 320) - LEFT - RIGHT;
    const px = clientX - rect.left;
    if (px < LEFT - 10 || px > LEFT + plotW + 10) {
        hover = null;
    } else {
        hover = Math.max(0, Math.min(n - 1, Math.round(((px - LEFT) / plotW) * (n - 1))));
    }
    drawBandwidth();
}

bwChart.addEventListener('pointermove', (e) => hoverAt(e.clientX));
bwChart.addEventListener('pointerleave', () => {
    hover = null;
    drawBandwidth();
});
bwChart.addEventListener('focus', () => {
    hover = (current?.bandwidth?.length ?? 1) - 1;
    drawBandwidth();
});
bwChart.addEventListener('blur', () => {
    hover = null;
    drawBandwidth();
});
bwChart.addEventListener('keydown', (e) => {
    const n = current?.bandwidth?.length ?? 0;
    if (n === 0 || (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight')) return;
    e.preventDefault();
    const step = e.key === 'ArrowLeft' ? -1 : 1;
    hover = Math.max(0, Math.min(n - 1, (hover ?? n - 1) + step));
    drawBandwidth();
});
portsToggle.addEventListener('click', () => {
    showAllPorts = !showAllPorts;
    if (current) drawPorts(current);
});
bwToggle.addEventListener('click', () => {
    showTable = !showTable;
    drawBandwidth();
});
window.addEventListener('resize', () => drawBandwidth());

// --- top processes ---

function drawTopProcesses(a: main.Analytics) {
    const procs = a.topProcesses ?? [];
    if (procs.length === 0) {
        topProcs.replaceChildren(empty('Waiting for traffic…'));
        return;
    }
    const most = Math.max(...procs.map((p) => p.up + p.down));
    topProcs.replaceChildren(...procs.map((p) => {
        const row = el('div', 'bar-row');
        row.append(
            nameCell(p.processName, p.pid, `↓ ${fmtBytes(p.down)} · ↑ ${fmtBytes(p.up)}`),
            bar([
                {value: p.down, cls: 's1', title: `Download ${fmtBytes(p.down)}`},
                {value: p.up, cls: 's2', title: `Upload ${fmtBytes(p.up)}`},
            ], (p.up + p.down) / most, fmtBytes(p.up + p.down)),
        );
        return row;
    }));
}

// --- top destinations ---

function drawTopDestinations(a: main.Analytics) {
    const dests = a.topDestinations ?? [];
    if (dests.length === 0) {
        topDests.replaceChildren(empty('Waiting for traffic…'));
        return;
    }
    const most = Math.max(...dests.map((d) => d.bytes));
    const table = el('table', 'data-table');
    const head = el('tr');
    head.append(el('th', '', 'Destination'), el('th', '', 'Processes'), el('th', 'right', 'Data'));
    table.append(head);
    for (const d of dests) {
        const tr = el('tr');
        const where = el('td');
        if (d.host) {
            where.append(el('div', '', d.host), el('div', 'mono muted', d.ip));
        } else {
            where.append(el('div', 'mono', d.ip), el('div', 'muted', 'no name seen'));
        }
        const data = el('td', 'right num', fmtBytes(d.bytes));
        const inline = el('div', 'inline-bar');
        inline.style.width = `${Math.max(4, (d.bytes / most) * 100)}%`;
        inline.style.marginLeft = 'auto';
        data.append(inline);
        tr.append(where, el('td', 'secondary', (d.processes ?? []).join(', ') || '—'), data);
        table.append(tr);
    }
    topDests.replaceChildren(table);
}

// --- plaintext ---

function drawPlaintext(a: main.Analytics) {
    const alerts = a.plaintext ?? [];
    if (alerts.length === 0) {
        plaintext.replaceChildren(empty('Nothing unencrypted seen yet.'));
        return;
    }
    plaintext.replaceChildren(...alerts.map((al) => {
        const item = el('div');
        const head = el('div', 'alert-head');
        head.append(
            warning(al.outgoing ? 'Sent' : 'Received'),
            el('strong', '', al.processName),
            el('span', 'secondary', `${al.outgoing ? 'to' : 'from'} ${al.remote}:${al.remotePort}`),
        );
        const meta = el('div', 'alert-meta', `${al.protocol} · ${al.count === 1 ? '1 packet' : `${al.count} packets`} · last at ${fmtTime(al.time)}`);
        item.append(head, meta, el('div', 'alert-snippet', al.snippet));
        return item;
    }));
}

// --- latency ---

function drawLatency(a: main.Analytics) {
    if (!a.flowsAvailable) {
        latency.replaceChildren(empty("Needs NetworkStatistics, which didn't load."));
        return;
    }
    const rows = a.latency ?? [];
    if (rows.length === 0) {
        latency.replaceChildren(empty('No round-trip times yet…'));
        return;
    }
    const most = Math.max(...rows.map((r) => r.medianMs));
    latency.replaceChildren(...rows.map((r) => {
        const conns = r.connections === 1 ? '1 connection' : `${r.connections} connections`;
        const detail = r.connections > 1 ? `${conns} · slowest ${r.slowest} ${fmtMs(r.slowestMs)}` : `${conns} · ${r.slowest}`;
        const row = el('div', 'bar-row');
        row.append(
            nameCell(r.processName, r.pid, detail),
            bar([{value: r.medianMs, cls: 's1', title: `Median ${fmtMs(r.medianMs)}`}], r.medianMs / most, fmtMs(r.medianMs)),
        );
        return row;
    }));
}

// --- open ports ---

function drawPorts(a: main.Analytics) {
    if (!a.flowsAvailable) {
        ports.replaceChildren(empty("Needs NetworkStatistics, which didn't load."));
        return;
    }
    const all = a.openPorts ?? [];
    portsToggle.hidden = all.length <= portsShown;
    portsToggle.textContent = showAllPorts ? 'Show fewer' : `Show all ${all.length}`;
    if (all.length === 0) {
        ports.replaceChildren(empty('No open ports found.'));
        return;
    }
    const list = showAllPorts ? all : all.slice(0, portsShown);
    const table = el('table', 'data-table');
    const head = el('tr');
    head.append(el('th', '', 'Port'), el('th', '', 'Process'), el('th', '', 'Bound to'), el('th', '', 'Reachable from'));
    table.append(head);
    for (const p of list) {
        const tr = el('tr');
        const who = el('td');
        who.append(document.createTextNode(p.processName + ' '), el('span', 'muted', p.pid));
        const reach = el('td');
        reach.append(p.network ? el('span', '', 'Network') : el('span', 'muted', 'This Mac only'));
        tr.append(el('td', 'mono', `${p.protocol} ${p.port}`), who, el('td', 'mono secondary', p.address), reach);
        table.append(tr);
    }
    ports.replaceChildren(table);
}

// canOpenProcess says whether a process can be shown in the processes tab,
// openProcess shows it
export function renderAnalytics(
    a: main.Analytics,
    canOpenProcess: (pid: string, name: string) => boolean,
    openProcess: (pid: string, name: string) => void,
) {
    canOpen = canOpenProcess;
    open = openProcess;
    const points = a.bandwidth ?? [];
    const second = points.length ? points[points.length - 1].time : 0;
    if (second === renderedSecond) return;
    renderedSecond = second;
    current = a;

    drawBandwidth();
    drawTopProcesses(a);
    drawTopDestinations(a);
    drawPlaintext(a);
    drawLatency(a);
    drawPorts(a);
}
