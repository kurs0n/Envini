import React, { useState } from 'react';
import { Github, Settings, Shield, Zap, Copy, Check, ExternalLink } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

export default function Login() {
  const { startGitHubAuth, pollForToken } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [authStep, setAuthStep] = useState<'initial' | 'device-code' | 'polling' | 'success' | 'error'>('initial');
  const [deviceCode, setDeviceCode] = useState('');
  const [userCode, setUserCode] = useState('');
  const [verificationUri, setVerificationUri] = useState('');
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  const handleGitHubLogin = async () => {
    setIsLoading(true);
    setError('');
    
    try {
      console.log('Starting GitHub authentication...');
      const authData = await startGitHubAuth();
      console.log('Auth data received:', authData);
      
      setDeviceCode(authData.deviceCode);
      setUserCode(authData.userCode);
      setVerificationUri(authData.verificationUri);
      setAuthStep('device-code');
      
      // Automatically open the verification URL
      window.open(authData.verificationUri, '_blank');
      
      // Start polling for token
      setAuthStep('polling');
      console.log('Starting to poll for token...');
      
      const pollResult = await pollForToken(authData.deviceCode);
      console.log('Poll result:', pollResult);
      
      if (pollResult.success) {
        setAuthStep('success');
        // Redirect to repositories page after successful authentication
        setTimeout(() => {
          window.location.href = '/repositories';
        }, 1500);
      } else {
        setError('Authentication timed out. Please try again.');
        setAuthStep('error');
      }
    } catch (err) {
      console.error('Authentication error:', err);
      const errorMessage = err instanceof Error ? err.message : 'Failed to start GitHub authentication. Please try again.';
      setError(errorMessage);
      setAuthStep('error');
    } finally {
      setIsLoading(false);
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy to clipboard:', err);
    }
  };

  const openVerificationUri = () => {
    window.open(verificationUri, '_blank');
  };

  const renderAuthStep = () => {
    switch (authStep) {
      case 'device-code':
        return (
          <div className="space-y-4">
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h3 className="text-lg font-semibold text-blue-900 mb-2">Complete GitHub Authentication</h3>
              <p className="text-blue-700 mb-4">
                Please visit the verification URL and enter the code below:
              </p>
              
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-blue-900 mb-1">Verification URL:</label>
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      value={verificationUri}
                      readOnly
                      className="flex-1 px-3 py-2 border border-blue-300 rounded-md bg-white text-blue-900 text-sm"
                    />
                    <button
                      onClick={openVerificationUri}
                      className="px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                    >
                      <ExternalLink className="w-4 h-4" />
                    </button>
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-blue-900 mb-1">Code:</label>
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      value={userCode}
                      readOnly
                      className="flex-1 px-3 py-2 border border-blue-300 rounded-md bg-white text-blue-900 text-sm font-mono"
                    />
                    <button
                      onClick={() => copyToClipboard(userCode)}
                      className="px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                    >
                      {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
              </div>
              
              <p className="text-sm text-blue-600 mt-3">
                After entering the code, you'll be redirected back here automatically.
              </p>
            </div>
          </div>
        );
      
      case 'polling':
        return (
          <div className="text-center space-y-4">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-green-600 mx-auto"></div>
            <p className="text-gray-600">Waiting for authentication...</p>
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-left">
              <h4 className="font-medium text-blue-900 mb-2">Next Steps:</h4>
              <ol className="text-sm text-blue-800 space-y-2">
                <li>1. A new tab should have opened with GitHub's device verification page</li>
                <li>2. If the tab didn't open, click the "Open GitHub" button below</li>
                <li>3. Enter the code: <span className="font-mono font-bold">{userCode}</span></li>
                <li>4. Complete the GitHub authentication</li>
                <li>5. This page will automatically detect when you're done</li>
              </ol>
              <button
                onClick={openVerificationUri}
                className="mt-3 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors text-sm"
              >
                Open GitHub Verification Page
              </button>
            </div>
          </div>
        );
      
      case 'success':
        return (
          <div className="text-center space-y-4">
            <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto">
              <Check className="w-6 h-6 text-green-600" />
            </div>
            <p className="text-green-600 font-medium">Authentication successful!</p>
            <p className="text-sm text-gray-600">Redirecting to repositories...</p>
          </div>
        );
      
      case 'error':
        return (
          <div className="text-center space-y-4">
            <p className="text-red-600">{error}</p>
            <button
              onClick={handleGitHubLogin}
              className="bg-gray-900 hover:bg-gray-800 text-white font-medium py-2 px-4 rounded-lg transition-colors"
            >
              Try Again
            </button>
          </div>
        );
      
      default:
        return (
          <button
            onClick={handleGitHubLogin}
            disabled={isLoading}
            className="w-full bg-gray-900 hover:bg-gray-800 disabled:bg-gray-400 text-white font-medium py-3 px-4 rounded-lg transition-all duration-200 flex items-center justify-center space-x-3 hover:scale-105 disabled:hover:scale-100"
          >
            {isLoading ? (
              <>
                <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                <span>Connecting...</span>
              </>
            ) : (
              <>
                <Github className="w-5 h-5" />
                <span>Continue with GitHub</span>
              </>
            )}
          </button>
        );
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 flex items-center justify-center px-4 animate-fadeIn">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center animate-slideDown">
          <div className="flex justify-center mb-4">
            <div className="w-16 h-16 bg-green-600 rounded-2xl flex items-center justify-center shadow-lg animate-scaleIn">
              <Settings className="w-8 h-8 text-white" />
            </div>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Welcome to envini</h1>
          <p className="text-gray-600">Secure environment management for your repositories</p>
        </div>

        <div className="bg-white rounded-2xl shadow-xl p-8 space-y-6 animate-slideUp">
          <div className="space-y-4">
            <div className="flex items-center space-x-3 text-gray-700">
              <Shield className="w-5 h-5 text-green-600" />
              <span className="text-sm">Secure environment variable management</span>
            </div>
            <div className="flex items-center space-x-3 text-gray-700">
              <Zap className="w-5 h-5 text-green-600" />
              <span className="text-sm">Multiple environment support</span>
            </div>
            <div className="flex items-center space-x-3 text-gray-700">
              <Github className="w-5 h-5 text-green-600" />
              <span className="text-sm">Seamless GitHub integration</span>
            </div>
          </div>

          {renderAuthStep()}

          <p className="text-xs text-gray-500 text-center">
            By continuing, you agree to our Terms of Service and Privacy Policy
          </p>
        </div>
      </div>
    </div>
  );
}