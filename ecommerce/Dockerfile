FROM node:24-alpine AS deps
WORKDIR /app
COPY ecommerce/package.json ecommerce/package-lock.json* ./
RUN npm install

FROM node:24-alpine AS build
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:4005
ARG NEXT_PUBLIC_STORE_SLUG=rangavali
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
ENV NEXT_PUBLIC_STORE_SLUG=$NEXT_PUBLIC_STORE_SLUG
COPY --from=deps /app/node_modules ./node_modules
COPY ecommerce ./
RUN npm run build

FROM node:24-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
COPY --from=build /app ./
EXPOSE 3000
CMD ["npm", "run", "start", "--", "-p", "3000"]
