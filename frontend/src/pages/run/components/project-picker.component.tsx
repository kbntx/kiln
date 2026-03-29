import { useState } from 'react';

import { Badge } from '@/shared/components/generic/ui/badge.component';
import { Button } from '@/shared/components/generic/ui/button.component';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/generic/ui/card.component';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/shared/components/generic/ui/select.component';
import { Skeleton } from '@/shared/components/generic/ui/skeleton.component';
import type { Project } from '@/shared/services/sse.service';

interface ProjectPickerProps {
  projects: Project[];
  onSelect: (project: Project, stack: string, destroy?: boolean) => void;
  isLoading: boolean;
}

function ProjectCard({
  project,
  onSelect
}: {
  project: Project;
  onSelect: (project: Project, stack: string, destroy?: boolean) => void;
}) {
  const [selectedStack, setSelectedStack] = useState(project.stacks[0] ?? '');

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {project.name}
          {/* TODO(pulumi): Add Pulumi label back when Pulumi support is implemented. */}
          <Badge variant="outline">Terraform</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <p className="text-muted-foreground text-sm">{project.dir}</p>

        {project.stacks.length > 1 && (
          <Select value={selectedStack} onValueChange={v => setSelectedStack(v ?? '')}>
            <SelectTrigger>
              <SelectValue placeholder="Select stack" />
            </SelectTrigger>
            <SelectContent>
              {project.stacks.map(stack => (
                <SelectItem key={stack} value={stack}>
                  {stack}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        <div className="flex gap-2">
          <Button className="flex-1" onClick={() => onSelect(project, selectedStack)} disabled={!selectedStack}>
            Plan
          </Button>
          <Button
            variant="destructive"
            className="flex-1"
            onClick={() => onSelect(project, selectedStack, true)}
            disabled={!selectedStack}
          >
            Plan Destroy
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export function ProjectPicker({ projects, onSelect, isLoading }: ProjectPickerProps) {
  if (isLoading) {
    return (
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <Skeleton className="h-5 w-40" />
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              <Skeleton className="h-4 w-56" />
              <Skeleton className="h-8 w-full" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (projects.length === 0) {
    return (
      <p className="text-muted-foreground py-8 text-center">
        No infrastructure projects discovered.
      </p>
    );
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {projects.map(project => (
        <ProjectCard key={`${project.dir}-${project.name}`} project={project} onSelect={onSelect} />
      ))}
    </div>
  );
}
