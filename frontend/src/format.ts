// number + time formatting shared by both tabs

export function fmtBytes(n: number): string {
    const units = ['B', 'KB', 'MB', 'GB'];
    let i = 0;
    while (n >= 1000 && i < units.length - 1) {
        n /= 1000;
        i++;
    }
    const digits = i === 0 || n >= 10 ? 0 : 1;
    return `${n.toFixed(digits)} ${units[i]}`;
}

export const fmtRate = (n: number) => `${fmtBytes(n)}/s`;
export const fmtMs = (ms: number) => `${ms < 10 ? ms.toFixed(1) : Math.round(ms)} ms`;
export const fmtTime = (ms: number) => new Date(ms).toLocaleTimeString([], {hour12: false});
