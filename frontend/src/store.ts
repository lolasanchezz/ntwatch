// holds the latest snapshot from the engine. no DOM in here, so a react version
// can reuse it as-is with useSyncExternalStore(subscribe, getSnapshot)
import {EventsOn} from '../wailsjs/runtime/runtime';
import {main} from '../wailsjs/go/models';

export type Snapshot = main.Snapshot;
export type Packet = main.PacketView;

let current: Snapshot | null = null;
const listeners = new Set<(snap: Snapshot) => void>();

EventsOn('ntwatch:snapshot', (snap: Snapshot) => {
    current = snap;
    listeners.forEach((fn) => fn(snap));
});

export function subscribe(fn: (snap: Snapshot) => void): () => void {
    listeners.add(fn);
    return () => listeners.delete(fn);
}

export function getSnapshot(): Snapshot | null {
    return current;
}
