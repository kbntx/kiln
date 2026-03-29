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
  status: 'pending' | 'cloning' | 'discovering' | 'running' | 'success' | 'failed';
  projects: import('@/shared/services/sse.service').Project[];
  createdAt: string;
}

export interface CreateRunRequest {
  owner: string;
  repo: string;
  prNumber: number;
  prBranch: string;
  projectDir?: string;
  stack?: string;
  operation?: string;
}

export function createRun(req: CreateRunRequest): Promise<Run> {
  return post<Run>('/api/runs', req);
}

export function getRun(id: string): Promise<Run> {
  return get<Run>(`/api/runs/${id}`);
}
