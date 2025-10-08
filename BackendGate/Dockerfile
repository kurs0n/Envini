# syntax=docker/dockerfile:1

# Build stage
FROM --platform=$BUILDPLATFORM node:20-alpine AS builder
WORKDIR /app

COPY BackendGate/package*.json ./
RUN npm ci --omit=dev=false

COPY BackendGate .
RUN npm run build

# Production stage
FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production

# Copy only needed files
COPY --from=builder /app/package*.json ./
RUN npm ci --omit=dev
COPY --from=builder /app/dist ./dist
# Copy repo root protos to match '../proto/*.proto' resolution from /app
COPY proto /proto

ENV PORT=3000
EXPOSE 3000
CMD ["node", "dist/main.js"]
