import { get, post } from '@/shared/services/api.service';

export type { LogLine, Project } from '@/shared/services/sse.service';

export interface Run {
  id: string;
  owner: string;
  repo: string;
  prNumber: number;
  prBranch: string;
  projectDir: string;
  stack: string;
  operation: string;
  destroy: boolean;
  status: 'pending' | 'cloning' | 'discovering' | 'running' | 'success' | 'failed';
  projects: import('@/shared/services/sse.service').Project[];
  createdAt: string;
}

export interface DiscoverRequest {
  owner: string;
  repo: string;
  prNumber: number;
  prBranch: string;
  headSha: string;
}

export interface ExecuteRequest {
  owner: string;
  repo: string;
  prNumber: number;
  prBranch: string;
  headSha: string;
  projectDir: string;
  stack: string;
  profile: string;
  operation: 'plan' | 'apply';
  destroy: boolean;
  planRunId?: string;
}

export function createDiscovery(req: DiscoverRequest): Promise<Run> {
  return post<Run>('/api/runs', req);
}

export function createExecution(req: ExecuteRequest): Promise<Run> {
  return post<Run>('/api/runs', req);
}

export function getRun(id: string): Promise<Run> {
  return get<Run>(`/api/runs/${id}`);
}
