FROM node:22-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json nx.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN npm ci

RUN cd apps/react-settings && npx webpack --config webpack.apps.config.mjs --mode production

FROM nginx:alpine
COPY --from=builder /build/apps/react-settings/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
