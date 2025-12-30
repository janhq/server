# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy all files
COPY . .

# Install dependencies and build
RUN npm install
ARG NEXT_PUBLIC_JAN_BASE_URL
ENV NEXT_PUBLIC_JAN_BASE_URL=$NEXT_PUBLIC_JAN_BASE_URL
RUN npm run build

# Production stage
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Create non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy built application and node_modules from builder with ownership
COPY --from=builder --chown=nextjs:nodejs /app/.next ./.next
COPY --from=builder --chown=nextjs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nextjs:nodejs /app/package.json ./package.json
COPY --from=builder --chown=nextjs:nodejs /app/next.config.mjs ./next.config.mjs
COPY --from=builder --chown=nextjs:nodejs /app/source.config.ts ./source.config.ts
COPY --from=builder --chown=nextjs:nodejs /app/tsconfig.json ./tsconfig.json
COPY --from=builder --chown=nextjs:nodejs /app/content ./content
COPY --from=builder --chown=nextjs:nodejs /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/src ./src

USER nextjs

EXPOSE 3000

CMD ["node_modules/.bin/next", "start"]