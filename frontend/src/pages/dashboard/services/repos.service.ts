import { get } from '@/shared/services/api.service';

export interface Repo {
  owner: string;
  name: string;
}

export interface PullRequest {
  number: number;
  title: string;
  author: string;
  authorAvatar: string;
  branch: string;
  baseBranch: string;
  approved: boolean;
  createdAt: string;
  updatedAt: string;
}

export function fetchRepos(): Promise<Repo[]> {
  return get<Repo[]>('/api/repos');
}

export function fetchPRs(owner: string, repo: string): Promise<PullRequest[]> {
  return get<PullRequest[]>(`/api/repos/${owner}/${repo}/prs`);
}
