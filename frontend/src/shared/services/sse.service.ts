export interface LogLine {
  text: string;
  time: string;
  stream: string;
  separator?: boolean;
}

export interface Project {
  name: string;
  dir: string;
  // TODO(pulumi): Add 'pulumi' back when Pulumi support is implemented.
  engine: 'terraform';
  stacks: string[];
  profile: string;
}

interface RunHandlers {
  onLog?: (line: LogLine) => void;
  onStatus?: (status: string) => void;
  onProjects?: (projects: Project[]) => void;
  onHasChanges?: (hasChanges: boolean) => void;
  onRunError?: (message: string) => void;
  onError?: (error: Event) => void;
}

export function subscribeToRun(runId: string, handlers: RunHandlers): () => void {
  const source = new EventSource(`/api/runs/${runId}/stream`);

  if (handlers.onLog) {
    source.addEventListener('log', e => {
      handlers.onLog!(JSON.parse(e.data) as LogLine);
    });
  }

  if (handlers.onStatus) {
    source.addEventListener('status', e => {
      handlers.onStatus!(e.data);
    });
  }

  if (handlers.onProjects) {
    source.addEventListener('projects', e => {
      handlers.onProjects!(JSON.parse(e.data) as Project[]);
    });
  }

  if (handlers.onHasChanges) {
    source.addEventListener('has_changes', e => {
      handlers.onHasChanges!((e as MessageEvent).data === 'true');
    });
  }

  if (handlers.onRunError) {
    source.addEventListener('run_error', e => {
      handlers.onRunError!((e as MessageEvent).data);
    });
  }

  source.addEventListener('done', () => {
    source.close();
  });

  if (handlers.onError) {
    source.onerror = handlers.onError;
  }

  return () => source.close();
}
