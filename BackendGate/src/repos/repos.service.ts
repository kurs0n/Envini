import { Injectable } from '@nestjs/common';
import { AuthService } from '../auth/auth.service';
import { SecretOperationClientService } from '../grpc/secretoperation-client.service';

interface Repo {
  id: number;
  name: string;
  fullName: string;
  htmlUrl: string;
  description: string;
  private: boolean;
  ownerLogin: string;
  ownerAvatarUrl: string;
}

export interface ListReposResult {
  repos?: Repo[];
  error?: string;
  errorDescription?: string;
}

@Injectable()
export class ReposService {
  constructor(
    private readonly authService: AuthService,
    private readonly secretOperationClientService: SecretOperationClientService,
  ) {}

  async listRepos(jwt: string): Promise<ListReposResult> {
    try {
      const authTokenResponse = await this.authService.getAuthToken(jwt);
      
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

      const reposResponse = await this.secretOperationClientService.listRepos(
        authTokenResponse.accessToken
      );

      if (reposResponse.error) {
        return {
          error: reposResponse.error,
        };
      }

      return {
        repos: reposResponse.repos,
      };
    } catch (error) {
      throw new Error(`Failed to list repositories: ${error.message}`);
    }
  }
} 