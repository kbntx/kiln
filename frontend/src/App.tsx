import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from '@/shared/hooks/use-auth.hook';
import { ThemeProvider } from '@/shared/hooks/use-theme.hook';
import { AuthGuard } from '@/shared/components/auth-guard.component';
import { AppShell } from '@/shared/components/app-shell.component';
import { LoginPage } from '@/pages/login/login.component';
import { DashboardPage } from '@/pages/dashboard/dashboard.component';
import { RunPage } from '@/pages/run/run.component';

export default function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route element={<AuthGuard />}>
              <Route element={<AppShell />}>
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/run/:owner/:repo/:prNumber" element={<RunPage />} />
              </Route>
            </Route>
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </ThemeProvider>
  );
}
