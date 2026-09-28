FROM oven/bun:1.3.5-alpine AS builder
WORKDIR /app
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_PAYWALL_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL NEXT_PUBLIC_PAYWALL_URL=$NEXT_PUBLIC_PAYWALL_URL NEXT_TELEMETRY_DISABLED=1
COPY package.json bun.lockb ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

# Ship only the standalone application, with no source tree, credentials or build tools.
FROM oven/bun:1.3.5-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 HOSTNAME=0.0.0.0 PORT=3000
COPY --from=builder --chown=bun:bun /app/.next/standalone ./
COPY --from=builder --chown=bun:bun /app/.next/static ./.next/static
COPY --from=builder --chown=bun:bun /app/public ./public
USER bun
EXPOSE 3000
CMD ["bun", "server.js"]
