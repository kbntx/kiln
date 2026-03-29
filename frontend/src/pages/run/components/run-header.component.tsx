import { ChevronRight, GitBranch } from 'lucide-react';
import { Link } from 'react-router-dom';

import { Button } from '@/shared/components/generic/ui/button.component';

interface RunHeaderProps {
  owner: string;
  repo: string;
  prNumber: number;
  prTitle?: string;
  showBackToProjects?: boolean;
  onBackToProjects?: () => void;
}

export function RunHeader({
  owner,
  repo,
  prNumber,
  prTitle,
  showBackToProjects,
  onBackToProjects
}: RunHeaderProps) {
  return (
    <header className="flex items-center gap-2 border-b px-6 py-4">
      <Button
        variant="link"
        size="sm"
        className="text-muted-foreground p-0"
        render={<Link to="/" />}
      >
        Dashboard
      </Button>

      <ChevronRight className="text-muted-foreground size-4" />

      <span className="font-medium">
        {owner}/{repo}
      </span>

      <ChevronRight className="text-muted-foreground size-4" />

      {showBackToProjects && onBackToProjects ? (
        <>
          <Button
            variant="link"
            size="sm"
            className="text-muted-foreground truncate p-0"
            onClick={onBackToProjects}
            title={prTitle}
          >
            {prTitle ? (
              <span className="max-w-[200px] truncate">{prTitle}</span>
            ) : null}
            <span>#{prNumber}</span>
          </Button>

          <ChevronRight className="text-muted-foreground size-4" />

          <span className="text-sm">Logs</span>
        </>
      ) : (
        <div className="flex min-w-0 items-center gap-2">
          {prTitle && (
            <span className="truncate text-sm" title={prTitle}>
              {prTitle}
            </span>
          )}
          <span className="text-muted-foreground shrink-0">#{prNumber}</span>
        </div>
      )}

      <div className="text-muted-foreground ml-2 flex items-center gap-1 text-sm">
        <GitBranch className="size-3.5" />
      </div>
    </header>
  );
}
