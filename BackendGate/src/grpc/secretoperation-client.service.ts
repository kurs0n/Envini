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

interface SecretsService {
  listRepos(request: { accessToken: string }): any;
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
} 