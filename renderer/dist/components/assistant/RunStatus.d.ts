import React from 'react';
import type { ComponentProps } from '../../registry';
export type RunState = 'idle' | 'running' | 'complete' | 'error' | 'cancelled';
export interface RunStatusProps {
    showElapsed?: boolean;
    showTokenCount?: boolean;
    compact?: boolean;
}
export declare function AssistantRunStatus({ instance, children }: ComponentProps): React.ReactElement;
export declare namespace AssistantRunStatus {
    var displayName: string;
}
//# sourceMappingURL=RunStatus.d.ts.map