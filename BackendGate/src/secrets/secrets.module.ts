import { Module } from '@nestjs/common';
import { SecretsController } from './secrets.controller';
import { SecretsService } from './secrets.service';
import { SecretOperationClientService } from '../grpc/secretoperation-client.service';
import { AuthModule } from '../auth/auth.module';

@Module({
  imports: [AuthModule],
  controllers: [SecretsController],
  providers: [SecretsService, SecretOperationClientService],
})
export class SecretsModule {} 