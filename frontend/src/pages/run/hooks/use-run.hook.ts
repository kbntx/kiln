import { useCallback, useEffect, useRef, useState } from 'react';

import { subscribeToRun } from '@/shared/services/sse.service';
import type { LogLine, Project } from '@/shared/services/sse.service';

import { createRun } from '../services/run.service';
import type { Run } from '../services/run.service';

type Phase = 'idle' | 'discovering' | 'ready' | 'running' | 'done' | 'error';

interface RunState {
  discoveryRun: Run | null;
  executionRun: Run | null;
  logs: LogLine[];
  projects: Project[];
  status: string;
  phase: Phase;
  error: string | null;
}

export function useRun(owner: string, repo: string, prNumber: number) {
  const [state, setState] = useState<RunState>({
    discoveryRun: null,
    executionRun: null,
    logs: [],
    projects: [],
    status: '',
    phase: 'idle',
    error: null
  });

  const unsubRef = useRef<(() => void) | null>(null);

  const cleanup = useCallback(() => {
    if (unsubRef.current) {
      unsubRef.current();
      unsubRef.current = null;
    }
  }, []);

  useEffect(() => {
    return cleanup;
  }, [cleanup]);

  const startDiscovery = useCallback(async () => {
    cleanup();

    setState(prev => ({
      ...prev,
      phase: 'discovering',
      projects: [],
      error: null
    }));

    try {
      const run = await createRun({ owner, repo, prNumber, prBranch: '' });

      setState(prev => ({ ...prev, discoveryRun: run }));

      unsubRef.current = subscribeToRun(run.id, {
        onProjects(projects) {
          setState(prev => ({ ...prev, projects, phase: 'ready' }));
        },
        onStatus(status) {
          setState(prev => {
            if (status === 'failed') {
              return { ...prev, status, phase: 'error', error: 'Discovery failed' };
            }
            return { ...prev, status };
          });
        },
        onError() {
          setState(prev => ({
            ...prev,
            phase: 'error',
            error: 'Connection lost during discovery'
          }));
        }
      });
    } catch (err) {
      setState(prev => ({
        ...prev,
        phase: 'error',
        error: err instanceof Error ? err.message : 'Failed to start discovery'
      }));
    }
  }, [owner, repo, prNumber, cleanup]);

  const startExecution = useCallback(
    async (projectDir: string, stack: string, operation: 'plan' | 'apply') => {
      cleanup();

      setState(prev => ({
        ...prev,
        phase: 'running',
        logs: [],
        status: 'pending',
        error: null
      }));

      try {
        const run = await createRun({
          owner,
          repo,
          prNumber,
          prBranch: '',
          projectDir,
          stack,
          operation
        });

        setState(prev => ({ ...prev, executionRun: run }));

        unsubRef.current = subscribeToRun(run.id, {
          onLog(line) {
            setState(prev => ({ ...prev, logs: [...prev.logs, line] }));
          },
          onStatus(status) {
            setState(prev => {
              if (status === 'success' || status === 'failed') {
                return { ...prev, status, phase: 'done' };
              }
              return { ...prev, status };
            });
          },
          onError() {
            setState(prev => ({
              ...prev,
              phase: 'error',
              error: 'Connection lost during execution'
            }));
          }
        });
      } catch (err) {
        setState(prev => ({
          ...prev,
          phase: 'error',
          error: err instanceof Error ? err.message : 'Failed to start run'
        }));
      }
    },
    [owner, repo, prNumber, cleanup]
  );

  return {
    ...state,
    startDiscovery,
    startExecution
  };
}
