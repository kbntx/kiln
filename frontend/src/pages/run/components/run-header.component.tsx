import { ChevronRight, GitBranch } from 'lucide-react';
import { Link } from 'react-router-dom';

import { Button } from '@/shared/components/generic/ui/button';

interface RunHeaderProps {
  owner: string;
  repo: string;
  prNumber: number;
}

export function RunHeader({ owner, repo, prNumber }: RunHeaderProps) {
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

      <div className="flex items-center gap-2">
        <span className="font-medium">
          {owner}/{repo}
        </span>
        <span className="text-muted-foreground">#{prNumber}</span>
      </div>

      <div className="text-muted-foreground ml-2 flex items-center gap-1 text-sm">
        <GitBranch className="size-3.5" />
      </div>
    </header>
  );
}
