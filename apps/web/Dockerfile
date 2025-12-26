# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm install

# Copy source files
COPY . .

# Build arguments for environment variables
ARG JAN_API_BASE_URL
ARG VITE_AUTH_URL
ARG VITE_AUTH_REALM
ARG VITE_AUTH_CLIENT_ID
ARG VITE_OAUTH_REDIRECT_URI
ARG VITE_GA_ID

# Set environment variables for build
ENV JAN_API_BASE_URL=$JAN_API_BASE_URL
ENV VITE_AUTH_URL=$VITE_AUTH_URL
ENV VITE_AUTH_REALM=$VITE_AUTH_REALM
ENV VITE_AUTH_CLIENT_ID=$VITE_AUTH_CLIENT_ID
ENV VITE_OAUTH_REDIRECT_URI=$VITE_OAUTH_REDIRECT_URI
ENV VITE_GA_ID=$VITE_GA_ID

# Build application (skip TypeScript type checking for now)
RUN npx vite build

# Production stage
FROM nginx:alpine AS runner

WORKDIR /usr/share/nginx/html

# Remove default nginx static assets
RUN rm -rf ./*

# Copy built application from builder
COPY --from=builder /app/dist .

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=40s \
  CMD wget -q --spider http://localhost:3001 || exit 1

CMD ["nginx", "-g", "daemon off;"]
