FROM node:22-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json nx.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN npm ci

RUN echo "PUBLIC_PROTOCOL=http" > apps/react-resident/.env && \
    echo "PUBLIC_URL=localhost" >> apps/react-resident/.env && \
    echo "PUBLIC_PORT=80" >> apps/react-resident/.env && \
    echo "PUBLIC_PATH=api/swagger/doc.json" >> apps/react-resident/.env

RUN cd apps/react-resident && npx webpack --config webpack.apps.config.mjs --mode production

FROM nginx:alpine
COPY --from=builder /build/apps/react-resident/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
