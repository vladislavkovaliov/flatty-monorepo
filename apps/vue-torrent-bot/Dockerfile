FROM node:24-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json nx.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN npm ci

RUN cd apps/vue-torrent-bot && npx vite build

FROM nginx:alpine
COPY --from=builder /build/apps/vue-torrent-bot/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
