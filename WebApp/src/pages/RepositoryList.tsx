import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Search, Folder, Clock, Users, Star, GitBranch, Loader } from 'lucide-react';
import { reposAPI } from '../services/api';

interface Repository {
  id: string;
  name: string;
  owner: string;
  description?: string;
  language?: string;
  stars?: number;
  lastUpdated?: string;
  environmentCount: number;
  isPrivate?: boolean;
  versions?: SecretVersion[];
}

interface SecretVersion {
  version: number;
  tag: string;
  createdAt: string;
  uploadedBy: string;
}

export default function RepositoryList() {
  const [searchTerm, setSearchTerm] = useState('');
  const [filter, setFilter] = useState<'all' | 'private' | 'public'>('all');
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchRepositories = async () => {
      try {
        setIsLoading(true);
        setError('');
        
        // Try to get repositories with versions first
        try {
          const response = await reposAPI.listReposWithVersions();
          if (response.success && response.repositories) {
            setRepositories(response.repositories.map((repo: any) => ({
              id: repo.id?.toString() || '',
              name: repo.name || '',
              owner: repo.owner || '',
              description: repo.description || '',
              language: repo.language || 'Unknown',
              stars: repo.stars || 0,
              lastUpdated: repo.lastUpdated || 'Unknown',
              environmentCount: repo.versions?.length || 0,
              isPrivate: repo.isPrivate || false,
              versions: repo.versions || []
            })));
          } else {
            // Fallback to basic repository list
            const basicResponse = await reposAPI.listRepos();
            if (basicResponse.success && basicResponse.repositories) {
              setRepositories(basicResponse.repositories.map((repo: any) => ({
                id: repo.id?.toString() || '',
                name: repo.name || '',
                owner: repo.owner || '',
                description: repo.description || '',
                language: repo.language || 'Unknown',
                stars: repo.stars || 0,
                lastUpdated: repo.lastUpdated || 'Unknown',
                environmentCount: 0,
                isPrivate: repo.isPrivate || false
              })));
            }
          }
        } catch (err) {
          console.error('Failed to fetch repositories:', err);
          setError('Failed to load repositories. Please try again.');
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchRepositories();
  }, []);

  const filteredRepositories = repositories.filter(repo => {
    const matchesSearch = repo.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         (repo.description && repo.description.toLowerCase().includes(searchTerm.toLowerCase()));
    const matchesFilter = filter === 'all' || 
                         (filter === 'private' && repo.isPrivate) ||
                         (filter === 'public' && !repo.isPrivate);
    
    return matchesSearch && matchesFilter;
  });

  const getLanguageColor = (language: string) => {
    const colors: Record<string, string> = {
      'TypeScript': 'bg-blue-100 text-blue-800',
      'JavaScript': 'bg-yellow-100 text-yellow-800',
      'Python': 'bg-green-100 text-green-800',
      'Java': 'bg-red-100 text-red-800',
      'Go': 'bg-cyan-100 text-cyan-800',
      'Rust': 'bg-orange-100 text-orange-800',
    };
    return colors[language] || 'bg-gray-100 text-gray-800';
  };

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="flex items-center justify-center py-12">
          <Loader className="w-8 h-8 animate-spin text-green-600" />
          <span className="ml-3 text-gray-600">Loading repositories...</span>
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
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Repositories</h1>
        <p className="text-gray-600">Manage environment variables across your repositories</p>
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
            All
          </button>
          <button
            onClick={() => setFilter('public')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              filter === 'public' 
                ? 'bg-green-600 text-white' 
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            }`}
          >
            Public
          </button>
          <button
            onClick={() => setFilter('private')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              filter === 'private' 
                ? 'bg-green-600 text-white' 
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            }`}
          >
            Private
          </button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filteredRepositories.map(repo => (
          <Link
            key={repo.id}
            to={`/repositories/${repo.owner}/${repo.name}`}
            className="block bg-white rounded-lg border border-gray-200 p-6 hover:shadow-lg hover:border-gray-300 transition-all duration-200 transform hover:scale-102 animate-slideDown"
            style={{ animationDelay: `${filteredRepositories.indexOf(repo) * 50}ms` }}
          >
            <div className="flex items-start justify-between mb-3">
              <div className="flex items-center space-x-2">
                <Folder className="w-5 h-5 text-gray-600" />
                <h3 className="text-lg font-semibold text-gray-900">{repo.name}</h3>
              </div>
              <div className="flex items-center space-x-2">
                {repo.isPrivate ? (
                  <span className="px-2 py-1 text-xs font-medium bg-red-100 text-red-800 rounded-full">
                    Private
                  </span>
                ) : (
                  <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 rounded-full">
                    Public
                  </span>
                )}
              </div>
            </div>
            
            <p className="text-gray-600 text-sm mb-4 line-clamp-2">{repo.description || 'No description available'}</p>
            
            <div className="flex items-center justify-between text-sm text-gray-500 mb-3">
              <div className="flex items-center space-x-4">
                <span className={`px-2 py-1 text-xs font-medium rounded-full ${getLanguageColor(repo.language || 'Unknown')}`}>
                  {repo.language || 'Unknown'}
                </span>
                {repo.stars !== undefined && (
                  <div className="flex items-center space-x-1">
                    <Star className="w-4 h-4" />
                    <span>{repo.stars}</span>
                  </div>
                )}
              </div>
              {repo.lastUpdated && (
                <div className="flex items-center space-x-1">
                  <Clock className="w-4 h-4" />
                  <span>{repo.lastUpdated}</span>
                </div>
              )}
            </div>
            
            <div className="flex items-center justify-between pt-3 border-t border-gray-100">
              <div className="flex items-center space-x-1 text-sm text-gray-600">
                <GitBranch className="w-4 h-4" />
                <span>{repo.environmentCount} environments</span>
              </div>
              <div className="flex items-center space-x-1 text-sm text-gray-600">
                <Users className="w-4 h-4" />
                <span>{repo.owner}</span>
              </div>
            </div>
          </Link>
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