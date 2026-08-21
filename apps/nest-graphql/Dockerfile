FROM node:24-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN npm ci

RUN cd apps/nest-graphql && npm run build

FROM node:24-alpine
RUN apk add --no-cache curl
WORKDIR /app
COPY --from=builder /build/apps/nest-graphql/dist ./dist
COPY --from=builder /build/node_modules ./node_modules
COPY --from=builder /build/apps/nest-graphql/node_modules ./node_modules
ENV NODE_OPTIONS="--require ./dist/instrumentation.js"
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD curl -f http://localhost:3000/api/config || exit 1
CMD ["node", "dist/main"]
