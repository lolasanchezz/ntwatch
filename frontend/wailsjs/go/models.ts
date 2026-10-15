export namespace main {
	
	export class OpenPort {
	    processName: string;
	    pid: string;
	    protocol: string;
	    port: string;
	    address: string;
	    network: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OpenPort(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.processName = source["processName"];
	        this.pid = source["pid"];
	        this.protocol = source["protocol"];
	        this.port = source["port"];
	        this.address = source["address"];
	        this.network = source["network"];
	    }
	}
	export class ProcessLatency {
	    processName: string;
	    pid: string;
	    medianMs: number;
	    connections: number;
	    slowest: string;
	    slowestMs: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessLatency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.processName = source["processName"];
	        this.pid = source["pid"];
	        this.medianMs = source["medianMs"];
	        this.connections = source["connections"];
	        this.slowest = source["slowest"];
	        this.slowestMs = source["slowestMs"];
	    }
	}
	export class PlaintextAlert {
	    processName: string;
	    pid: string;
	    remote: string;
	    remotePort: string;
	    protocol: string;
	    outgoing: boolean;
	    snippet: string;
	    count: number;
	    time: number;
	
	    static createFrom(source: any = {}) {
	        return new PlaintextAlert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.processName = source["processName"];
	        this.pid = source["pid"];
	        this.remote = source["remote"];
	        this.remotePort = source["remotePort"];
	        this.protocol = source["protocol"];
	        this.outgoing = source["outgoing"];
	        this.snippet = source["snippet"];
	        this.count = source["count"];
	        this.time = source["time"];
	    }
	}
	export class Destination {
	    ip: string;
	    host: string;
	    bytes: number;
	    packets: number;
	    processes: string[];
	
	    static createFrom(source: any = {}) {
	        return new Destination(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.host = source["host"];
	        this.bytes = source["bytes"];
	        this.packets = source["packets"];
	        this.processes = source["processes"];
	    }
	}
	export class ProcessUsage {
	    processName: string;
	    pid: string;
	    up: number;
	    down: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessUsage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.processName = source["processName"];
	        this.pid = source["pid"];
	        this.up = source["up"];
	        this.down = source["down"];
	    }
	}
	export class BandwidthPoint {
	    time: number;
	    up: number;
	    down: number;
	
	    static createFrom(source: any = {}) {
	        return new BandwidthPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.up = source["up"];
	        this.down = source["down"];
	    }
	}
	export class Analytics {
	    flowsAvailable: boolean;
	    bandwidth: BandwidthPoint[];
	    topProcesses: ProcessUsage[];
	    topDestinations: Destination[];
	    plaintext: PlaintextAlert[];
	    latency: ProcessLatency[];
	    openPorts: OpenPort[];
	
	    static createFrom(source: any = {}) {
	        return new Analytics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flowsAvailable = source["flowsAvailable"];
	        this.bandwidth = this.convertValues(source["bandwidth"], BandwidthPoint);
	        this.topProcesses = this.convertValues(source["topProcesses"], ProcessUsage);
	        this.topDestinations = this.convertValues(source["topDestinations"], Destination);
	        this.plaintext = this.convertValues(source["plaintext"], PlaintextAlert);
	        this.latency = this.convertValues(source["latency"], ProcessLatency);
	        this.openPorts = this.convertValues(source["openPorts"], OpenPort);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class PacketView {
	    protocol: string;
	    sourceIP: string;
	    destIP: string;
	    sourcePort: string;
	    destPort: string;
	    size: number;
	    time: number;
	    outgoing: boolean;
	    remoteHost: string;
	    processName: string;
	    system: boolean;
	    pid: string;
	
	    static createFrom(source: any = {}) {
	        return new PacketView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.protocol = source["protocol"];
	        this.sourceIP = source["sourceIP"];
	        this.destIP = source["destIP"];
	        this.sourcePort = source["sourcePort"];
	        this.destPort = source["destPort"];
	        this.size = source["size"];
	        this.time = source["time"];
	        this.outgoing = source["outgoing"];
	        this.remoteHost = source["remoteHost"];
	        this.processName = source["processName"];
	        this.system = source["system"];
	        this.pid = source["pid"];
	    }
	}
	
	
	
	export class SocketView {
	    processName: string;
	    pid: string;
	    packets: PacketView[];
	    up: number;
	    down: number;
	    connections: number;
	    medianMs: number;
	    connectedTo: string[];
	
	    static createFrom(source: any = {}) {
	        return new SocketView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.processName = source["processName"];
	        this.pid = source["pid"];
	        this.packets = this.convertValues(source["packets"], PacketView);
	        this.up = source["up"];
	        this.down = source["down"];
	        this.connections = source["connections"];
	        this.medianMs = source["medianMs"];
	        this.connectedTo = source["connectedTo"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Snapshot {
	    interface: string;
	    capturing: boolean;
	    err: string;
	    recentSockets: SocketView[];
	    selected?: SocketView;
	    newPackets: PacketView[];
	    analytics: Analytics;
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.interface = source["interface"];
	        this.capturing = source["capturing"];
	        this.err = source["err"];
	        this.recentSockets = this.convertValues(source["recentSockets"], SocketView);
	        this.selected = this.convertValues(source["selected"], SocketView);
	        this.newPackets = this.convertValues(source["newPackets"], PacketView);
	        this.analytics = this.convertValues(source["analytics"], Analytics);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

