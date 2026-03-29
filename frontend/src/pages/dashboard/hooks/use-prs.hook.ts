import { useEffect, useState } from 'react';

import { fetchPRs, type PullRequest } from '../services/repos.service';

export function usePRs(owner: string, repo: string) {
  const [prs, setPRs] = useState<PullRequest[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!owner || !repo) return;

    let cancelled = false;
    setIsLoading(true);
    setError(null);

    fetchPRs(owner, repo)
      .then(data => {
        if (!cancelled) {
          setPRs(data);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load pull requests');
        }
      })
      .finally(() => {
        if (!cancelled) {
          setIsLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [owner, repo]);

  return { prs, isLoading, error };
}
