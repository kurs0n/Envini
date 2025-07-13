import { Module } from '@nestjs/common';
import { AuthModule } from './auth/auth.module';
import { ReposModule } from './repos/repos.module';

@Module({
  imports: [AuthModule, ReposModule],
})
export class AppModule {}
