import { useCallback, useEffect, useRef, useState } from 'react';

import { subscribeToRun } from '@/shared/services/sse.service';
import type { LogLine, Project } from '@/shared/services/sse.service';

import { createDiscovery, createExecution } from '../services/run.service';
import type { Run } from '../services/run.service';

type Phase = 'idle' | 'discovering' | 'ready' | 'running' | 'done' | 'error';

interface RunState {
  discoveryRun: Run | null;
  executionRun: Run | null;
  logs: LogLine[];
  projects: Project[];
  status: string;
  phase: Phase;
  hasChanges: boolean;
  error: string | null;
}

export function useRun(owner: string, repo: string, prNumber: number, prBranch: string, headSha: string) {
  const [state, setState] = useState<RunState>({
    discoveryRun: null,
    executionRun: null,
    logs: [],
    projects: [],
    status: '',
    phase: 'idle',
    hasChanges: false,
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
      const run = await createDiscovery({ owner, repo, prNumber, prBranch, headSha });

      setState(prev => ({ ...prev, discoveryRun: run }));

      unsubRef.current = subscribeToRun(run.id, {
        onProjects(projects) {
          setState(prev => ({ ...prev, projects, phase: 'ready' }));
        },
        onRunError(message) {
          setState(prev => ({ ...prev, error: message }));
        },
        onStatus(status) {
          setState(prev => {
            if (status === 'failed') {
              return { ...prev, status, phase: 'error', error: prev.error ?? 'Discovery failed' };
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
  }, [owner, repo, prNumber, headSha, cleanup]);

  const startExecution = useCallback(
    async (
      projectDir: string,
      stack: string,
      operation: 'plan' | 'apply',
      profile: string,
      destroy: boolean,
      keepLogs?: boolean,
      planRunId?: string
    ) => {
      cleanup();

      setState(prev => ({
        ...prev,
        phase: 'running',
        logs: keepLogs
          ? [
              ...prev.logs,
              {
                text: destroy ? `▶ ${operation} (destroy)` : `▶ ${operation}`,
                time: new Date().toISOString(),
                stream: 'separator',
                separator: true
              }
            ]
          : [],
        status: 'pending',
        hasChanges: false,
        error: null
      }));

      try {
        const run = await createExecution({
          owner,
          repo,
          prNumber,
          prBranch,
          headSha,
          projectDir,
          stack,
          profile: profile ?? '',
          operation,
          destroy: destroy ?? false,
          planRunId
        });

        setState(prev => ({ ...prev, executionRun: run }));

        unsubRef.current = subscribeToRun(run.id, {
          onLog(line) {
            setState(prev => ({ ...prev, logs: [...prev.logs, line] }));
          },
          onHasChanges(hasChanges) {
            setState(prev => ({ ...prev, hasChanges }));
          },
          onRunError(message) {
            setState(prev => ({ ...prev, error: message }));
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
              phase: prev.logs.length > 0 ? 'done' : 'error',
              status: prev.logs.length > 0 ? 'failed' : prev.status,
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
    [owner, repo, prNumber, headSha, cleanup]
  );

  const backToProjects = useCallback(() => {
    cleanup();
    setState(prev => ({
      ...prev,
      phase: 'ready',
      logs: [],
      status: '',
      executionRun: null,
      error: null
    }));
  }, [cleanup]);

  return {
    ...state,
    startDiscovery,
    startExecution,
    backToProjects
  };
}
