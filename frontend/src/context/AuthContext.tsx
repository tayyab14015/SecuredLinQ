import { createContext, useContext, useEffect, useState, useCallback } from 'react';
import type { ReactNode } from 'react';
import { authApi } from '../api/client';
import type { LoginCredentials, DriverRegisterData, Driver } from '../api/client';

type UserType = 'admin' | 'driver' | null;

interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  userType: UserType;
  username: string | null;
  userId: number | null;
  driver: Driver | null;
}

interface AuthContextType extends AuthState {
  login: (credentials: LoginCredentials) => Promise<{ success: boolean; message: string }>;
  driverLogin: (credentials: LoginCredentials) => Promise<{ success: boolean; message: string }>;
  driverRegister: (data: DriverRegisterData) => Promise<{ success: boolean; message: string }>;
  logout: () => Promise<void>;
  checkSession: () => Promise<void>;
  isAdmin: boolean;
  isDriver: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [state, setState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    userType: null,
    username: null,
    userId: null,
    driver: null,
  });

  const checkSession = useCallback(async () => {
    try {
      const result = await authApi.validateSession();
      if (result.valid) {
        setState({
          isAuthenticated: true,
          isLoading: false,
          userType: result.user_type as UserType,
          username: result.admin?.username || result.driver?.username || null,
          userId: result.user_id || null,
          driver: result.driver || null,
        });
      } else {
        setState({
          isAuthenticated: false,
          isLoading: false,
          userType: null,
          username: null,
          userId: null,
          driver: null,
        });
      }
    } catch {
      setState({
        isAuthenticated: false,
        isLoading: false,
        userType: null,
        username: null,
        userId: null,
        driver: null,
      });
    }
  }, []);

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  // Admin login
  const login = useCallback(async (credentials: LoginCredentials) => {
    try {
      const result = await authApi.login(credentials);
      if (result.success) {
        setState({
          isAuthenticated: true,
          isLoading: false,
          userType: 'admin',
          username: credentials.username,
          userId: 0,
          driver: null,
        });
      }
      return result;
    } catch (error: any) {
      // Extract error message from API response
      let message = 'Login failed. Please try again.';
      if (error?.response?.data?.error) {
        message = error.response.data.error;
      } else if (error?.response?.data?.message) {
        message = error.response.data.message;
      } else if (error instanceof Error && error.message) {
        // Only use error.message if it's not a generic axios error
        if (!error.message.includes('status code') && !error.message.includes('Network Error')) {
          message = error.message;
        }
      }
      return { success: false, message };
    }
  }, []);

  // Driver login
  const driverLogin = useCallback(async (credentials: LoginCredentials) => {
    try {
      const result = await authApi.driverLogin(credentials);
      if (result.success && result.driver) {
        setState({
          isAuthenticated: true,
          isLoading: false,
          userType: 'driver',
          username: result.driver.username,
          userId: result.driver.id,
          driver: result.driver,
        });
      }
      return { success: result.success, message: result.message };
    } catch (error: any) {
      // Extract error message from API response
      let message = 'Login failed. Please try again.';
      if (error?.response?.data?.error) {
        message = error.response.data.error;
      } else if (error?.response?.data?.message) {
        message = error.response.data.message;
      } else if (error instanceof Error && error.message) {
        // Only use error.message if it's not a generic axios error
        if (!error.message.includes('status code') && !error.message.includes('Network Error')) {
          message = error.message;
        }
      }
      return { success: false, message };
    }
  }, []);

  // Driver registration
  const driverRegister = useCallback(async (data: DriverRegisterData) => {
    try {
      const result = await authApi.driverRegister(data);
      return { success: result.success, message: result.message };
    } catch (error: any) {
      // Extract error message from API response
      let message = 'Registration failed. Please try again.';
      if (error?.response?.data?.error) {
        message = error.response.data.error;
      } else if (error?.response?.data?.message) {
        message = error.response.data.message;
      } else if (error instanceof Error && error.message) {
        // Only use error.message if it's not a generic axios error
        if (!error.message.includes('status code') && !error.message.includes('Network Error')) {
          message = error.message;
        }
      }
      return { success: false, message };
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } finally {
      setState({
        isAuthenticated: false,
        isLoading: false,
        userType: null,
        username: null,
        userId: null,
        driver: null,
      });
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        driverLogin,
        driverRegister,
        logout,
        checkSession,
        isAdmin: state.userType === 'admin',
        isDriver: state.userType === 'driver',
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export default AuthContext;
