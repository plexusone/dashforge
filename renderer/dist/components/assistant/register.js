import { registerComponent } from '../../registry';
import { AssistantThread } from './Thread';
import { AssistantComposer } from './Composer';
import { AssistantThreadList } from './ThreadList';
import { AssistantToolCall } from './ToolCall';
import { AssistantRunStatus } from './RunStatus';
export function registerAssistantComponents() {
    registerComponent('assistant.thread', AssistantThread);
    registerComponent('assistant.composer', AssistantComposer);
    registerComponent('assistant.thread-list', AssistantThreadList);
    registerComponent('assistant.tool-call', AssistantToolCall);
    registerComponent('assistant.run-status', AssistantRunStatus);
    return {
        components: [
            'assistant.thread',
            'assistant.composer',
            'assistant.thread-list',
            'assistant.tool-call',
            'assistant.run-status',
        ],
    };
}
//# sourceMappingURL=register.js.map