import { ArrowRight, ExternalLink, GitBranch } from 'lucide-react';
import { Link } from 'react-router-dom';

import { Avatar, AvatarFallback, AvatarImage } from '@/shared/components/generic/ui/avatar.component';
import { Badge } from '@/shared/components/generic/ui/badge.component';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/generic/ui/card.component';

import type { PullRequest } from '../services/repos.service';

interface PrCardProps {
  pr: PullRequest;
  owner: string;
  repo: string;
}

function timeAgo(date: string): string {
  const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000);

  if (seconds < 60) return 'just now';

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;

  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;

  const months = Math.floor(days / 30);
  return `${months}mo ago`;
}

export function PrCard({ pr, owner, repo }: PrCardProps) {
  return (
    <Link to={`/run/${owner}/${repo}/${pr.number}`} state={{ prTitle: pr.title, prBranch: pr.branch, headSha: pr.headSha }} className="block">
      <Card className="hover:ring-primary/30 transition-shadow hover:ring-2">
        <CardHeader>
          <CardTitle className="flex items-start justify-between gap-2">
            <span>
              {pr.title} <span className="text-muted-foreground font-normal">#{pr.number}</span>
            </span>
            <div className="flex items-center gap-2">
              <a
                href={`https://github.com/${owner}/${repo}/pull/${pr.number}`}
                target="_blank"
                rel="noopener noreferrer"
                onClick={e => e.stopPropagation()}
                className="text-muted-foreground hover:text-foreground transition-colors"
                title="Open on GitHub"
              >
                <ExternalLink className="size-4" />
              </a>
              <Badge
                variant={pr.approved ? 'default' : 'secondary'}
                className={
                  pr.approved ? 'bg-emerald-600 text-white' : 'bg-amber-500/15 text-amber-600'
                }
              >
                {pr.approved ? 'Approved' : 'Pending review'}
              </Badge>
            </div>
          </CardTitle>
        </CardHeader>

        <CardContent className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Avatar size="sm">
                <AvatarImage src={pr.authorAvatar} alt={pr.author} />
                <AvatarFallback>{pr.author.slice(0, 2).toUpperCase()}</AvatarFallback>
              </Avatar>
              <span className="text-muted-foreground text-sm">{pr.author}</span>
            </div>

            <span className="text-muted-foreground text-xs">{timeAgo(pr.updatedAt)}</span>
          </div>

          <div className="text-muted-foreground flex items-center gap-1.5 text-xs">
            <GitBranch className="size-3.5" />
            <code className="bg-muted rounded px-1 py-0.5">{pr.branch}</code>
            <ArrowRight className="size-3" />
            <code className="bg-muted rounded px-1 py-0.5">{pr.baseBranch}</code>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
