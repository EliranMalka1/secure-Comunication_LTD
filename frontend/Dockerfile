# ---- build stage ----
FROM node:20-alpine AS build
WORKDIR /app

# Install dependencies with efficient caching
COPY package.json package-lock.json* pnpm-lock.yaml* yarn.lock* ./
RUN npm ci || npm install

# Application code
COPY . .
RUN npm run build

# ---- run stage (nginx) ----
FROM nginx:1.27-alpine
# Config file for SPA (client-side routes)
COPY nginx.conf /etc/nginx/conf.d/default.conf
# Vite build artifacts
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
