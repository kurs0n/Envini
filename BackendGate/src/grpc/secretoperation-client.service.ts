import { Injectable, OnModuleInit } from '@nestjs/common';
import { Client, ClientGrpc, Transport } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';

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

interface ListReposResponse {
  repos?: Repo[];
  error?: string;
}

interface GrpcRepo {
  id: any;
  name: string;
  fullName: string;
  htmlUrl: string;
  description: string;
  private: boolean;
  ownerLogin: string;
  ownerAvatarUrl: string;
}

interface GrpcListReposResponse {
  repos?: GrpcRepo[];
  error?: string;
}

interface SecretVersion {
  version: number;
  tag: string;
  checksum: string;
  uploadedBy: string;
  createdAt: string;
}

interface UploadSecretRequest {
  accessToken: string;
  ownerLogin: string;
  repoName: string;
  tag: string;
  envFileContent: Buffer;
}

interface UploadSecretResponse {
  success: boolean;
  version: number;
  checksum: string;
  error: string;
}

interface ListSecretVersionsRequest {
  accessToken: string;
  ownerLogin: string;
  repoName: string;
}

interface ListSecretVersionsResponse {
  versions: SecretVersion[];
  error: string;
}

interface DownloadSecretRequest {
  accessToken: string;
  ownerLogin: string;
  repoName: string;
  version: number;
}

interface DownloadSecretByTagRequest {
  accessToken: string;
  ownerLogin: string;
  repoName: string;
  tag: string;
}

interface DownloadSecretResponse {
  success: boolean;
  version: number;
  tag: string;
  envFileContent: Buffer;
  checksum: string;
  uploadedBy: string;
  createdAt: string;
  error: string;
  isEncrypted: boolean;
}

interface DeleteSecretRequest {
  accessToken: string;
  ownerLogin: string;
  repoName: string;
  version: number;
}

interface DeleteSecretResponse {
  success: boolean;
  deletedVersions: number;
  error: string;
}

interface SecretsService {
  listRepos(request: { accessToken: string }): any;
  uploadSecret(request: UploadSecretRequest): any;
  listSecretVersions(request: ListSecretVersionsRequest): any;
  downloadSecret(request: DownloadSecretRequest): any;
  downloadSecretByTag(request: DownloadSecretByTagRequest): any;
  deleteSecret(request: DeleteSecretRequest): any;
  listAllRepositoriesWithVersions(request: { accessToken: string }): any;
}

@Injectable()
export class SecretOperationClientService implements OnModuleInit {
  @Client({
    transport: Transport.GRPC,
    options: {
      url: "localhost:50053",
      package: 'secretsservice',
      protoPath: '../proto/secrets.proto',
    },
  })
  private client: ClientGrpc;

  private secretsService: SecretsService;

  onModuleInit() {
    this.secretsService = this.client.getService<SecretsService>('SecretsService');
  }

  async listRepos(accessToken: string): Promise<ListReposResponse> {
    const response = await firstValueFrom(this.secretsService.listRepos({ accessToken })) as GrpcListReposResponse;
    
    if (response.repos) {
      response.repos = response.repos.map(repo => ({
        ...repo,
        id: typeof repo.id === 'object' && repo.id !== null ? repo.id.low : repo.id,
      }));
    }
    
    return response as ListReposResponse;
  }

  async uploadSecret(request: UploadSecretRequest): Promise<UploadSecretResponse> {
    const response = await firstValueFrom(this.secretsService.uploadSecret(request));
    return response as UploadSecretResponse;
  }

  async listSecretVersions(request: ListSecretVersionsRequest): Promise<ListSecretVersionsResponse> {
    const response = await firstValueFrom(this.secretsService.listSecretVersions(request));
    return response as ListSecretVersionsResponse;
  }

  async downloadSecret(request: DownloadSecretRequest): Promise<DownloadSecretResponse> {
    const response = await firstValueFrom(this.secretsService.downloadSecret(request));
    return response as DownloadSecretResponse;
  }

  async downloadSecretByTag(request: DownloadSecretByTagRequest): Promise<DownloadSecretResponse> {
    const response = await firstValueFrom(this.secretsService.downloadSecretByTag(request));
    return response as DownloadSecretResponse;
  }

  async deleteSecret(request: DeleteSecretRequest): Promise<DeleteSecretResponse> {
    const response = await firstValueFrom(this.secretsService.deleteSecret(request));
    return response as DeleteSecretResponse;
  }

  async listAllRepositoriesWithVersions(accessToken: string): Promise<any> {
    const response = await firstValueFrom(this.secretsService.listAllRepositoriesWithVersions({ accessToken })) as any;
    
    // Convert Long objects to regular numbers
    if (response.repositories) {
      response.repositories = response.repositories.map((repo: any) => ({
        ...repo,
        repoId: typeof repo.repoId === 'object' && repo.repoId !== null ? repo.repoId.low : repo.repoId,
      }));
    }
    
    return response;
  }
} 