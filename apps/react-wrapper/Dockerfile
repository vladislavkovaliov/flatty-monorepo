FROM node:22-alpine AS builder
WORKDIR /build

COPY package.json package-lock.json nx.json ./
COPY apps/ apps/
COPY packages/ packages/

RUN npm ci

RUN echo "PUBLIC_PROTOCOL=http" > apps/react-wrapper/.env && \
    echo "PUBLIC_URL=localhost" >> apps/react-wrapper/.env && \
    echo "PUBLIC_PORT=80" >> apps/react-wrapper/.env && \
    echo "PUBLIC_PATH=api/swagger/doc.json" >> apps/react-wrapper/.env

RUN cd apps/react-wrapper && npx vite build

FROM nginx:alpine
RUN echo 'server { listen 80; root /usr/share/nginx/html; index index.html; location / { try_files $uri $uri/ /index.html; } }' > /etc/nginx/conf.d/default.conf
COPY --from=builder /build/apps/react-wrapper/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
