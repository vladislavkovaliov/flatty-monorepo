FROM node:22-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json nx.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN npm ci

RUN cd apps/react-launcher && npx vite build --config vite.launcher.config.ts

FROM nginx:alpine
COPY --from=builder /build/apps/react-launcher/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
