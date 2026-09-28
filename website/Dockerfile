FROM node:24-alpine AS build
WORKDIR /app
RUN npm install --global bun@1.3.5
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
ARG XEM_BUILD_SITE_URL=https://xem.email
ARG XEM_BUILD_APP_URL=https://app.xem.email
ENV XEM_BUILD_SITE_URL=$XEM_BUILD_SITE_URL
ENV XEM_BUILD_APP_URL=$XEM_BUILD_APP_URL
RUN bun run build

FROM nginxinc/nginx-unprivileged:stable-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/out /usr/share/nginx/html
EXPOSE 8080
