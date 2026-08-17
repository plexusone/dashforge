export function createAgentOSRuntime(config) {
    const { baseUrl, headers: extraHeaders = {}, credentials = 'include' } = config;
    function buildHeaders() {
        const h = {
            'Content-Type': 'application/json',
            ...extraHeaders,
        };
        if (config.conversationId) {
            h['X-Conversation-ID'] = config.conversationId;
        }
        return h;
    }
    return {
        config,
        async *sendMessage(content, options) {
            const resp = await fetch(`${baseUrl}/api/v1/chat/completions`, {
                method: 'POST',
                headers: buildHeaders(),
                credentials,
                body: JSON.stringify({
                    model: options?.model ?? 'default',
                    messages: [{ role: 'user', content }],
                    stream: true,
                }),
            });
            if (!resp.ok) {
                const text = await resp.text();
                const err = new Error(`Chat request failed: ${resp.status} ${text}`);
                config.onError?.(err);
                throw err;
            }
            const reader = resp.body?.getReader();
            if (!reader)
                throw new Error('No response body');
            const decoder = new TextDecoder();
            let buffer = '';
            let accumulated = '';
            const assistantMsg = {
                id: crypto.randomUUID(),
                role: 'assistant',
                content: '',
            };
            try {
                while (true) {
                    const { done, value } = await reader.read();
                    if (done)
                        break;
                    buffer += decoder.decode(value, { stream: true });
                    const lines = buffer.split('\n');
                    buffer = lines.pop() ?? '';
                    for (const line of lines) {
                        if (!line.startsWith('data: '))
                            continue;
                        const data = line.slice(6);
                        if (data === '[DONE]')
                            break;
                        try {
                            const parsed = JSON.parse(data);
                            const delta = parsed.choices?.[0]?.delta?.content;
                            if (typeof delta === 'string') {
                                accumulated += delta;
                                assistantMsg.content = accumulated;
                                config.onMessage?.(assistantMsg);
                                yield { ...assistantMsg };
                            }
                        }
                        catch {
                            // skip malformed SSE chunks
                        }
                    }
                }
            }
            finally {
                reader.releaseLock();
            }
            assistantMsg.content = accumulated;
            config.onMessage?.(assistantMsg);
            yield assistantMsg;
        },
        async listConversations() {
            const resp = await fetch(`${baseUrl}/api/v1/conversations`, {
                headers: buildHeaders(),
                credentials,
            });
            if (!resp.ok)
                throw new Error(`List conversations failed: ${resp.status}`);
            return resp.json();
        },
        async loadConversation(id) {
            const resp = await fetch(`${baseUrl}/api/v1/conversations/${encodeURIComponent(id)}/messages`, {
                headers: buildHeaders(),
                credentials,
            });
            if (!resp.ok)
                throw new Error(`Load conversation failed: ${resp.status}`);
            return resp.json();
        },
        async createConversation(title) {
            const resp = await fetch(`${baseUrl}/api/v1/conversations`, {
                method: 'POST',
                headers: buildHeaders(),
                credentials,
                body: JSON.stringify({ title }),
            });
            if (!resp.ok)
                throw new Error(`Create conversation failed: ${resp.status}`);
            return resp.json();
        },
        async deleteConversation(id) {
            const resp = await fetch(`${baseUrl}/api/v1/conversations/${encodeURIComponent(id)}`, {
                method: 'DELETE',
                headers: buildHeaders(),
                credentials,
            });
            if (!resp.ok)
                throw new Error(`Delete conversation failed: ${resp.status}`);
        },
    };
}
//# sourceMappingURL=runtime.js.map