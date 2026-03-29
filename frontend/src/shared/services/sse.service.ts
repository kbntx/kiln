export interface LogLine {
  text: string;
  timestamp: string;
  stream: string;
}

export interface Project {
  name: string;
  dir: string;
  engine: 'pulumi' | 'terraform';
  stacks: string[];
}

interface RunHandlers {
  onLog?: (line: LogLine) => void;
  onStatus?: (status: string) => void;
  onProjects?: (projects: Project[]) => void;
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

  if (handlers.onError) {
    source.onerror = handlers.onError;
  }

  return () => source.close();
}
