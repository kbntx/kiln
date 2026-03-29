import { FileQuestion, Loader2 } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useLocation, useParams } from 'react-router-dom';

import { Button } from '@/shared/components/generic/ui/button.component';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/shared/components/generic/ui/dialog.component';
import type { Project } from '@/shared/services/sse.service';

import { LogViewer } from './components/log-viewer.component';
import { ProjectPicker } from './components/project-picker.component';
import { RunHeader } from './components/run-header.component';
import { RunStatusBadge } from './components/run-status-badge.component';
import { useRun } from './hooks/use-run.hook';

export function RunPage() {
  const { owner = '', repo = '', prNumber = '' } = useParams();
  const prNum = Number(prNumber);
  const location = useLocation();
  const routeState = location.state as { prTitle?: string; prBranch?: string; headSha?: string } | null;
  const prTitle = routeState?.prTitle;
  const prBranch = routeState?.prBranch ?? '';
  const headSha = routeState?.headSha ?? '';

  const {
    phase,
    projects,
    logs,
    status,
    hasChanges,
    error,
    executionRun,
    startDiscovery,
    startExecution,
    backToProjects
  } = useRun(owner, repo, prNum, prBranch, headSha);

  const lastSelectionRef = useRef<{
    projectDir: string;
    stack: string;
    profile: string;
    destroy: boolean;
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
    (project: Project, stack: string, destroy?: boolean) => {
      lastSelectionRef.current = {
        projectDir: project.dir,
        stack,
        profile: project.profile,
        destroy: destroy ?? false
      };
      startExecution(project.dir, stack, 'plan', project.profile, destroy ?? false);
    },
    [startExecution]
  );

  const handleApply = useCallback(() => {
    setApplyDialogOpen(false);
    if (lastSelectionRef.current) {
      startExecution(
        lastSelectionRef.current.projectDir,
        lastSelectionRef.current.stack,
        'apply',
        lastSelectionRef.current.profile,
        lastSelectionRef.current.destroy,
        true, // keepLogs — preserve plan output and add separator
        executionRun?.id // planRunId — reuse the plan run's workspace
      );
    }
  }, [startExecution, executionRun]);

  const isPlanSuccess =
    phase === 'done' && status === 'success' && executionRun?.operation === 'plan';
  const showApply = isPlanSuccess && hasChanges;

  const isDestroy = executionRun?.destroy ?? lastSelectionRef.current?.destroy;
  const operationLabel = isDestroy
    ? executionRun?.operation === 'apply'
      ? 'Apply Destroy'
      : 'Plan Destroy'
    : executionRun?.operation === 'apply'
      ? 'Apply'
      : 'Plan';

  return (
    <div className="bg-background flex min-h-screen flex-col">
      <RunHeader
        owner={owner}
        repo={repo}
        prNumber={prNum}
        prTitle={prTitle}
        showBackToProjects={phase === 'running' || phase === 'done'}
        onBackToProjects={backToProjects}
      />

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
              <h2 className="text-lg font-medium">{operationLabel} Output</h2>
              <RunStatusBadge status={status} />
            </div>

            <LogViewer logs={logs} isStreaming={phase === 'running'} />

            {showApply && (
              <div className="flex justify-end">
                <Dialog open={applyDialogOpen} onOpenChange={setApplyDialogOpen}>
                  <DialogTrigger
                    render={<Button variant={isDestroy ? 'destructive' : 'default'} />}
                  >
                    {isDestroy ? 'Apply Destroy' : 'Apply Changes'}
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>
                        {isDestroy ? 'Confirm Destroy' : 'Confirm Apply'}
                      </DialogTitle>
                      <DialogDescription>
                        {isDestroy
                          ? 'Are you sure? This will DESTROY infrastructure resources. This action cannot be undone.'
                          : 'Are you sure? This will apply changes to your infrastructure. This action cannot be undone.'}
                      </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                      <Button variant="outline" onClick={() => setApplyDialogOpen(false)}>
                        Cancel
                      </Button>
                      <Button
                        variant={isDestroy ? 'destructive' : 'default'}
                        onClick={handleApply}
                      >
                        {isDestroy ? 'Destroy' : 'Apply'}
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              </div>
            )}

            {isPlanSuccess && !hasChanges && (
              <p className="text-muted-foreground text-sm">No infrastructure changes detected.</p>
            )}

            {phase === 'done' && status === 'failed' && (
              <p className="text-destructive text-sm">
                The {operationLabel.toLowerCase()} failed. Check the logs above for details.
              </p>
            )}
          </div>
        )}

        {phase === 'error' && (
          <div className="flex flex-col items-center gap-4 py-16">
            {error?.includes('kiln.yaml') ? (
              <>
                <FileQuestion className="text-muted-foreground size-12 opacity-50" />
                <div className="text-center">
                  <h3 className="text-lg font-medium">No kiln.yaml found</h3>
                  <p className="text-muted-foreground mt-1 max-w-md text-sm">
                    This branch doesn't have a <code className="bg-muted rounded px-1.5 py-0.5">kiln.yaml</code> file
                    at the repository root. Add one to define your Terraform projects.
                  </p>
                </div>
                <pre className="bg-muted mt-2 max-w-lg overflow-x-auto rounded-lg p-4 text-left text-xs">
{`profiles:
  default:
    env:
      AWS_PROFILE: my-profile

projects:
  - name: my-infra
    dir: "."
    engine: terraform
    stacks: [default]
    profile: default`}
                </pre>
                <Button variant="outline" onClick={startDiscovery} className="mt-2">
                  Retry
                </Button>
              </>
            ) : (
              <>
                <p className="text-destructive text-sm">{error}</p>
                <Button variant="outline" onClick={startDiscovery}>
                  Retry
                </Button>
              </>
            )}
          </div>
        )}
      </main>
    </div>
  );
}
