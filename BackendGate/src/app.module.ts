import { Module } from '@nestjs/common';
import { AuthModule } from './auth/auth.module';
import { ReposModule } from './repos/repos.module';
import { ConfigModule } from '@nestjs/config';

@Module({
  imports: [AuthModule, ReposModule, ConfigModule.forRoot({
    isGlobal: true
  })],
})
export class AppModule {}
