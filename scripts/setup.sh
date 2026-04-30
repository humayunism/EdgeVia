#!/bin/bash
echo "Setting up Edgevia development environment..."
cp .env.example .env
cd infra && docker-compose up -d
echo "Waiting for PostgreSQL..."
sleep 3
cd ../apps/api && npm install -g pnpm && pnpm install && pnpm run db:migrate && pnpm run db:generate
echo ""
echo "Setup complete! Now run:"
echo "  Terminal 1: cd apps/proxy && go run cmd/server/main.go"
echo "  Terminal 2: cd apps/api && pnpm run dev"
echo "  Terminal 3: cd apps/dashboard && pnpm run dev"
