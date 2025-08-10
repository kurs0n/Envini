import { Controller, Post, Get, Body, Query, Headers, HttpException, HttpStatus } from '@nestjs/common';
import { AuthService } from './auth.service';

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Post('github/start')
  async startGitHubAuth() {
    return await this.authService.startGitHubAuth();
  }

  @Get('github/poll')
  async pollForToken(@Query('deviceCode') deviceCode: string) {
    return await this.authService.pollForToken(deviceCode);
  }

  @Get('validate')
  async validateSession(@Query('jwt') jwt: string) {
    return await this.authService.validateSession(jwt);
  }

  @Post('logout')
  async logout(@Body('jwt') jwt: string) {
    return await this.authService.logout(jwt);
  }

  @Get("user")
  async getUserLogin(@Headers('authorization') authHeader: string) {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
          throw new HttpException('Authorization header required', HttpStatus.UNAUTHORIZED);
    }
    
    const jwt = authHeader.substring(7);
    const response = await this.authService.getUserLogin(jwt);
    if (response.error) {
      return {
        error: response.error,
        errorDescription: response.errorDescription,
      };
    }
    return {
      userLogin: response.userLogin,
      avatarUrl: response.avatarUrl,
      htmlUrl: response.htmlUrl,
      name: response.name,
    };
  }
} 