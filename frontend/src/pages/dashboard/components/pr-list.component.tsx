import { GitPullRequest } from 'lucide-react';

import { Skeleton } from '@/shared/components/generic/ui/skeleton';

import type { PullRequest } from '../services/repos.service';
import { PrCard } from './pr-card.component';

interface PrListProps {
  prs: PullRequest[];
  owner: string;
  repo: string;
  isLoading: boolean;
}

function PrSkeleton() {
  return (
    <div className="ring-foreground/10 flex flex-col gap-4 rounded-xl p-4 ring-1">
      <div className="flex items-start justify-between">
        <Skeleton className="h-5 w-3/5" />
        <Skeleton className="h-5 w-20 rounded-full" />
      </div>
      <div className="flex items-center gap-2">
        <Skeleton className="size-6 rounded-full" />
        <Skeleton className="h-4 w-24" />
      </div>
      <Skeleton className="h-4 w-2/5" />
    </div>
  );
}

export function PrList({ prs, owner, repo, isLoading }: PrListProps) {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-4">
        <PrSkeleton />
        <PrSkeleton />
        <PrSkeleton />
      </div>
    );
  }

  if (prs.length === 0) {
    return (
      <div className="text-muted-foreground flex flex-col items-center justify-center gap-3 rounded-xl border border-dashed py-16">
        <GitPullRequest className="size-10 opacity-40" />
        <p className="text-sm font-medium">No open pull requests</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {prs.map(pr => (
        <PrCard key={pr.number} pr={pr} owner={owner} repo={repo} />
      ))}
    </div>
  );
}
