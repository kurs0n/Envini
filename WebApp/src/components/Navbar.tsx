import { Link } from 'react-router-dom';
import { Settings, LogOut, User, Folder, Github } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();

  return (
    <nav className="bg-white border-b border-gray-200 px-4 py-3 animate-slideDown">
      <div className="max-w-7xl mx-auto flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Link to="/" className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-green-600 rounded-lg flex items-center justify-center transition-transform duration-200 hover:scale-110">
              <Settings className="w-5 h-5 text-white" />
            </div>
            <span className="text-xl font-bold text-gray-900">envini</span>
          </Link>
          
          {isAuthenticated && (
            <div className="flex items-center space-x-6 ml-8">
              <Link 
                to="/repositories" 
                className="flex items-center space-x-2 text-gray-600 hover:text-gray-900 transition-all duration-200 hover:scale-105"
              >
                <Folder className="w-4 h-4" />
                <span>Secrets</span>
              </Link>
              <Link 
                to="/github-repositories" 
                className="flex items-center space-x-2 text-gray-600 hover:text-gray-900 transition-all duration-200 hover:scale-105"
              >
                <Github className="w-4 h-4" />
                <span>GitHub Repos</span>
              </Link>
            </div>
          )}
        </div>

        {isAuthenticated && user && (
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-3">
              <img 
                src={user.avatar_url} 
                alt={user.name}
                className="w-8 h-8 rounded-full"
              />
              <div className="text-sm">
                <div className="font-medium text-gray-900">{user.name}</div>
                <div className="text-gray-500">@{user.login}</div>
              </div>
            </div>
            <button
              onClick={logout}
              className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
              title="Logout"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        )}
      </div>
    </nav>
  );
}