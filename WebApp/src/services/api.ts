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
  return config;
});

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
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
      const response = await api.post('/auth/github/start');
      return response.data;
    } catch (error) {
      throw error;
    }
  },

  pollForToken: async (deviceCode: string) => {
    try {
      const response = await api.get(`/auth/github/poll?deviceCode=${deviceCode}`);
      return response.data;
    } catch (error) {
      console.error('pollForToken error:', error);
      throw error;
    }
  },

  validateSession: async (jwt: string) => {
    try {
      const response = await api.get(`/auth/validate?jwt=${jwt}`);
      return response.data;
    } catch (error) {
      console.error('validateSession error:', error);
      throw error;
    }
  },

  logout: async (jwt: string) => {
    try {
      const response = await api.post('/auth/logout', { jwt });
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
      const response = await api.get('/repos/list');
      return response.data;
    } catch (error) {
      console.error('listRepos error:', error);
      throw error;
    }
  },

  listReposWithVersions: async () => {
    try {
      const response = await api.get('/repos/list-with-versions');
      return response.data;
    } catch (error) {
      console.error('listReposWithVersions error:', error);
      throw error;
    }
  },
  
  listAllTheReposWithSecretsCheck: async () => {
    const githubResponse = await reposAPI.listRepos();
    const databaseResponse = await reposAPI.listReposWithVersions();
    const reposWithSecrets = [] as any[];

    const dbRepoIdentifiers = new Set<string>();
    databaseResponse.repositories.forEach((dbRepo: any) => {
      dbRepoIdentifiers.add(`${dbRepo.ownerLogin}/${dbRepo.repoName}`);
    });

    githubResponse.repos.forEach((repo: any) => {
      const repoIdentifier = `${repo.ownerLogin}/${repo.name}`;
      const hasSecrets = dbRepoIdentifiers.has(repoIdentifier);

      reposWithSecrets.push({
        ...repo,
        hasSecrets: hasSecrets,
      });
    });

    return reposWithSecrets;
  }
}

// Secrets API
export const secretsAPI = {
  uploadSecret: async (ownerLogin: string, repoName: string, tag: string, envFileContent: string) => {
    try {
      const response = await api.post(`/secrets/upload/${ownerLogin}/${repoName}`, {
        tag,
        envFileContent: btoa(envFileContent),
      });
      return response.data;
    } catch (error) {
      console.error('uploadSecret error:', error);
      throw error;
    }
  },

  listSecretVersions: async (ownerLogin: string, repoName: string) => {
    try {
      const response = await api.get(`/secrets/versions/${ownerLogin}/${repoName}`);
      return response.data;
    } catch (error) {
      console.error('listSecretVersions error:', error);
      throw error;
    }
  },

  downloadSecret: async (ownerLogin: string, repoName: string, version?: number, tag?: string) => {
    try {
      const params = new URLSearchParams();
      if (version !== undefined) params.append('version', version.toString());
      if (tag) params.append('tag', tag);
      
      const response = await api.get(`/secrets/download/${ownerLogin}/${repoName}?${params.toString()}`);
      return response.data;
    } catch (error) {
      console.error('downloadSecret error:', error);
      throw error;
    }
  },

  getSecretContent: async (ownerLogin: string, repoName: string, version?: number, tag?: string) => {
    try {
      const params = new URLSearchParams();
      if (version !== undefined) params.append('version', version.toString());
      if (tag) params.append('tag', tag);
      const response = await api.get(`/secrets/content/${ownerLogin}/${repoName}?${params.toString()}`);
      return response.data;
    } catch (error) {
      console.error('getSecretContent error:', error);
      throw error;
    }
  },

  deleteSecret: async (ownerLogin: string, repoName: string, version?: number, deleteAll?: boolean) => {
    try {
      const params = new URLSearchParams();
      if (version !== undefined) params.append('version', version.toString());
      if (deleteAll) params.append('all', 'true');
      
      const response = await api.delete(`/secrets/delete/${ownerLogin}/${repoName}?${params.toString()}`);
      return response.data;
    } catch (error) {
      console.error('deleteSecret error:', error);
      throw error;
    }
  },
};

export default api; 