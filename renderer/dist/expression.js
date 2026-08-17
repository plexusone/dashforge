const EXPR_RE = /\$\{([^}]+)\}/g;
export function containsExpression(s) {
    EXPR_RE.lastIndex = 0;
    return EXPR_RE.test(s);
}
export function extractPaths(s) {
    const paths = [];
    EXPR_RE.lastIndex = 0;
    let match;
    while ((match = EXPR_RE.exec(s)) !== null) {
        paths.push(match[1].trim());
    }
    return paths;
}
function resolvePath(path, ctx) {
    const segments = path.split('.');
    let current = ctx;
    for (const seg of segments) {
        if (current === null || current === undefined) {
            throw new Error(`Cannot resolve path "${path}": reached null/undefined at "${seg}"`);
        }
        if (typeof current === 'object') {
            const obj = current;
            if (seg in obj) {
                current = obj[seg];
            }
            else if (Array.isArray(current)) {
                const idx = Number(seg);
                if (Number.isInteger(idx) && idx >= 0 && idx < current.length) {
                    current = current[idx];
                }
                else {
                    throw new Error(`Cannot resolve path "${path}": "${seg}" not found`);
                }
            }
            else {
                throw new Error(`Cannot resolve path "${path}": "${seg}" not found`);
            }
        }
        else {
            throw new Error(`Cannot resolve path "${path}": cannot traverse into ${typeof current}`);
        }
    }
    return current;
}
export function evaluateExpression(expr, ctx) {
    if (!containsExpression(expr)) {
        return expr;
    }
    const fullMatch = expr.match(/^\$\{([^}]+)\}$/);
    if (fullMatch) {
        return resolvePath(fullMatch[1].trim(), ctx);
    }
    return expr.replace(EXPR_RE, (_match, path) => {
        const value = resolvePath(path.trim(), ctx);
        return String(value);
    });
}
//# sourceMappingURL=expression.js.map