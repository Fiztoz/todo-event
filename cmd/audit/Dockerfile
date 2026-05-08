FROM oven/bun:1-alpine
WORKDIR /app
COPY cmd/audit/package.json cmd/audit/bun.lock* ./
RUN bun install --frozen-lockfile || bun install
COPY cmd/audit/main.ts ./
USER bun
ENTRYPOINT ["bun", "run", "main.ts"]
