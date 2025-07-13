import { Module } from '@nestjs/common';
import { ReposController } from './repos.controller';
import { ReposService } from './repos.service';
import { SecretOperationClientService } from '../grpc/secretoperation-client.service';
import { AuthModule } from '../auth/auth.module';

@Module({
  imports: [AuthModule],
  controllers: [ReposController],
  providers: [ReposService, SecretOperationClientService],
})
export class ReposModule {} 