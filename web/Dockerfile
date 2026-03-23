FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps

COPY . .

# Принудительно даём права на выполнение всем скриптам в node_modules
RUN find node_modules/.bin -type f -exec chmod +x {} \;

# Запускаем сборку через node, а не через npx
RUN node node_modules/react-scripts/bin/react-scripts.js build

FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
EXPOSE 3000
CMD ["nginx", "-g", "daemon off;"]