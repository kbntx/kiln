import { Sun, Moon } from 'lucide-react';
import { Button } from '@/shared/components/generic/ui/button.component';
import { useTheme } from '@/shared/hooks/use-theme.hook';

export function ThemeToggle() {
  const { theme, toggleTheme } = useTheme();

  return (
    <Button variant="ghost" size="icon" onClick={toggleTheme} aria-label="Toggle theme">
      {theme === 'dark' ? <Sun className="size-4" /> : <Moon className="size-4" />}
    </Button>
  );
}
