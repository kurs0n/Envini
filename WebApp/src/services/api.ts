import axios, { AxiosInstance } from 'axios';

// Create axios instance with base configuration
const api: AxiosInstance = axios.create({
  baseURL: 'http://localhost:3000', // BackendGate default port
  timeout: 10000,
});

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('envini_jwt');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  console.log('API Request:', config.method?.toUpperCase(), config.url, config.data);
  return config;
});

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => {
    console.log('API Response:', response.status, response.config.url, response.data);
    return response;
  },
  (error) => {
    console.error('API Error:', error.response?.status, error.response?.config?.url, error.message);
    if (error.response?.status === 401) {
      // Clear invalid token
      localStorage.removeItem('envini_jwt');
      localStorage.removeItem('envini_user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  startGitHubAuth: async () => {
    try {
      console.log('Making startGitHubAuth request...');
      const response = await api.post('/auth/github/start');
      console.log('startGitHubAuth response:', response.data);
      return response.data;
    } catch (error) {
      console.error('startGitHubAuth error:', error);
      throw error;
    }
  },

  pollForToken: async (deviceCode: string) => {
    try {
      console.log('Making pollForToken request with deviceCode:', deviceCode);
      const response = await api.get(`/auth/github/poll?deviceCode=${deviceCode}`);
      console.log('pollForToken response:', response.data);
      return response.data;
    } catch (error) {
      console.error('pollForToken error:', error);
      throw error;
    }
  },

  validateSession: async (jwt: string) => {
    try {
      console.log('Making validateSession request...');
      const response = await api.get(`/auth/validate?jwt=${jwt}`);
      console.log('validateSession response:', response.data);
      return response.data;
    } catch (error) {
      console.error('validateSession error:', error);
      throw error;
    }
  },

  logout: async (jwt: string) => {
    try {
      console.log('Making logout request...');
      const response = await api.post('/auth/logout', { jwt });
      console.log('logout response:', response.data);
      return response.data;
    } catch (error) {
      console.error('logout error:', error);
      throw error;
    }
  },
};

// Repositories API
export const reposAPI = {
  listRepos: async () => {
    try {
      console.log('Making listRepos request...');
      const response = await api.get('/repos/list');
      console.log('listRepos response:', response.data);
      return response.data;
    } catch (error) {
      console.error('listRepos error:', error);
      throw error;
    }
  },

  listReposWithVersions: async () => {
    try {
      console.log('Making listReposWithVersions request...');
      const response = await api.get('/repos/list-with-versions');
      console.log('listReposWithVersions response:', response.data);
      return response.data;
    } catch (error) {
      console.error('listReposWithVersions error:', error);
      throw error;
    }
  },
};

// Secrets API
export const secretsAPI = {
  uploadSecret: async (ownerLogin: string, repoName: string, tag: string, envFileContent: string) => {
    try {
      console.log('Making uploadSecret request...');
      const response = await api.post(`/secrets/upload/${ownerLogin}/${repoName}`, {
        tag,
        envFileContent: btoa(envFileContent),
      });
      console.log('uploadSecret response:', response.data);
      return response.data;
    } catch (error) {
      console.error('uploadSecret error:', error);
      throw error;
    }
  },

  listSecretVersions: async (ownerLogin: string, repoName: string) => {
    try {
      console.log('Making listSecretVersions request...');
      const response = await api.get(`/secrets/versions/${ownerLogin}/${repoName}`);
      console.log('listSecretVersions response:', response.data);
      return response.data;
    } catch (error) {
      console.error('listSecretVersions error:', error);
      throw error;
    }
  },

  downloadSecret: async (ownerLogin: string, repoName: string, version?: number, tag?: string) => {
    try {
      console.log('Making downloadSecret request...');
      const params = new URLSearchParams();
      if (version !== undefined) params.append('version', version.toString());
      if (tag) params.append('tag', tag);
      
      const response = await api.get(`/secrets/download/${ownerLogin}/${repoName}?${params.toString()}`);
      console.log('downloadSecret response:', response.data);
      return response.data;
    } catch (error) {
      console.error('downloadSecret error:', error);
      throw error;
    }
  },

  getSecretContent: async (ownerLogin: string, repoName: string, version?: number, tag?: string) => {
    try {
      console.log('Making getSecretContent request...');
      const params = new URLSearchParams();
      if (version !== undefined) params.append('version', version.toString());
      if (tag) params.append('tag', tag);
      
      const response = await api.get(`/secrets/content/${ownerLogin}/${repoName}?${params.toString()}`);
      console.log('getSecretContent response:', response.data);
      return response.data;
    } catch (error) {
      console.error('getSecretContent error:', error);
      throw error;
    }
  },

  deleteSecret: async (ownerLogin: string, repoName: string, version?: number, deleteAll?: boolean) => {
    try {
      console.log('Making deleteSecret request...');
      const params = new URLSearchParams();
      if (version !== undefined) params.append('version', version.toString());
      if (deleteAll) params.append('all', 'true');
      
      const response = await api.delete(`/secrets/delete/${ownerLogin}/${repoName}?${params.toString()}`);
      console.log('deleteSecret response:', response.data);
      return response.data;
    } catch (error) {
      console.error('deleteSecret error:', error);
      throw error;
    }
  },
};

export default api; 