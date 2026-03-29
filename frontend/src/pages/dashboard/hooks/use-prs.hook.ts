import { useCallback, useEffect, useState } from 'react';

import { fetchPRs, type PullRequest } from '../services/repos.service';

export function usePRs(owner: string, repo: string) {
  const [prs, setPRs] = useState<PullRequest[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(() => {
    if (!owner || !repo) return;

    setIsLoading(true);
    setError(null);

    fetchPRs(owner, repo)
      .then(setPRs)
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : 'Failed to load pull requests');
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, [owner, repo]);

  useEffect(() => {
    load();
  }, [load]);

  return { prs, isLoading, error, refresh: load };
}
