import { Injectable } from '@nestjs/common';
import { SecretOperationClientService } from '../grpc/secretoperation-client.service';
import { AuthService } from '../auth/auth.service';

export interface UploadSecretResult {
  success?: boolean;
  version?: number;
  checksum?: string;
  error?: string;
  errorDescription?: string;
}

export interface ListSecretVersionsResult {
  versions?: Array<{
    version: number;
    tag: string;
    checksum: string;
    uploadedBy: string;
    createdAt: string;
  }>;
  error?: string;
  errorDescription?: string;
}

export interface DownloadSecretResult {
  success?: boolean;
  version?: number;
  tag?: string;
  envFileContent?: Buffer;
  checksum?: string;
  uploadedBy?: string;
  createdAt?: string;
  error?: string;
  errorDescription?: string;
}

export interface DeleteSecretResult {
  success?: boolean;
  deletedVersions?: number;
  error?: string;
  errorDescription?: string;
}

@Injectable()
export class SecretsService {
  constructor(
    private readonly secretOperationClient: SecretOperationClientService,
    private readonly authService: AuthService,
  ) {}

  private generateRequestId(): string {
    const timestamp = Date.now();
    const randomBytes = Math.random().toString(16).substring(2, 10);
    return `${timestamp}-${randomBytes}`;
  }

  async uploadSecret(
    jwt: string,
    ownerLogin: string,
    repoName: string,
    tag: string,
    envFileContent: Buffer,
  ): Promise<UploadSecretResult> {
    try {
      const authTokenResponse = await this.authService.getAuthToken(jwt);
      const userLoginResponse = await this.authService.getUserLogin(jwt);

      if (authTokenResponse.error) {
        return {
          error: authTokenResponse.error,
          errorDescription: authTokenResponse.errorDescription,
        };
      }

      if (!authTokenResponse.accessToken) {
        return {
          error: 'no_access_token',
          errorDescription: 'No access token received from auth service',
        };
      }

      if(!userLoginResponse.userLogin) {
        return {
          error: 'no_user_login',
          errorDescription: 'No user login received from auth service',
        };
      }

      const response = await this.secretOperationClient.uploadSecret({
        accessToken: authTokenResponse.accessToken,
        ownerLogin,
        repoName,
        tag,
        envFileContent,
        userLogin: userLoginResponse.userLogin,
      });

      if (response.success) {
        return {
          success: true,
          version: response.version,
          checksum: response.checksum,
        };
      } else {
        return {
          error: 'upload_failed',
          errorDescription: response.error || 'Failed to upload secret',
        };
      }
    } catch (error) {
      return {
        error: 'upload_error',
        errorDescription: error.message || 'Internal server error during upload',
      };
    }
  }

  async listSecretVersions(
    jwt: string,
    ownerLogin: string,
    repoName: string,
  ): Promise<ListSecretVersionsResult> {
    try {
      const authTokenResponse = await this.authService.getAuthToken(jwt);
      const userLoginResponse = await this.authService.getUserLogin(jwt);

      if (authTokenResponse.error) {
        return {
          error: authTokenResponse.error,
          errorDescription: authTokenResponse.errorDescription,
        };
      }

      if (!authTokenResponse.accessToken) {
        return {
          error: 'no_access_token',
          errorDescription: 'No access token received from auth service',
        };
      }

      if(!userLoginResponse.userLogin) {
        return {
          error: 'no_user_login',
          errorDescription: 'No user login received from auth service',
        };
      }

      const response = await this.secretOperationClient.listSecretVersions({
        accessToken: authTokenResponse.accessToken,
        ownerLogin,
        repoName,
        userLogin: userLoginResponse.userLogin,
      });

      if (response.versions) {
        return {
          versions: response.versions,
        };
      } else {
        return {
          error: 'list_versions_failed',
          errorDescription: response.error || 'Failed to list secret versions',
        };
      }
    } catch (error) {
      return {
        error: 'list_versions_error',
        errorDescription: error.message || 'Internal server error during list versions',
      };
    }
  }

