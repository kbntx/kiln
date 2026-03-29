import { Badge } from '@/shared/components/generic/ui/badge.component';
import { cn } from '@/shared/helpers/utils';

interface RunStatusBadgeProps {
  status: string;
}

export function RunStatusBadge({ status }: RunStatusBadgeProps) {
  switch (status) {
    case 'pending':
      return <Badge variant="secondary">Pending</Badge>;

    case 'cloning':
      return (
        <Badge variant="outline" className="border-blue-500/30 text-blue-500">
          Cloning
        </Badge>
      );

    case 'discovering':
      return (
        <Badge variant="outline" className="border-blue-500/30 text-blue-500">
          Discovering
        </Badge>
      );

    case 'running':
      return (
        <Badge className="bg-blue-500/10 text-blue-500">
          <span
            className={cn('mr-1 inline-block size-1.5 rounded-full bg-blue-500', 'animate-pulse')}
          />
          Running
        </Badge>
      );

    case 'success':
      return <Badge className="bg-green-500/10 text-green-600 dark:text-green-400">Success</Badge>;

    case 'failed':
      return <Badge variant="destructive">Failed</Badge>;

    default:
      return <Badge variant="secondary">{status || 'Unknown'}</Badge>;
  }
}
