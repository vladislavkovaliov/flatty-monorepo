FROM node:24-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json nx.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN npm ci

ARG PUBLIC_PROTOCOL=http
ARG PUBLIC_URL=localhost
ARG PUBLIC_PORT=80
ARG PUBLIC_PATH=api/swagger/doc.json
RUN printf 'PUBLIC_PROTOCOL=%s\nPUBLIC_URL=%s\nPUBLIC_PORT=%s\nPUBLIC_PATH=%s\n' \
    "$PUBLIC_PROTOCOL" "$PUBLIC_URL" "$PUBLIC_PORT" "$PUBLIC_PATH" \
    > apps/react-resident/.env

RUN cd apps/react-resident && npx webpack --config webpack.apps.config.mjs --mode production

FROM nginx:alpine
COPY --from=builder /build/apps/react-resident/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
