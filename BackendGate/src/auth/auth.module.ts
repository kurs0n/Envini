import { Module } from '@nestjs/common';
import { AuthController } from './auth.controller';
import { AuthService } from './auth.service';
import { AuthClientService } from '../grpc/auth-client.service';

@Module({
  controllers: [AuthController],
  providers: [AuthService, AuthClientService],
  exports: [AuthService],
})
export class AuthModule {} 