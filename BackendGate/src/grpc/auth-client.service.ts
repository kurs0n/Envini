import { Injectable, OnModuleInit } from '@nestjs/common';
import { Client, ClientGrpc, Transport } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';

// Response interfaces based on the proto definitions
interface StartDeviceFlowResponse {
  verificationUri: string;
  userCode: string;
  deviceCode: string;
  expiresIn: number;
  interval: number;
}

interface PollForTokenResponse {
  sessionId?: string;
  jwt?: string;
  error?: string;
  errorDescription?: string;
}

interface GetAuthTokenResponse {
  accessToken?: string;
  tokenType?: string;
  scope?: string;
  error?: string;
  errorDescription?: string;
}

interface ValidateSessionResponse {
  valid: boolean;
  error?: string;
}

interface LogoutResponse {
  success: boolean;
  error?: string;
}

interface AuthService {
  startDeviceFlow(request: { clientId?: string }): any;
  pollForToken(request: { deviceCode: string }): any;
  getAuthToken(request: { jwt: string }): any;
  validateSession(request: { jwt: string }): any;
  logout(request: { jwt: string }): any;
}

@Injectable()
export class AuthClientService implements OnModuleInit {
  @Client({
    transport: Transport.GRPC,
    options: {
      url: 'localhost:50051',
      package: 'authservice',
      protoPath: '../proto/auth.proto',
    },
  })
  private client: ClientGrpc;

  private authService: AuthService;

  onModuleInit() {
    this.authService = this.client.getService<AuthService>('AuthService');
  }

  async startDeviceFlow(): Promise<StartDeviceFlowResponse> {
    return await firstValueFrom(this.authService.startDeviceFlow({}));
  }

  async pollForToken(deviceCode: string): Promise<PollForTokenResponse> {
    return await firstValueFrom(this.authService.pollForToken({ deviceCode }));
  }

  async getAuthToken(jwt: string): Promise<GetAuthTokenResponse> {
    return await firstValueFrom(this.authService.getAuthToken({ jwt }));
  }

  async validateSession(jwt: string): Promise<ValidateSessionResponse> {
    return await firstValueFrom(this.authService.validateSession({ jwt }));
  }

  async logout(jwt: string): Promise<LogoutResponse> {
    return await firstValueFrom(this.authService.logout({ jwt }));
  }
} 