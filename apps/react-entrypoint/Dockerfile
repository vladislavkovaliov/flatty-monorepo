FROM node:24-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json nx.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN npm ci

RUN cd apps/react-entrypoint && npx next build

FROM node:24-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /build/apps/react-entrypoint/.next/standalone ./
COPY --from=builder /build/apps/react-entrypoint/.next/static ./apps/react-entrypoint/.next/static
COPY --from=builder /build/apps/react-entrypoint/public ./apps/react-entrypoint/public

EXPOSE 3000

CMD ["node", "apps/react-entrypoint/server.js"]
