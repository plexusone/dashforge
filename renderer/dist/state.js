export class PageState {
    constructor() {
        this.values = {};
        this.listeners = new Map();
    }
    get(path) {
        const segments = path.split('.');
        let current = this.values;
        for (const seg of segments) {
            if (current === null || current === undefined)
                return undefined;
            if (typeof current === 'object') {
                current = current[seg];
            }
            else {
                return undefined;
            }
        }
        return current;
    }
    set(path, value) {
        const segments = path.split('.');
        if (segments.length === 1) {
            this.values[path] = value;
            this.notify(path, value);
            return;
        }
        let current = this.values;
        for (let i = 0; i < segments.length - 1; i++) {
            const seg = segments[i];
            if (!(seg in current) || typeof current[seg] !== 'object' || current[seg] === null) {
                current[seg] = {};
            }
            current = current[seg];
        }
        current[segments[segments.length - 1]] = value;
        this.notify(path, value);
    }
    toggle(path) {
        const current = this.get(path);
        if (typeof current !== 'boolean') {
            throw new Error(`Cannot toggle non-boolean value at "${path}": got ${typeof current}`);
        }
        this.set(path, !current);
    }
    subscribe(pathPrefix, fn) {
        let listeners = this.listeners.get(pathPrefix);
        if (!listeners) {
            listeners = new Set();
            this.listeners.set(pathPrefix, listeners);
        }
        listeners.add(fn);
        return () => {
            listeners.delete(fn);
            if (listeners.size === 0) {
                this.listeners.delete(pathPrefix);
            }
        };
    }
    snapshot() {
        return JSON.parse(JSON.stringify(this.values));
    }
    load(initial) {
        this.values = JSON.parse(JSON.stringify(initial));
    }
    notify(path, value) {
        for (const [prefix, listeners] of this.listeners) {
            if (path === prefix || path.startsWith(prefix + '.') || prefix === '*') {
                for (const fn of listeners) {
                    fn(path, value);
                }
            }
        }
    }
}
//# sourceMappingURL=state.js.map