  async downloadSecret(
    jwt: string,
    ownerLogin: string,
    repoName: string,
    version: number,
  ): Promise<DownloadSecretResult> {
    try {
      const authTokenResponse = await this.authService.getAuthToken(jwt);
      const userLoginResponse = await this.authService.getUserLogin(jwt);

      if (authTokenResponse.error) {
        return {
          error: authTokenResponse.error,
          errorDescription: authTokenResponse.errorDescription,
        };
      }

      if (!authTokenResponse.accessToken) {
        return {
          error: 'no_access_token',
          errorDescription: 'No access token received from auth service',
        };
      }

      if(!userLoginResponse.userLogin) {
        return {
          error: 'no_user_login',
          errorDescription: 'No user login received from auth service',
        };
      }

      const response = await this.secretOperationClient.downloadSecret({
        accessToken: authTokenResponse.accessToken,
        ownerLogin,
        repoName,
        version,
        userLogin: userLoginResponse.userLogin,
      });

      if (response.success) {
        return {
          success: true,
          version: response.version,
          tag: response.tag,
          envFileContent: response.envFileContent,
          checksum: response.checksum,
          uploadedBy: response.uploadedBy,
          createdAt: response.createdAt,
        };
      } else {
        return {
          error: 'download_failed',
          errorDescription: response.error || 'Failed to download secret',
        };
      }
    } catch (error) {
      return {
        error: 'download_error',
        errorDescription: error.message || 'Internal server error during download',
      };
    }
  }

  async downloadSecretByTag(
    jwt: string,
    ownerLogin: string,
    repoName: string,
    tag: string,
  ): Promise<DownloadSecretResult> {
    try {
      const authTokenResponse = await this.authService.getAuthToken(jwt);
      const userLoginResponse = await this.authService.getUserLogin(jwt);

      if (authTokenResponse.error) {
        return {
          error: authTokenResponse.error,
          errorDescription: authTokenResponse.errorDescription,
        };
      }

      if (!authTokenResponse.accessToken) {
        return {
          error: 'no_access_token',
          errorDescription: 'No access token received from auth service',
        };
      }

      if(!userLoginResponse.userLogin) {
        return {
          error: 'no_user_login',
          errorDescription: 'No user login received from auth service',
        };
      }

      const response = await this.secretOperationClient.downloadSecretByTag({
        accessToken: authTokenResponse.accessToken,
        ownerLogin,
        repoName,
        tag,
        userLogin: userLoginResponse.userLogin,
      });

      if (response.success) {
        return {
          success: true,
          version: response.version,
          tag: response.tag,
          envFileContent: response.envFileContent,
          checksum: response.checksum,
          uploadedBy: response.uploadedBy,
          createdAt: response.createdAt,
        };
      } else {
        return {
          error: 'download_failed',
          errorDescription: response.error || 'Failed to download secret',
        };
      }
    } catch (error) {
      return {
        error: 'download_error',
        errorDescription: error.message || 'Internal server error during download',
      };
    }
  }

  async deleteSecret(
    jwt: string,
    ownerLogin: string,
    repoName: string,
    version: number,
  ): Promise<DeleteSecretResult> {
    try {
      const authTokenResponse = await this.authService.getAuthToken(jwt);
      const userLoginResponse = await this.authService.getUserLogin(jwt);

      if (authTokenResponse.error) {
        return {
          error: authTokenResponse.error,
          errorDescription: authTokenResponse.errorDescription,
        };
      }

      if (!authTokenResponse.accessToken) {
        return {
          error: 'no_access_token',
          errorDescription: 'No access token received from auth service',
        };
      }

      if(!userLoginResponse.userLogin) {
        return {
          error: 'no_user_login',
          errorDescription: 'No user login received from auth service',
        };
      }

      const response = await this.secretOperationClient.deleteSecret({
        accessToken: authTokenResponse.accessToken,
        ownerLogin,
        repoName,
        version,
        userLogin: userLoginResponse.userLogin,
      });

      if (response.success) {
        return {
          success: true,
          deletedVersions: response.deletedVersions,
        };
      } else {
        return {
          error: 'delete_failed',
          errorDescription: response.error || 'Failed to delete secret',
        };
      }
    } catch (error) {
      return {
        error: 'delete_error',
        errorDescription: error.message || 'Internal server error during delete',
      };
    }
  }
} 