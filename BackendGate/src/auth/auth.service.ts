import { Injectable } from '@nestjs/common';
import { AuthClientService } from '../grpc/auth-client.service';

@Injectable()
export class AuthService {
  constructor(private readonly authClientService: AuthClientService) {}

  async startGitHubAuth() {
    try {
      const response = await this.authClientService.startDeviceFlow();
      return {
        verificationUri: response.verificationUri,
        userCode: response.userCode,
        deviceCode: response.deviceCode,
        expiresIn: response.expiresIn,
        interval: response.interval,
      };
    } catch (error) {
      throw new Error(`Failed to start GitHub auth: ${error.message}`);
    }
  }

  async pollForToken(deviceCode: string) {
    try {
      const response = await this.authClientService.pollForToken(deviceCode);
      
      if (response.error) {
        return {
          error: response.error,
          errorDescription: response.errorDescription,
        };
      }

      return {
        sessionId: response.sessionId,
        jwt: response.jwt,
      };
    } catch (error) {
      throw new Error(`Failed to poll for token: ${error.message}`);
    }
  }

  async getAuthToken(jwt: string) {
    try {
      const response = await this.authClientService.getAuthToken(jwt);
      
      if (response.error) {
        return {
          error: response.error,
          errorDescription: response.errorDescription,
        };
      }

      return {
        accessToken: response.accessToken,
        tokenType: response.tokenType,
        scope: response.scope,
      };
    } catch (error) {
      throw new Error(`Failed to get auth token: ${error.message}`);
    }
  }

  async validateSession(jwt: string) {
    try {
      const response = await this.authClientService.validateSession(jwt);
      return {
        valid: response.valid,
        error: response.error,
      };
    } catch (error) {
      throw new Error(`Failed to validate session: ${error.message}`);
    }
  }

  async logout(jwt: string) {
    try {
      const response = await this.authClientService.logout(jwt);
      return {
        success: response.success,
        error: response.error,
      };
    } catch (error) {
      throw new Error(`Failed to logout: ${error.message}`);
    }
  }

  async getUserLogin(jwt: string) {
    try {
      console.log('DEBUG: Calling AuthService getUserLogin with JWT:', jwt.substring(0, 20) + '...');
      const response = await this.authClientService.getUserLogin(jwt);
      console.log('DEBUG: AuthService getUserLogin response:', response);
      return {
        userLogin: response.userLogin,
        error: response.error,
        errorDescription: response.errorDescription,
      };
    } catch (error) {
      console.error('DEBUG: AuthService getUserLogin error:', error);
      throw new Error(`Failed to get user login: ${error.message}`);
    }
  }
}