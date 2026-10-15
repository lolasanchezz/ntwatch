// tiny dom helpers. text always goes in through textContent - names come off the network

const SVG_NS = 'http://www.w3.org/2000/svg';

export function el<K extends keyof HTMLElementTagNameMap>(tag: K, cls?: string, text?: string): HTMLElementTagNameMap[K] {
    const node = document.createElement(tag);
    if (cls) node.className = cls;
    if (text !== undefined) node.textContent = text;
    return node;
}

export function svg<K extends keyof SVGElementTagNameMap>(tag: K, attrs: Record<string, string | number>): SVGElementTagNameMap[K] {
    const node = document.createElementNS(SVG_NS, tag);
    for (const [k, v] of Object.entries(attrs)) node.setAttribute(k, String(v));
    return node;
}

export function empty(text: string): HTMLDivElement {
    return el('div', 'empty', text);
}

// a short colored line + label, the same keys the charts' legends use
export function keyed(cls: 's1' | 's2', label: string): HTMLSpanElement {
    const span = el('span', 'keyed');
    span.append(el('i', `key-line ${cls}`), document.createTextNode(label));
    return span;
}
