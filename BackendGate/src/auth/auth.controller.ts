import { Controller, Post, Get, Body, Query } from '@nestjs/common';
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
} 