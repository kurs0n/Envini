import { Controller, Get, Headers } from '@nestjs/common';
import { ReposService, ListReposResult } from './repos.service';

@Controller('repos')
export class ReposController {
  constructor(private readonly reposService: ReposService) {}

  @Get('list')
  async listRepos(@Headers('authorization') authHeader: string): Promise<ListReposResult> {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return {
        error: 'invalid_authorization',
        errorDescription: 'Authorization header must be in format: Bearer <JWT>',
      };
    }
    
    const jwt = authHeader.substring(7);
    return await this.reposService.listRepos(jwt);
  }
}
