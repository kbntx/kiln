import { cn } from '@/shared/helpers/utils';

import type { Repo } from '../services/repos.service';

interface RepoListProps {
  repos: Repo[];
  selectedRepo: Repo | null;
  onSelect: (repo: Repo) => void;
}

export function RepoList({ repos, selectedRepo, onSelect }: RepoListProps) {
  return (
    <nav className="flex flex-col gap-1">
      {repos.map(repo => {
        const isSelected = selectedRepo?.owner === repo.owner && selectedRepo?.name === repo.name;

        return (
          <button
            key={`${repo.owner}/${repo.name}`}
            type="button"
            onClick={() => onSelect(repo)}
            className={cn(
              'rounded-lg px-3 py-2 text-left text-sm font-medium transition-colors',
              isSelected
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground'
            )}
          >
            {repo.owner}/{repo.name}
          </button>
        );
      })}
    </nav>
  );
}
