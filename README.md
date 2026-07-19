# RetailPulse AI

RetailPulse AI is a multi-tenant SaaS foundation for Amazon sellers. This repository is being built incrementally by phase.

## Phase 1 Scope

- Go/Fiber backend with clean architecture boundaries.
- PostgreSQL schema migrations for authentication, organizations, RBAC, audit logs, API tokens, sync history, and future Amazon stores.
- JWT access tokens and rotating refresh tokens.
- Password-based local SaaS account auth for RetailPulse users. Amazon Seller Central credentials are never requested.
- Next.js frontend skeleton for registration, login, and dashboard entry.
- Docker Compose with PostgreSQL, Redis, MinIO, backend, and frontend.
- Kubernetes manifests for API, web, Postgres, Redis, and MinIO.
- OpenAPI documentation.
- Unit and integration test scaffolding.

## Amazon SP-API Connection

The dashboard includes a seller-safe Amazon connection flow. RetailPulse never asks for an Amazon password. Sellers are redirected to Amazon Seller Central, Amazon returns an SP-API OAuth code to the backend callback, and the backend exchanges that code for an LWA refresh token.

To use a real seller account, configure these values in `.env`:

```bash
AMAZON_LWA_CLIENT_ID=
AMAZON_LWA_CLIENT_SECRET=
AMAZON_SPAPI_APP_ID=
AMAZON_SPAPI_AUTH_VERSION=
AMAZON_LWA_REDIRECT_URL=http://localhost:4005/v1/amazon/oauth/callback
AMAZON_SELLER_CENTRAL_URL=https://sellercentral.amazon.com
AMAZON_SPAPI_ENDPOINT=https://sellingpartnerapi-na.amazon.com
AMAZON_AWS_ACCESS_KEY_ID=
AMAZON_AWS_SECRET_ACCESS_KEY=
AMAZON_AWS_SESSION_TOKEN=
AMAZON_AWS_REGION=us-east-1
```

The SP-API import endpoint currently imports recent Orders API data into PostgreSQL. The request is signed with AWS Signature Version 4 and includes the seller's refreshed LWA access token in `x-amz-access-token`.

Without real Amazon developer credentials and approved SP-API roles, the Connect Amazon button will return a configuration error.

For a draft SP-API application, set `AMAZON_SPAPI_AUTH_VERSION=beta` so Amazon receives `version=beta` in the consent URL. For public onboarding where any Amazon seller can authorize your SaaS, the SP-API application must be approved/published and `AMAZON_SPAPI_AUTH_VERSION` should be empty. `AMAZON_SPAPI_APP_ID` must be the SP-API application ID from the application details page, not a Solution Provider / solution ID.

Use `GET /v1/amazon/oauth/status` to verify whether seller onboarding is ready. The seller-facing UI must never ask sellers for Solution Provider credentials; those credentials stay only in the SaaS backend environment.

## Amazon SP-API Sandbox

Sandbox support runs alongside the production OAuth flow. Add the sandbox LWA credentials and token created in Amazon's Solution Provider Portal:

```bash
AMAZON_SANDBOX_LWA_CLIENT_ID=
AMAZON_SANDBOX_LWA_CLIENT_SECRET=
AMAZON_SANDBOX_REFRESH_TOKEN=
AMAZON_SANDBOX_SPAPI_ENDPOINT=https://sandbox.sellingpartnerapi-na.amazon.com
```

After restarting the API, sign in and choose **Connect sandbox**. Sandbox stores are labeled separately and their imports are routed only to the sandbox endpoint. Production stores continue using Seller Central OAuth and the production endpoint.

## Demo analytics data

After connecting a store, choose **Generate 6-month demo data** on the dashboard. The generator is deterministic and idempotent and creates store-scoped products, inventory, orders with items, advertising campaigns and daily metrics, financial transactions, reports, revenue, refunds, and profit data. It never deletes non-demo records.

Analytics are available from `GET /v1/analytics/dashboard?days=180`. Demo generation is available from `POST /v1/analytics/demo/generate` with a `storeId` and `months` value.

## Quick Start

1. Copy `.env.example` to `.env`.
2. Run `docker compose up --build`.
3. Open the apps:

| App | URL |
| --- | --- |
| SaaS admin dashboard | `http://localhost:3005` |
| Storefront | `http://localhost:3006` |
| API | `http://localhost:4005` |

The Compose stack starts PostgreSQL, Redis, MinIO, API, SaaS admin web, and storefront together. MinIO is available internally to the API at `http://minio:9000` and is not exposed on the host by default, which avoids local port conflicts.

You can also use the root helper scripts:

```bash
npm run docker:up
npm run docker:up:detached
npm run docker:logs
npm run docker:down
```

## Local Backend

```bash
cd apps/api
go test ./...
go run ./cmd/api
```

## Local Frontend

```bash
cd apps/web
npm install
npm run dev
```
Email: owner@retailpulse.local
Password: RetailPulse@12345
