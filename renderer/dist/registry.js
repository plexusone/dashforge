const components = new Map();
export function registerComponent(type, component) {
    components.set(type, component);
}
export function getComponent(type) {
    return components.get(type);
}
export function hasComponent(type) {
    return components.has(type);
}
export function listComponents() {
    return Array.from(components.keys());
}
export function clearRegistry() {
    components.clear();
}
//# sourceMappingURL=registry.js.map