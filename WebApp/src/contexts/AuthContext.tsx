import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { authAPI } from '../services/api';

interface User {
  id: string;
  login: string;
  name: string;
  avatar_url: string;
  html_url: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: () => Promise<void>;
  logout: () => Promise<void>;
  startGitHubAuth: () => Promise<{ deviceCode: string; userCode: string; verificationUri: string }>;
  pollForToken: (deviceCode: string) => Promise<{ success: boolean; jwt?: string; user?: User; sessionId?: string }>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is already logged in
    const jwt = localStorage.getItem('envini_jwt');
    const savedUser = localStorage.getItem('envini_user');
    
    if (jwt && savedUser) {
      // Validate the session
      authAPI.validateSession(jwt)
        .then((response) => {
          if (response.valid) {
            setUser(JSON.parse(savedUser));
            setIsAuthenticated(true);
          } else {
            // Clear invalid session
            localStorage.removeItem('envini_jwt');
            localStorage.removeItem('envini_user');
          }
        })
        .catch((error) => {
          console.error('Session validation error:', error);
          // Clear invalid session
          localStorage.removeItem('envini_jwt');
          localStorage.removeItem('envini_user');
        })
        .finally(() => {
          setIsLoading(false);
        });
    } else {
      setIsLoading(false);
    }
  }, []);

  const startGitHubAuth = async () => {
    try {
      console.log('Calling startGitHubAuth API...');
      const response = await authAPI.startGitHubAuth();
      console.log('startGitHubAuth API response:', response);
      
      if (!response.deviceCode || !response.userCode || !response.verificationUri) {
        throw new Error('Invalid response from GitHub auth API');
      }
      
      return {
        deviceCode: response.deviceCode,
        userCode: response.userCode,
        verificationUri: response.verificationUri,
      };
    } catch (error) {
      console.error('startGitHubAuth error:', error);
      if (error instanceof Error) {
        throw new Error(`Failed to start GitHub authentication: ${error.message}`);
      }
      throw new Error('Failed to start GitHub authentication');
    }
  };

  const pollForToken = async (deviceCode: string) => {
    try {
      console.log('Starting to poll for token with deviceCode:', deviceCode);
      
      // Poll for up to 10 minutes (120 attempts with 5-second intervals)
      const maxAttempts = 120;
      let attempts = 0;
      
      while (attempts < maxAttempts) {
        try {
          console.log(`Poll attempt ${attempts + 1}/${maxAttempts}`);
          const response = await authAPI.pollForToken(deviceCode);
          console.log('Poll response:', response);
          
          if (response.jwt && response.sessionId) {
            // Store the JWT
            localStorage.setItem('envini_jwt', response.jwt);
            
            const responseUser = await authAPI.getUser(response.jwt);
            console.log('User response:', responseUser);
            // Create a basic user object from the session info
            const user = {
              id: response.sessionId,
              login: responseUser.userLogin,
              name: responseUser.name,
              avatar_url: responseUser.avatarUrl,
              html_url: responseUser.htmlUrl, // This will be set after fetching user details
            };
            
            localStorage.setItem('envini_user', JSON.stringify(user));
            setUser(user);
            setIsAuthenticated(true);
            
            return { success: true, jwt: response.jwt, user: user, sessionId: response.sessionId };
          } else if (response.error === 'authorization_pending') {
            console.log('Authorization still pending, continuing to poll...');
            // Continue polling - this is expected
          } else if (response.error) {
            console.log('Poll returned error:', response.error);
            throw new Error(response.error);
          }
        } catch (pollError) {
          console.error('Poll error:', pollError);
          if (pollError instanceof Error && pollError.message === 'authorization_pending') {
            // This is expected, continue polling
            console.log('Authorization pending, continuing to poll...');
          } else {
            throw pollError;
          }
        }
        
        // Wait 5 seconds before next attempt
        await new Promise(resolve => setTimeout(resolve, 5000));
        attempts++;
      }
      
      // If we get here, we've timed out
      console.log('Authentication timed out after', maxAttempts, 'attempts');
      return { success: false };
    } catch (error) {
      console.error('pollForToken error:', error);
      if (error instanceof Error) {
        throw new Error(`Authentication failed: ${error.message}`);
      }
      throw new Error('Authentication failed');
    }
  };

  const login = async () => {
    // This is now handled by startGitHubAuth and pollForToken
    throw new Error('Use startGitHubAuth and pollForToken instead of login');
  };

  const logout = async () => {
    try {
      const jwt = localStorage.getItem('envini_jwt');
      if (jwt) {
        await authAPI.logout(jwt);
      }
    } catch (error) {
      console.error('Failed to logout:', error);
    } finally {
      // Clear local storage regardless of API call success
      localStorage.removeItem('envini_jwt');
      localStorage.removeItem('envini_user');
      setUser(null);
      setIsAuthenticated(false);
    }
  };

  return (
    <AuthContext.Provider value={{ 
      user, 
      isAuthenticated, 
      isLoading, 
      login, 
      logout, 
      startGitHubAuth, 
      pollForToken 
    }}>
      {children}
    </AuthContext.Provider>
  );
}