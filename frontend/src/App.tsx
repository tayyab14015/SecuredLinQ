import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';

// Admin Pages
import AdminLogin from './pages/admin/Login';
import AdminDashboard from './pages/admin/Dashboard';
import LoadDetail from './pages/admin/LoadDetail';
import Gallery from './pages/admin/Gallery';

// Driver Pages
import DriverLogin from './pages/driver/Login';
import DriverRegister from './pages/driver/Register';
import DriverDashboard from './pages/driver/Dashboard';

// Shared Pages
import MeetingRoom from './pages/MeetingRoom';

// Protected Route Component for Admin
interface ProtectedRouteProps {
  children: ReactNode;
  requiredRole?: 'admin' | 'driver';
}

function ProtectedRoute({ children, requiredRole }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, userType, isAdmin, isDriver } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--bg-primary)]">
        <div className="spinner"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    // Redirect based on required role
    if (requiredRole === 'driver') {
      return <Navigate to="/driver/login" replace />;
    }
    return <Navigate to="/admin/login" replace />;
  }

  // Check role if required
  if (requiredRole) {
    if (requiredRole === 'admin' && !isAdmin) {
      return <Navigate to="/driver/dashboard" replace />;
    }
    if (requiredRole === 'driver' && !isDriver) {
      return <Navigate to="/admin/dashboard" replace />;
    }
  }

  return <>{children}</>;
}

// Redirect authenticated users away from auth pages
function AuthRoute({ children, forRole }: { children: ReactNode; forRole: 'admin' | 'driver' }) {
  const { isAuthenticated, isLoading, isAdmin, isDriver } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--bg-primary)]">
        <div className="spinner"></div>
      </div>
    );
  }

  if (isAuthenticated) {
    // Redirect to appropriate dashboard
    if (isAdmin) {
      return <Navigate to="/admin/dashboard" replace />;
    }
    if (isDriver) {
      return <Navigate to="/driver/dashboard" replace />;
    }
  }

  return <>{children}</>;
}

function AppRoutes() {
  const { isAuthenticated, isAdmin, isDriver, isLoading } = useAuth();

  // Root redirect logic
  const getRootRedirect = () => {
    if (isLoading) return null;
    if (isAuthenticated) {
      if (isAdmin) return '/admin/dashboard';
      if (isDriver) return '/driver/dashboard';
    }
    return '/driver/login';
  };

  return (
    <Routes>
      {/* Root - redirect to appropriate page */}
      <Route 
        path="/" 
        element={
          isLoading ? (
            <div className="min-h-screen flex items-center justify-center bg-[var(--bg-primary)]">
              <div className="spinner"></div>
            </div>
          ) : (
            <Navigate to={getRootRedirect() || '/driver/login'} replace />
          )
        } 
      />

      {/* Driver Auth Routes */}
      <Route
        path="/driver/login"
        element={
          <AuthRoute forRole="driver">
            <DriverLogin />
          </AuthRoute>
        }
      />
      <Route
        path="/driver/register"
        element={
          <AuthRoute forRole="driver">
            <DriverRegister />
          </AuthRoute>
        }
      />

      {/* Driver Protected Routes */}
      <Route
        path="/driver/dashboard"
        element={
          <ProtectedRoute requiredRole="driver">
            <DriverDashboard />
          </ProtectedRoute>
        }
      />

      {/* Admin Auth Routes */}
      <Route
        path="/admin/login"
        element={
          <AuthRoute forRole="admin">
            <AdminLogin />
          </AuthRoute>
        }
      />

      {/* Admin Protected Routes */}
      <Route
        path="/admin/dashboard"
        element={
          <ProtectedRoute requiredRole="admin">
            <AdminDashboard />
          </ProtectedRoute>
        }
      />
      <Route
        path="/admin/load/:guestRand"
        element={
          <ProtectedRoute requiredRole="admin">
            <LoadDetail />
          </ProtectedRoute>
        }
      />
      <Route
        path="/admin/gallery/:loadId"
        element={
          <ProtectedRoute requiredRole="admin">
            <Gallery />
          </ProtectedRoute>
        }
      />

      {/* Meeting Routes - require authentication */}
      <Route
        path="/join/:roomId"
        element={
          <ProtectedRoute>
            <MeetingRoom />
          </ProtectedRoute>
        }
      />
      <Route
        path="/admin/meeting/:roomId"
        element={
          <ProtectedRoute requiredRole="admin">
            <MeetingRoom isAdmin />
          </ProtectedRoute>
        }
      />

      {/* Fallback - redirect to driver login */}
      <Route path="*" element={<Navigate to="/driver/login" replace />} />
    </Routes>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRoutes />
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
