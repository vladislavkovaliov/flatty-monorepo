FROM node:24-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json nx.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN test -f packages/sdk/src/types/api.d.ts

RUN npm ci

# Rewrites are evaluated during `next build`, so the destinations must be
# supplied as build args (runtime env cannot change them afterwards).
ARG GO_API_ORIGIN=http://localhost:8080
ARG NEST_ORIGIN=http://localhost:3000
ARG SETTINGS_ORIGIN=http://localhost:8081
ARG RESIDENT_ORIGIN=http://localhost:8082
ARG VUE_TORRENT_BOT_ORIGIN=http://localhost:8084
ENV GO_API_ORIGIN=$GO_API_ORIGIN \
    NEST_ORIGIN=$NEST_ORIGIN \
    SETTINGS_ORIGIN=$SETTINGS_ORIGIN \
    RESIDENT_ORIGIN=$RESIDENT_ORIGIN \
    VUE_TORRENT_BOT_ORIGIN=$VUE_TORRENT_BOT_ORIGIN

RUN cd apps/react-entrypoint && npx next build

FROM node:24-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /build/apps/react-entrypoint/.next/standalone ./
COPY --from=builder /build/apps/react-entrypoint/.next/static ./apps/react-entrypoint/.next/static
COPY --from=builder /build/apps/react-entrypoint/public ./apps/react-entrypoint/public

EXPOSE 3000

CMD ["node", "apps/react-entrypoint/server.js"]
