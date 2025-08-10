import React, { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import {
  ChevronLeft,
  Download,
  Upload,
  AlertCircle,
  GitBranch,
  FileText,
  Loader,
  Trash2,
} from "lucide-react";
import { secretsAPI } from "../services/api";

interface SecretVersion {
  id: string;
  version: number;
  tag: string;
  checksum: string;
  uploadedBy: string;
  createdAt: string;
}

interface EnvironmentVariable {
  key: string;
  value: string;
  description?: string;
}

export default function RepositoryDetails() {
  const { owner, repo } = useParams<{ owner: string; repo: string }>();
  const [versions, setVersions] = useState<SecretVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<SecretVersion | null>(
    null
  );
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [uploading, setUploading] = useState(false);
  const [uploadTag, setUploadTag] = useState("development");
  const [loadingVariables, setLoadingVariables] = useState(false);

  useEffect(() => {
    const fetchSecretVersions = async () => {
      if (!owner || !repo) return;

      try {
        setIsLoading(true);
        setError("");

        const response = await secretsAPI.listSecretVersions(owner, repo);

        if (response.versions && response.versions.length > 0) {
          const versionsWithIds = response.versions.map((version: any) => ({
            ...version,
            id: `${version.tag}-${version.version}`,
          }));

          setVersions(versionsWithIds);
          setSelectedVersion(versionsWithIds[0]); // Select latest version
          await loadEnvironmentVariables(versionsWithIds[0]);
        }
      } catch (err) {
        console.error("Failed to fetch secret versions:", err);
        setError("Failed to load repository details. Please try again.");
      } finally {
        setIsLoading(false);
      }
    };

    fetchSecretVersions();
  }, [owner, repo]);

  const loadEnvironmentVariables = async (version: SecretVersion) => {
    try {
      if (!version) return;

      setLoadingVariables(true);

      const response = await secretsAPI.getSecretContent(
        owner!,
        repo!,
        version.version,
        version.tag
      );

      if (response.success && response.envFileContent) {
        const variables = parseEnvContent(response.envFileContent);
        setSelectedVersion({
          ...version,
          variables,
        });
      }
    } catch (err) {
      console.error("Failed to load environment variables:", err);
    } finally {
      setLoadingVariables(false);
    }
  };

  const parseEnvContent = (content: {
    type: string;
    data: number[];
  }): EnvironmentVariable[] => {
    // Convert number array to string using TextDecoder
    const decoder = new TextDecoder("utf-8");
    const uint8Array = new Uint8Array(content.data);
    const text = decoder.decode(uint8Array);

    const lines = text.split("\n");
    const variables: EnvironmentVariable[] = [];
    let currentDescription = "";

    lines.forEach((line) => {
      const trimmedLine = line.trim();
      if (trimmedLine.startsWith("#")) {
        currentDescription = trimmedLine.substring(1).trim();
      } else if (trimmedLine.includes("=")) {
        const [key, ...valueParts] = trimmedLine.split("=");
        const value = valueParts.join("=").replace(/^"(.*)"$/, "$1"); // Remove quotes if present
        variables.push({
          key: key.trim(),
          value: value.trim(),
          description: currentDescription || undefined,
        });
        currentDescription = "";
      }
    });

    return variables;
  };

  const downloadEnvFile = async () => {
    if (!selectedVersion || !owner || !repo) return;

    try {
      const response = await secretsAPI.getSecretContent(
        owner,
        repo,
        selectedVersion.version,
        selectedVersion.tag
      );

      if (response.success && response.envFileContent) {
        // Convert Buffer data to string
        const decoder = new TextDecoder("utf-8");
        const content = decoder.decode(
          new Uint8Array(response.envFileContent.data)
        );

        // Create blob with proper line endings
        const blob = new Blob([content], { type: "text/plain;charset=utf-8" });

        // Create download link
        const downloadUrl = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = `.env.${selectedVersion.tag}.v${selectedVersion.version}`;

        // Append, click, and cleanup
        document.body.appendChild(link);
        link.click();
        window.URL.revokeObjectURL(downloadUrl);
        document.body.removeChild(link);
      } else {
        console.error('Failed to download: No content in response', response);
      }
    } catch (err) {
      console.error("Failed to download environment file:", err);
    }
  };

  const handleFileUpload = (file: File) => {
    const reader = new FileReader();
    reader.onload = async (e) => {
      const content = e.target?.result as string;
      await uploadEnvFile(content);
    };
    reader.readAsText(file);
  };

  const uploadEnvFile = async (content: string) => {
    if (!owner || !repo) return;

    try {
      setUploading(true);

      const response = await secretsAPI.uploadSecret(
        owner,
        repo,
        uploadTag,
        content
      );

      if (response.success) {
        // Refresh the versions list
        window.location.reload();
      } else {
        setError(response.error || "Failed to upload environment file");
      }
    } catch (err) {
      console.error("Failed to upload environment file:", err);
      setError("Failed to upload environment file. Please try again.");
    } finally {
      setUploading(false);
      setShowUploadModal(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);

    const files = Array.from(e.dataTransfer.files);
    if (files.length > 0) {
      handleFileUpload(files[0]);
    }
  };

  const handleDeleteVersion = async () => {
    if (!selectedVersion) return;

    try {
      await secretsAPI.deleteSecret(
        owner!, 
        repo!,
        selectedVersion.version,
        selectedVersion.tag
      )
      window.location.reload();
    } catch (err) {
      console.error("Failed to delete version:", err);
      setError("Failed to delete version. Please try again.");
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="flex items-center justify-center py-12">
          <Loader className="w-8 h-8 animate-spin text-green-600" />
          <span className="ml-3 text-gray-600">
            Loading repository details...
          </span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="text-center py-12">
          <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <AlertCircle className="w-6 h-6 text-red-600" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            Error loading repository
          </h3>
          <p className="text-gray-600 mb-4">{error}</p>
          <Link
            to="/repositories"
            className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors"
          >
            Back to Repositories
          </Link>
        </div>
      </div>
    );
  }

  if (!versions.length) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-6">
        <div className="mb-8">
          <Link
            to="/repositories"
            className="inline-flex items-center text-gray-600 hover:text-gray-900 mb-4"
          >
            <ChevronLeft className="w-4 h-4 mr-1" />
            Back to Repositories
          </Link>

          <div>
            <h1 className="text-3xl font-bold text-gray-900 mb-2">
              {owner}/{repo}
            </h1>
          </div>
        </div>

        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <FileText className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <h3 className="text-xl font-medium text-gray-900 mb-2">
            No Environment Variables Yet
          </h3>
          <p className="text-gray-600 mb-6">
            Get started by uploading your first environment file!
          </p>
          <button
            onClick={() => setShowUploadModal(true)}
            className="bg-green-600 text-white px-6 py-3 rounded-lg hover:bg-green-700 transition-colors flex items-center space-x-2 mx-auto"
          >
            <Upload className="w-5 h-5" />
            <span>Upload First Environment</span>
          </button>
        </div>

        {/* Keep the upload modal code */}
        {showUploadModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-md mx-4">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                Upload Environment File
              </h3>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Environment Tag
                  </label>
                  <select
                    value={uploadTag}
                    onChange={(e) => setUploadTag(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-green-500 focus:border-transparent"
                  >
                    <option value="development">Development</option>
                    <option value="staging">Staging</option>
                    <option value="production">Production</option>
                    <option value="testing">Testing</option>
                  </select>
                </div>

                <div
                  className={`border-2 border-dashed rounded-lg p-6 text-center ${
                    dragOver
                      ? "border-green-500 bg-green-50"
                      : "border-gray-300"
                  }`}
                  onDragOver={(e) => {
                    e.preventDefault();
                    setDragOver(true);
                  }}
                  onDragLeave={() => setDragOver(false)}
                  onDrop={handleDrop}
                >
                  <Upload className="w-8 h-8 text-gray-400 mx-auto mb-2" />
                  <p className="text-gray-600 mb-2">
                    Drag and drop your .env file here
                  </p>
                  <p className="text-sm text-gray-500">or</p>
                  <input
                    type="file"
                    accept=".env,.env.*"
                    onChange={(e) => {
                      const file = e.target.files?.[0];
                      if (file) handleFileUpload(file);
                    }}
                    className="hidden"
                    id="file-upload"
                  />
                  <label
                    htmlFor="file-upload"
                    className="mt-2 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 cursor-pointer"
                  >
                    Choose File
                  </label>
                </div>
              </div>

              <div className="flex justify-end space-x-3 mt-6">
                <button
                  onClick={() => setShowUploadModal(false)}
                  className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                >
                  Cancel
                </button>
                {uploading && (
                  <div className="flex items-center space-x-2 px-4 py-2">
                    <Loader className="w-4 h-4 animate-spin" />
                    <span>Uploading...</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    );
  }
  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      {/* Header */}
      <div className="mb-8">
        <Link
          to="/repositories"
          className="inline-flex items-center text-gray-600 hover:text-gray-900 mb-4"
        >
          <ChevronLeft className="w-4 h-4 mr-1" />
          Back to Repositories
        </Link>

        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 mb-2">
              {owner}/{repo}
            </h1>
            <p className="text-gray-600">
              Manage environment variables for this repository
            </p>
          </div>

          <div className="flex space-x-3">
            <button
              onClick={() => setShowUploadModal(true)}
              className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors flex items-center space-x-2"
            >
              <Upload className="w-4 h-4" />
              <span>Upload Environment</span>
            </button>

            {selectedVersion && (
              <>
                <button
                  onClick={downloadEnvFile}
                  className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"
                >
                  <Download className="w-4 h-4" />
                  <span>Download</span>
                </button>
                <button
                  onClick={handleDeleteVersion}
                  className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors flex items-center space-x-2"
                >
                  <Trash2 className="w-4 h-4" />
                  <span>Delete</span>
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Environment Tabs */}
      <div className="mb-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            {/* Group environments by tag */}
            {Array.from(new Set(versions.map((env) => env.tag))).map((tag) => (
              <div key={tag} className="space-y-2">
                <button
                  className={`py-2 px-1 border-b-2 font-medium text-sm ${
                    selectedVersion?.tag === tag
                      ? "border-green-500 text-green-600"
                      : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                  }`}
                >
                  {tag.charAt(0).toUpperCase() + tag.slice(1)}
                </button>
                {/* List versions under each tag */}
                <div className="flex flex-col space-y-1">
                  {versions
                    .filter((env) => env.tag === tag)
                    .sort((a, b) => b.version - a.version)
                    .map((env) => (
                      <button
                        key={env.id}
                        onClick={() => {
                          setSelectedVersion(env);
                          loadEnvironmentVariables(env);
                        }}
                        className={`ml-4 py-1 px-2 text-xs rounded-md ${
                          selectedVersion?.id === env.id
                            ? "bg-green-100 text-green-800"
                            : "text-gray-500 hover:bg-gray-100"
                        }`}
                      >
                        Version {env.version}
                      </button>
                    ))}
                </div>
              </div>
            ))}
          </nav>
        </div>
      </div>

      {/* Environment Details */}
      {selectedVersion && (
        <div className="bg-white rounded-lg border border-gray-200">
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <h2 className="text-xl font-semibold text-gray-900">
                  {selectedVersion.tag}
                </h2>
              </div>

              <div className="flex items-center space-x-4 text-sm text-gray-500">
                <div className="flex items-center space-x-1">
                  <GitBranch className="w-4 h-4" />
                  <span>{selectedVersion.version}</span>
                </div>
              </div>
            </div>
          </div>

          {/* Variables List */}
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-gray-900">
                Environment Variables
              </h3>
              <span className="text-sm text-gray-500">
                {selectedVersion.variables?.length || 0} variables
              </span>
            </div>

            {loadingVariables ? (
              <div className="flex items-center justify-center py-12">
                <Loader className="w-8 h-8 animate-spin text-green-600" />
                <span className="ml-3 text-gray-600">Loading variables...</span>
              </div>
            ) : (
              <>
                <div className="space-y-3">
                  {selectedVersion.variables?.map((variable, index) => (
                    <div key={index} className="bg-gray-50 p-4 rounded-lg">
                      <div className="flex items-start justify-between">
                        <div>
                          <h4 className="text-sm font-medium text-gray-900">
                            {variable.key}
                          </h4>
                          <p className="mt-1 text-sm text-gray-600 font-mono">
                            {variable.value}
                          </p>
                          {variable.description && (
                            <p className="mt-1 text-xs text-gray-500">
                              {variable.description}
                            </p>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>

                {(!selectedVersion.variables ||
                  selectedVersion.variables.length === 0) && (
                  <div className="text-center py-8">
                    <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                    <h3 className="text-lg font-medium text-gray-900 mb-2">
                      No environment variables
                    </h3>
                    <p className="text-gray-600">
                      Upload an environment file to get started.
                    </p>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      )}

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Upload Environment File
            </h3>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Environment Tag
                </label>
                <select
                  value={uploadTag}
                  onChange={(e) => setUploadTag(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-green-500 focus:border-transparent"
                >
                  <option value="development">Development</option>
                  <option value="staging">Staging</option>
                  <option value="production">Production</option>
                  <option value="testing">Testing</option>
                </select>
              </div>

              <div
                className={`border-2 border-dashed rounded-lg p-6 text-center ${
                  dragOver ? "border-green-500 bg-green-50" : "border-gray-300"
                }`}
                onDragOver={(e) => {
                  e.preventDefault();
                  setDragOver(true);
                }}
                onDragLeave={() => setDragOver(false)}
                onDrop={handleDrop}
              >
                <Upload className="w-8 h-8 text-gray-400 mx-auto mb-2" />
                <p className="text-gray-600 mb-2">
                  Drag and drop your .env file here
                </p>
                <p className="text-sm text-gray-500">or</p>
                <input
                  type="file"
                  accept=".env,.env.*"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) handleFileUpload(file);
                  }}
                  className="hidden"
                  id="file-upload"
                />
                <label
                  htmlFor="file-upload"
                  className="mt-2 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 cursor-pointer"
                >
                  Choose File
                </label>
              </div>
            </div>

            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => setShowUploadModal(false)}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
              >
                Cancel
              </button>
              {uploading && (
                <div className="flex items-center space-x-2 px-4 py-2">
                  <Loader className="w-4 h-4 animate-spin" />
                  <span>Uploading...</span>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}