import { Controller, Get, Headers, HttpException, HttpStatus } from '@nestjs/common';
import { ReposService } from './repos.service';

@Controller('repos')
export class ReposController {
  constructor(private readonly reposService: ReposService) {}

  @Get('list')
  async listRepos(@Headers('authorization') authHeader: string) {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new HttpException('Authorization header required', HttpStatus.UNAUTHORIZED);
    }

    const jwt = authHeader.substring(7);
    return await this.reposService.listRepos(jwt);
  }

  @Get('list-with-versions')
  async listReposWithVersions(@Headers('authorization') authHeader: string) {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new HttpException('Authorization header required', HttpStatus.UNAUTHORIZED);
    }

    const jwt = authHeader.substring(7);
    return await this.reposService.listReposWithVersions(jwt);
  }
}
