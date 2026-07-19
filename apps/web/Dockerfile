FROM node:24-alpine AS deps
WORKDIR /app
COPY apps/web/package.json ./
RUN npm install

FROM node:24-alpine AS build
WORKDIR /app
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:4005
ARG NEXT_PUBLIC_ADMIN_BASE_URL=http://localhost:3005
ARG NEXT_PUBLIC_STOREFRONT_BASE_URL=http://localhost:3006
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV NEXT_PUBLIC_ADMIN_BASE_URL=$NEXT_PUBLIC_ADMIN_BASE_URL
ENV NEXT_PUBLIC_STOREFRONT_BASE_URL=$NEXT_PUBLIC_STOREFRONT_BASE_URL
COPY --from=deps /app/node_modules ./node_modules
COPY apps/web ./
RUN npm run build

FROM node:24-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:4005
ARG NEXT_PUBLIC_ADMIN_BASE_URL=http://localhost:3005
ARG NEXT_PUBLIC_STOREFRONT_BASE_URL=http://localhost:3006
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV NEXT_PUBLIC_ADMIN_BASE_URL=$NEXT_PUBLIC_ADMIN_BASE_URL
ENV NEXT_PUBLIC_STOREFRONT_BASE_URL=$NEXT_PUBLIC_STOREFRONT_BASE_URL
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
COPY --from=build /app/public ./public
EXPOSE 3005
CMD ["node", "server.js"]
