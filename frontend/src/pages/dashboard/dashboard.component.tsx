import { RefreshCw } from 'lucide-react';
import { useState } from 'react';

import { Button } from '@/shared/components/generic/ui/button.component';
import { Skeleton } from '@/shared/components/generic/ui/skeleton.component';

import { RepoList } from './components/repo-list.component';
import { PrList } from './components/pr-list.component';
import { usePRs } from './hooks/use-prs.hook';
import { useRepos } from './hooks/use-repos.hook';
import type { Repo } from './services/repos.service';

export function DashboardPage() {
  const { repos, isLoading: reposLoading, error: reposError } = useRepos();
  const [selectedRepo, setSelectedRepo] = useState<Repo | null>(null);

  const activeRepo = selectedRepo ?? repos[0] ?? null;

  const {
    prs,
    isLoading: prsLoading,
    refresh: refreshPRs
  } = usePRs(activeRepo?.owner ?? '', activeRepo?.name ?? '');

  return (
    <>
      <h2 className="mb-6 text-2xl font-semibold tracking-tight">Dashboard</h2>

      {reposError && (
        <div className="border-destructive/50 bg-destructive/10 text-destructive mb-6 rounded-lg border px-4 py-3 text-sm">
          {reposError}
        </div>
      )}

      <div className="grid grid-cols-1 gap-8 md:grid-cols-[240px_1fr] lg:grid-cols-[280px_1fr]">
        <aside>
          <h3 className="text-muted-foreground mb-3 text-xs font-semibold tracking-wider uppercase">
            Repositories
          </h3>

          {reposLoading ? (
            <div className="flex flex-col gap-2">
              <Skeleton className="h-9 w-full rounded-lg" />
              <Skeleton className="h-9 w-full rounded-lg" />
              <Skeleton className="h-9 w-4/5 rounded-lg" />
            </div>
          ) : (
            <RepoList repos={repos} selectedRepo={activeRepo} onSelect={setSelectedRepo} />
          )}
        </aside>

        <section>
          {activeRepo && (
            <>
              <div className="mb-4 flex items-center justify-between">
                <h3 className="text-muted-foreground text-sm font-medium">
                  Pull requests for{' '}
                  <span className="text-foreground">
                    {activeRepo.owner}/{activeRepo.name}
                  </span>
                </h3>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={refreshPRs}
                  disabled={prsLoading}
                  title="Refresh"
                >
                  <RefreshCw className={`size-4 ${prsLoading ? 'animate-spin' : ''}`} />
                </Button>
              </div>
              <PrList
                prs={prs}
                owner={activeRepo.owner}
                repo={activeRepo.name}
                isLoading={prsLoading}
              />
            </>
          )}

          {!activeRepo && !reposLoading && (
            <div className="text-muted-foreground flex items-center justify-center rounded-xl border border-dashed py-16 text-sm">
              No repositories available
            </div>
          )}
        </section>
      </div>
    </>
  );
}
