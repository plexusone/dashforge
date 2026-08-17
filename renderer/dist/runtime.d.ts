export interface Message {
    id: string;
    role: 'user' | 'assistant' | 'system';
    content: string;
    toolCalls?: ToolCallRecord[];
    createdAt?: string;
}
export interface ToolCallRecord {
    id: string;
    name: string;
    args: string;
    result?: string;
    status: 'running' | 'complete' | 'error';
}
export interface Conversation {
    id: string;
    title: string;
    updatedAt: string;
}
export interface RuntimeConfig {
    baseUrl: string;
    headers?: Record<string, string>;
    credentials?: RequestCredentials;
    conversationId?: string | null;
    onMessage?: (message: Message) => void;
    onError?: (error: Error) => void;
}
export interface ExternalStoreRuntime {
    config: RuntimeConfig;
    sendMessage(content: string, options?: {
        model?: string;
    }): AsyncGenerator<Message>;
    listConversations(): Promise<Conversation[]>;
    loadConversation(id: string): Promise<Message[]>;
    createConversation(title: string): Promise<Conversation>;
    deleteConversation(id: string): Promise<void>;
}
export declare function createAgentOSRuntime(config: RuntimeConfig): ExternalStoreRuntime;
//# sourceMappingURL=runtime.d.ts.map