import { Loader2 } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';

import { Button } from '@/shared/components/generic/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/shared/components/generic/ui/dialog';
import type { Project } from '@/shared/services/sse.service';

import { LogViewer } from './components/log-viewer.component';
import { ProjectPicker } from './components/project-picker.component';
import { RunHeader } from './components/run-header.component';
import { RunStatusBadge } from './components/run-status-badge.component';
import { useRun } from './hooks/use-run.hook';

export function RunPage() {
  const { owner = '', repo = '', prNumber = '' } = useParams();
  const prNum = Number(prNumber);

  const { phase, projects, logs, status, error, executionRun, startDiscovery, startExecution } =
    useRun(owner, repo, prNum);

  const lastSelectionRef = useRef<{
    projectDir: string;
    stack: string;
  } | null>(null);

  const [applyDialogOpen, setApplyDialogOpen] = useState(false);

  const hasStarted = useRef(false);
  useEffect(() => {
    if (!hasStarted.current && owner && repo && prNum) {
      hasStarted.current = true;
      startDiscovery();
    }
  }, [owner, repo, prNum, startDiscovery]);

  const handleProjectSelect = useCallback(
    (project: Project, stack: string) => {
      lastSelectionRef.current = { projectDir: project.dir, stack };
      startExecution(project.dir, stack, 'plan');
    },
    [startExecution]
  );

  const handleApply = useCallback(() => {
    setApplyDialogOpen(false);
    if (lastSelectionRef.current) {
      startExecution(lastSelectionRef.current.projectDir, lastSelectionRef.current.stack, 'apply');
    }
  }, [startExecution]);

  const isPlanSuccess =
    phase === 'done' && status === 'success' && executionRun?.operation === 'plan';

  return (
    <div className="bg-background flex min-h-screen flex-col">
      <RunHeader owner={owner} repo={repo} prNumber={prNum} />

      <main className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-6 py-8">
        {phase === 'idle' && <p className="text-muted-foreground">Initializing...</p>}

        {phase === 'discovering' && (
          <div className="flex flex-col items-center gap-3 py-12">
            <Loader2 className="text-muted-foreground size-6 animate-spin" />
            <p className="text-muted-foreground text-sm">Discovering infrastructure projects...</p>
          </div>
        )}

        {phase === 'ready' && (
          <>
            <h2 className="text-lg font-medium">Select a project to plan</h2>
            <ProjectPicker projects={projects} onSelect={handleProjectSelect} isLoading={false} />
          </>
        )}

        {(phase === 'running' || phase === 'done') && (
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-medium">
                {executionRun?.operation === 'apply' ? 'Apply' : 'Plan'} Output
              </h2>
              <RunStatusBadge status={status} />
            </div>

            <LogViewer logs={logs} isStreaming={phase === 'running'} />

            {isPlanSuccess && (
              <div className="flex justify-end">
                <Dialog open={applyDialogOpen} onOpenChange={setApplyDialogOpen}>
                  <DialogTrigger render={<Button />}>Apply Changes</DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Confirm Apply</DialogTitle>
                      <DialogDescription>
                        Are you sure? This will apply changes to your infrastructure. This action
                        cannot be undone.
                      </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                      <Button variant="outline" onClick={() => setApplyDialogOpen(false)}>
                        Cancel
                      </Button>
                      <Button onClick={handleApply}>Apply</Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              </div>
            )}

            {phase === 'done' && status === 'failed' && (
              <p className="text-destructive text-sm">
                The {executionRun?.operation ?? 'operation'} failed. Check the logs above for
                details.
              </p>
            )}
          </div>
        )}

        {phase === 'error' && (
          <div className="flex flex-col items-center gap-4 py-12">
            <p className="text-destructive text-sm">{error}</p>
            <Button variant="outline" onClick={startDiscovery}>
              Retry
            </Button>
          </div>
        )}
      </main>
    </div>
  );
}
