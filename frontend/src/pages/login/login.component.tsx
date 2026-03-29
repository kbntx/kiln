import { Flame } from 'lucide-react';

import { Card, CardContent, CardHeader } from '@/shared/components/generic/ui/card.component';
import { ThemeToggle } from '@/shared/components/theme-toggle.component';

import { GitHubButton } from './components/github-button.component';

export function LoginPage() {
  return (
    <div className="bg-background relative flex min-h-screen items-center justify-center px-4">
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>

      <Card className="w-full max-w-md">
        <CardHeader className="items-center text-center">
          <div className="flex items-center gap-2">
            <Flame className="text-primary size-7" />
            <h1 className="text-2xl font-bold tracking-tight">Kiln</h1>
          </div>
          <p className="text-muted-foreground mt-1 text-sm font-medium">
            Infrastructure as Code, Simplified
          </p>
        </CardHeader>

        <CardContent className="flex flex-col gap-6">
          <p className="text-muted-foreground text-center text-sm">
            Plan and apply your Terraform changes directly from pull requests.
          </p>

          <GitHubButton />
        </CardContent>
      </Card>
    </div>
  );
}
