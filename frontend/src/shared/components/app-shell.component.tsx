import { Outlet, useNavigate } from 'react-router-dom';
import { Flame, LogOut } from 'lucide-react';
import { Avatar, AvatarImage, AvatarFallback } from '@/shared/components/generic/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/shared/components/generic/ui/dropdown-menu';
import { useAuth } from '@/shared/hooks/use-auth.hook';
import { ThemeToggle } from '@/shared/components/theme-toggle.component';

export function AppShell() {
  const { user } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="bg-background text-foreground min-h-screen">
      <header className="border-border bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50 border-b backdrop-blur">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
          <div className="flex items-center gap-2 text-lg font-semibold">
            <Flame className="size-5 text-orange-500" />
            <span>Kiln</span>
          </div>

          <div className="flex items-center gap-2">
            <ThemeToggle />

            {user && (
              <DropdownMenu>
                <DropdownMenuTrigger className="focus-visible:ring-ring cursor-pointer rounded-full outline-none focus-visible:ring-2">
                  <Avatar size="sm">
                    <AvatarImage src={user.avatar} alt={user.login} />
                    <AvatarFallback>{user.login.charAt(0).toUpperCase()}</AvatarFallback>
                  </Avatar>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" sideOffset={8}>
                  <DropdownMenuItem
                    className="cursor-pointer"
                    onClick={() => navigate('/auth/logout')}
                  >
                    <LogOut className="mr-2 size-4" />
                    Sign out
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-7xl px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
