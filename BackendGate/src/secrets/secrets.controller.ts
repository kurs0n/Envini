import {
  Controller,
  Post,
  Get,
  Delete,
  Headers,
  Body,
  Param,
  Query,
  Res,
  HttpStatus,
  BadRequestException,
} from '@nestjs/common';
import { Response } from 'express';
import { SecretsService, UploadSecretResult, ListSecretVersionsResult, DownloadSecretResult, DeleteSecretResult } from './secrets.service';

@Controller('secrets')
export class SecretsController {
  constructor(private readonly secretsService: SecretsService) {}

  @Post('upload/:ownerLogin/:repoName')
  async uploadSecret(
    @Headers('authorization') authHeader: string,
    @Param('ownerLogin') ownerLogin: string,
    @Param('repoName') repoName: string,
    @Body() body: { tag?: string; envFileContent: string },
  ): Promise<UploadSecretResult> {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new BadRequestException('Authorization header must be in format: Bearer <JWT>');
    }

    if (!body.envFileContent) {
      throw new BadRequestException('envFileContent is required');
    }

    const jwt = authHeader.substring(7);
    const envFileBuffer = Buffer.from(body.envFileContent, 'base64');

    return await this.secretsService.uploadSecret(
      jwt,
      ownerLogin,
      repoName,
      body.tag || '',
      envFileBuffer,
    );
  }

  @Get('versions/:ownerLogin/:repoName')
  async listSecretVersions(
    @Headers('authorization') authHeader: string,
    @Param('ownerLogin') ownerLogin: string,
    @Param('repoName') repoName: string,
  ): Promise<ListSecretVersionsResult> {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new BadRequestException('Authorization header must be in format: Bearer <JWT>');
    }

    const jwt = authHeader.substring(7);

    return await this.secretsService.listSecretVersions(jwt, ownerLogin, repoName);
  }

  @Get('download/:ownerLogin/:repoName')
  async downloadSecret(
    @Headers('authorization') authHeader: string,
    @Param('ownerLogin') ownerLogin: string,
    @Param('repoName') repoName: string,
    @Query('version') version: string,
    @Query('tag') tag: string,
    @Res() res: Response,
  ): Promise<void> {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new BadRequestException('Authorization header must be in format: Bearer <JWT>');
    }

    const jwt = authHeader.substring(7);

    let result: DownloadSecretResult;

    if (tag) {
      // Download by tag
      result = await this.secretsService.downloadSecretByTag(
        jwt,
        ownerLogin,
        repoName,
        tag,
      );
    } else {
      // Download by version
      const versionNumber = version ? parseInt(version, 10) : 0;

      if (version && isNaN(versionNumber)) {
        throw new BadRequestException('Version must be a valid number');
      }

      result = await this.secretsService.downloadSecret(
        jwt,
        ownerLogin,
        repoName,
        versionNumber,
      );
    }

    if (result.success && result.envFileContent && result.version !== undefined) {
      // Set response headers for file download
      const filename = tag 
        ? `${ownerLogin}-${repoName}-${tag}-v${result.version}.env`
        : `${ownerLogin}-${repoName}-v${result.version}.env`;
      
      res.setHeader('Content-Type', 'application/octet-stream');
      res.setHeader('Content-Disposition', `attachment; filename="${filename}"`);
      res.setHeader('X-Secret-Version', result.version.toString());
      res.setHeader('X-Secret-Tag', result.tag || '');
      res.setHeader('X-Secret-Checksum', result.checksum || '');
      res.setHeader('X-Secret-UploadedBy', result.uploadedBy || '');
      res.setHeader('X-Secret-CreatedAt', result.createdAt || '');

      res.status(HttpStatus.OK).send(result.envFileContent);
    } else {
      res.status(HttpStatus.BAD_REQUEST).json({
        error: result.error || 'download_failed',
        errorDescription: result.errorDescription || 'Failed to download secret',
      });
    }
  }

  @Delete('delete/:ownerLogin/:repoName')
  async deleteSecret(
    @Headers('authorization') authHeader: string,
    @Param('ownerLogin') ownerLogin: string,
    @Param('repoName') repoName: string,
    @Query('version') version: string,
    @Query('all') deleteAll: string,
  ): Promise<DeleteSecretResult> {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new BadRequestException('Authorization header must be in format: Bearer <JWT>');
    }

    const jwt = authHeader.substring(7);
    
    // If 'all' parameter is present, delete all versions (version = 0)
    // Otherwise, use the specified version
    let versionNumber = 0;
    if (deleteAll !== 'true') {
      if (!version) {
        throw new BadRequestException('Version parameter is required unless deleting all versions');
      }
      versionNumber = parseInt(version, 10);
      if (isNaN(versionNumber)) {
        throw new BadRequestException('Version must be a valid number');
      }
    }

    return await this.secretsService.deleteSecret(
      jwt,
      ownerLogin,
      repoName,
      versionNumber,
    );
  }

  @Get('content/:ownerLogin/:repoName')
  async getSecretContent(
    @Headers('authorization') authHeader: string,
    @Param('ownerLogin') ownerLogin: string,
    @Param('repoName') repoName: string,
    @Query('version') version: string,
    @Query('tag') tag: string,
  ): Promise<DownloadSecretResult> {
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new BadRequestException('Authorization header must be in format: Bearer <JWT>');
    }

    const jwt = authHeader.substring(7);

    if (tag) {
      // Get content by tag
      return await this.secretsService.downloadSecretByTag(
        jwt,
        ownerLogin,
        repoName,
        tag,
      );
    } else {
      // Get content by version
      const versionNumber = version ? parseInt(version, 10) : 0;

      if (version && isNaN(versionNumber)) {
        throw new BadRequestException('Version must be a valid number');
      }

      return await this.secretsService.downloadSecret(
        jwt,
        ownerLogin,
        repoName,
        versionNumber,
      );
    }
  }
} 