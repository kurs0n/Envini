import { useState, useEffect } from 'react';
import { Search, Folder, Users, Loader, ExternalLink, Lock, Globe } from 'lucide-react';
import { reposAPI } from '../services/api';
import { useAuth } from '../contexts/AuthContext';

interface GitHubRepository {
  id: number;
  name: string;
  fullName: string;
  htmlUrl: string;
  description: string;
  private: boolean;
  ownerLogin: string;
  ownerAvatarUrl: string;
  language?: string;
  stargazersCount?: number;
  updatedAt?: string;
  fork?: boolean;
  archived?: boolean;
}

export default function GitHubRepositories() {
  const { isAuthenticated } = useAuth();
  const [searchTerm, setSearchTerm] = useState('');
  const [filter, setFilter] = useState<'all' | 'private' | 'public'>('all');
  const [repositories, setRepositories] = useState<GitHubRepository[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchGitHubRepositories = async () => {
      if (!isAuthenticated) {
        setError('Please log in to view your GitHub repositories');
        setIsLoading(false);
        return;
      }

      try {
        setIsLoading(true);
        setError('');
        
        const response = await reposAPI.listRepos();
        console.log('GitHub repositories response:', response);
        
        if (response.error) {
          setError(response.errorDescription || response.error || 'Failed to fetch repositories');
        } else if (response.repos) {
          setRepositories(response.repos);
        } else {
          setError('No repositories found');
        }
      } catch (err) {
        console.error('Failed to fetch GitHub repositories:', err);
        setError('Failed to load repositories. Please try again.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchGitHubRepositories();
  }, [isAuthenticated]);

  const filteredRepositories = repositories.filter(repo => {
    const matchesSearch = repo.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         (repo.description && repo.description.toLowerCase().includes(searchTerm.toLowerCase())) ||
                         repo.ownerLogin.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesFilter = filter === 'all' || 
                         (filter === 'private' && repo.private) ||
                         (filter === 'public' && !repo.private);
    
    return matchesSearch && matchesFilter;
  });

  if (!isAuthenticated) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="text-center py-12">
          <div className="w-12 h-12 bg-yellow-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <Lock className="w-6 h-6 text-yellow-600" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">Authentication Required</h3>
          <p className="text-gray-600">Please log in to view your GitHub repositories.</p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="flex items-center justify-center py-12">
          <Loader className="w-8 h-8 animate-spin text-green-600" />
          <span className="ml-3 text-gray-600">Loading GitHub repositories...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="text-center py-12">
          <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <Folder className="w-6 h-6 text-red-600" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">Error loading repositories</h3>
          <p className="text-gray-600 mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">GitHub Repositories</h1>
        <p className="text-gray-600">Your repositories from GitHub</p>
      </div>

      <div className="mb-6 flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
          <input
            type="text"
            placeholder="Search repositories..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
          />
        </div>
        
        <div className="flex gap-2">
          <button
            onClick={() => setFilter('all')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              filter === 'all' 
                ? 'bg-green-600 text-white' 
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            }`}
          >
            All ({repositories.length})
          </button>
          <button
            onClick={() => setFilter('public')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              filter === 'public' 
                ? 'bg-green-600 text-white' 
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            }`}
          >
            Public ({repositories.filter(r => !r.private).length})
          </button>
          <button
            onClick={() => setFilter('private')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              filter === 'private' 
                ? 'bg-green-600 text-white' 
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            }`}
          >
            Private ({repositories.filter(r => r.private).length})
          </button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filteredRepositories.map(repo => (
          <div
            key={repo.id}
            className="bg-white rounded-lg border border-gray-200 p-6 hover:shadow-lg hover:border-gray-300 transition-all duration-200 transform hover:scale-102 animate-slideDown"
            style={{ animationDelay: `${filteredRepositories.indexOf(repo) * 50}ms` }}
          >
            <div className="flex items-start justify-between mb-3">
              <div className="flex items-center space-x-2">
                <Folder className="w-5 h-5 text-gray-600" />
                <h3 className="text-lg font-semibold text-gray-900">{repo.name}</h3>
              </div>
              <div className="flex items-center space-x-2">
                {repo.private ? (
                  <span className="px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded-full flex items-center">
                    <Lock className="w-3 h-3 mr-1" />
                    Private
                  </span>
                ) : (
                  <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 rounded-full flex items-center">
                    <Globe className="w-3 h-3 mr-1" />
                    Public
                  </span>
                )}
              </div>
            </div>
            
            <p className="text-gray-600 text-sm mb-4 line-clamp-2">{repo.description || 'No description available'}</p>
            <div className="flex items-center justify-between pt-3 border-t border-gray-100">
              <div className="flex items-center space-x-1 text-sm text-gray-600">
                <Users className="w-4 h-4" />
                <span>{repo.ownerLogin}</span>
              </div>
              <a
                href={repo.htmlUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center space-x-1 text-sm text-green-600 hover:text-green-700 transition-colors"
              >
                <ExternalLink className="w-4 h-4" />
                <span>View on GitHub</span>
              </a>
            </div>
          </div>
        ))}
      </div>

      {filteredRepositories.length === 0 && (
        <div className="text-center py-12">
          <Folder className="w-12 h-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No repositories found</h3>
          <p className="text-gray-600">Try adjusting your search or filter criteria.</p>
        </div>
      )}
    </div>
  );
} 