# Catalog Service

Stock Foundation slice 4a now provides internal piece eligibility, immutable
variant/unit bindings, versioned packs and coherent reference batches. See the
[Catalog stock-reference contract and roadmap](docs/stock-reference-spec.md).
Legacy variants remain ineligible until explicitly classified. No stock-reference
or management HTTP route is registered; machine authentication, secure transport
and approved freshness enforcement remain later integration work. Migration 000003
was verified only in disposable PostgreSQL, not applied to normal databases.

All business endpoints require an RS256 bearer token and exact GLOBAL permissions.
See [security configuration and the 58-endpoint permission matrix](docs/security.md).
Only GET /health and GET /ready are public. Direct host port 8081 is opt-in in ERP-Platform.


Catalog Service is the catalog-management microservice for the Fashion Wholesale ERP.

It owns catalog-related data such as categories, brands, units, products, product variants, attributes, variant attributes, product images, price lists, and prices.

The service is implemented in Go using Gin, PostgreSQL, pgx, sqlc, and Zerolog.

---

## Tech Stack

- Go 1.25.4
- Gin
- PostgreSQL
- pgx/v5
- sqlc
- golang-migrate
- Zerolog
- Google UUID

---

## Responsibilities

Catalog Service currently manages these 11 catalog areas:

| # | Area | Purpose |
|---|---|---|
| 1 | Category | Product category hierarchy and activation status |
| 2 | Brand | Product brands |
| 3 | Unit | Units used by product variants |
| 4 | Product | Main product records |
| 5 | Product Variant | SKU-level product variants |
| 6 | Attribute | Variant attributes such as color or size |
| 7 | Attribute Value | Allowed values for attributes |
| 8 | Variant Value | Assigns attribute values to product variants |
| 9 | Product Image | Product and variant image metadata |
| 10 | Price List | Named price lists and currencies |
| 11 | Price | Current variant price for a price list |

Catalog Service owns only:

```text
catalog_db
```

using the database user:

```text
catalog_service
```

It must not directly query another service's database.

---

## Project Structure

```text
Catalog-Service/
├── cmd/
│   └── server/
│       └── main.go
│
├── db/
│   ├── migrations/
│   │   ├── 000001_create_catalog_schema.up.sql
│   │   └── 000001_create_catalog_schema.down.sql
│   │
│   └── queries/
│       ├── attributes.sql
│       ├── attribute_values.sql
│       ├── brands.sql
│       ├── categories.sql
│       ├── prices.sql
│       ├── price_lists.sql
│       ├── products.sql
│       ├── product_images.sql
│       ├── product_variants.sql
│       ├── units.sql
│       └── variant_values.sql
│
├── internal/
│   ├── app/
│   ├── config/
│   ├── domain/
│   ├── handler/
│   ├── middleware/
│   ├── platform/
│   │   ├── database/
│   │   └── logger/
│   ├── repository/
│   │   └── postgres/
│   │       └── catalogdb/
│   ├── router/
│   └── service/
│
├── scripts/
├── .env.example
├── go.mod
├── go.sum
├── README.md
└── sqlc.yaml
```

---

## Requirements

Before running Catalog Service, install or have available:

- Go 1.25.4 or compatible version
- Docker Desktop
- PostgreSQL through the ERP-Platform infrastructure
- golang-migrate CLI
- Git

sqlc does not have to be installed locally because code generation can be run using Docker.

---

## Database Setup

PostgreSQL infrastructure is managed by the `ERP-Platform` repository.

The platform creates:

```text
catalog_service -> catalog_db
```

Catalog Service itself owns and applies the schema migrations inside `catalog_db`.

### 1. Start PostgreSQL

From the `ERP-Platform` repository:

```powershell
docker compose up -d postgres
```

Check that PostgreSQL is healthy:

```powershell
docker compose ps
```

---

### 2. Create Local Environment Configuration

From the Catalog Service repository:

```powershell
Copy-Item .\.env.example .\.env
```

Current `.env.example`:

```env
PORT=8081
DATABASE_URL=postgres://catalog_service:change_me_catalog_password@127.0.0.1:5433/catalog_db?sslmode=disable
```

Replace the placeholder database password in `.env` with the local Catalog database password configured in ERP-Platform.

Do not commit `.env`.

The current local configuration expects PostgreSQL to be available on host port:

```text
5433
```

If the ERP-Platform PostgreSQL host port is different, update `DATABASE_URL`.

---

## Environment Variables

| Variable | Description | Example |
|---|---|---|
| `PORT` | HTTP port used by Catalog Service | `8081` |
| `DATABASE_URL` | PostgreSQL connection string for `catalog_db` | `postgres://catalog_service:...@127.0.0.1:5433/catalog_db?sslmode=disable` |

---

## Database Migrations

Catalog Service migrations are stored in:

```text
db/migrations/
```

Current migration:

```text
000001_create_catalog_schema.up.sql
000001_create_catalog_schema.down.sql
```

Load the Catalog database URL into the PowerShell environment:

```powershell
$env:DATABASE_URL = (Get-Content .\.env |
    Where-Object { $_ -match '^DATABASE_URL=' }) -replace '^DATABASE_URL=', ''
```

Apply all migrations:

```powershell
migrate -path .\db\migrations -database $env:DATABASE_URL up
```

Check the current migration version:

```powershell
migrate -path .\db\migrations -database $env:DATABASE_URL version
```

Rollback the latest migration when required:

```powershell
migrate -path .\db\migrations -database $env:DATABASE_URL down 1
```

Do not manually modify an already shared migration to make later schema changes. Add a new sequential migration instead.

---

## Generate sqlc Code

SQL queries are stored in:

```text
db/queries/
```

sqlc configuration is stored in:

```text
sqlc.yaml
```

Generated database code is written to:

```text
internal/repository/postgres/catalogdb/
```

Generate the code using Docker:

```powershell
docker run --rm -v "${PWD}:/src" -w /src sqlc/sqlc generate
```

Do not manually edit files inside:

```text
internal/repository/postgres/catalogdb/
```

They are generated by sqlc.

---

## Format the Code

Run:

```powershell
gofmt -w .\cmd .\internal
```

---

## Run Unit Tests

Run the complete Go test suite:

```powershell
go test ./...
```

For verbose output:

```powershell
go test ./... -v
```

The router test suite also includes a route-registration smoke test to detect duplicate or conflicting Gin routes.

---

## Run Catalog Service

Start the service from the repository root:

```powershell
go run ./cmd/server
```

By default the service runs on:

```text
http://localhost:8081
```

---

## Health Checks

### Liveness

```http
GET /health
```

Example:

```powershell
Invoke-RestMethod http://localhost:8081/health
```

Expected response:

```json
{
  "service": "catalog",
  "status": "ok"
}
```

### Readiness

```http
GET /ready
```

This verifies that Catalog Service can communicate with PostgreSQL.

```powershell
Invoke-RestMethod http://localhost:8081/ready
```

---

# API

Base API path:

```text
/api/v1/catalog
```

---

## 1. Categories

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/categories` | Create category |
| GET | `/api/v1/catalog/categories` | List categories |
| GET | `/api/v1/catalog/categories/:id` | Get category |
| PUT | `/api/v1/catalog/categories/:id` | Update category |
| PATCH | `/api/v1/catalog/categories/:id/status` | Activate/deactivate category |

Categories are deactivated instead of physically deleted.

---

## 2. Brands

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/brands` | Create brand |
| GET | `/api/v1/catalog/brands` | List brands |
| GET | `/api/v1/catalog/brands/:id` | Get brand |
| PUT | `/api/v1/catalog/brands/:id` | Update brand |
| PATCH | `/api/v1/catalog/brands/:id/status` | Activate/deactivate brand |

---

## 3. Units

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/units` | Create unit |
| GET | `/api/v1/catalog/units` | List units |
| GET | `/api/v1/catalog/units/:id` | Get unit |
| PUT | `/api/v1/catalog/units/:id` | Update unit |
| PATCH | `/api/v1/catalog/units/:id/status` | Activate/deactivate unit |

---

## 4. Products

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/products` | Create product |
| GET | `/api/v1/catalog/products` | List products |
| GET | `/api/v1/catalog/products/:id` | Get product |
| PUT | `/api/v1/catalog/products/:id` | Update product |
| PATCH | `/api/v1/catalog/products/:id/status` | Activate/deactivate product |

---

## 5. Product Variants

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/product-variants` | Create variant |
| GET | `/api/v1/catalog/product-variants` | List variants |
| GET | `/api/v1/catalog/product-variants/:id` | Get variant |
| PUT | `/api/v1/catalog/product-variants/:id` | Update variant |
| PATCH | `/api/v1/catalog/product-variants/:id/status` | Activate/deactivate variant |
| GET | `/api/v1/catalog/products/:id/variants` | List variants for a product |

A product variant represents a sellable SKU and references a product and unit.

---

## 6. Attributes

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/attributes` | Create attribute |
| GET | `/api/v1/catalog/attributes` | List attributes |
| GET | `/api/v1/catalog/attributes/:id` | Get attribute |
| PUT | `/api/v1/catalog/attributes/:id` | Update attribute |

Examples of attributes include:

```text
Color
Size
Material
```

---

## 7. Attribute Values

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/attribute-values` | Create attribute value |
| GET | `/api/v1/catalog/attribute-values` | List attribute values |
| GET | `/api/v1/catalog/attribute-values/:id` | Get attribute value |
| PUT | `/api/v1/catalog/attribute-values/:id` | Update attribute value |
| GET | `/api/v1/catalog/attributes/:id/values` | List values belonging to an attribute |

Example:

```text
Attribute: Color

Values:
- Red
- Blue
- Black
```

---

## 8. Variant Values

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/variant-values` | Assign an attribute value to a variant |
| GET | `/api/v1/catalog/variant-values` | List variant values |
| GET | `/api/v1/catalog/variant-values/:variant_id/:attribute_id` | Get one variant attribute assignment |
| PUT | `/api/v1/catalog/variant-values/:variant_id/:attribute_id` | Update variant attribute assignment |
| GET | `/api/v1/catalog/product-variants/:id/values` | List attribute values assigned to a variant |

A variant can have one value for a given attribute.

Example:

```text
Variant: SHIRT-BLACK-L
Color -> Black
Size  -> Large
```

---

## 9. Product Images

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/product-images` | Create image metadata |
| GET | `/api/v1/catalog/product-images` | List image metadata |
| GET | `/api/v1/catalog/product-images/:id` | Get image metadata |
| PUT | `/api/v1/catalog/product-images/:id` | Update image metadata |
| GET | `/api/v1/catalog/products/:id/images` | List images for a product |
| GET | `/api/v1/catalog/product-variants/:id/images` | List images for a variant |

Catalog Service stores image metadata and the image URL.

It does not currently store the image binary itself.

---

## 10. Price Lists

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/price-lists` | Create price list |
| GET | `/api/v1/catalog/price-lists` | List price lists |
| GET | `/api/v1/catalog/price-lists/:id` | Get price list |
| PUT | `/api/v1/catalog/price-lists/:id` | Update price list |
| PATCH | `/api/v1/catalog/price-lists/:id/status` | Activate/deactivate price list |

Each price list contains a three-character currency code such as:

```text
INR
USD
EUR
```

---

## 11. Prices

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/catalog/prices` | Create price |
| GET | `/api/v1/catalog/prices` | List prices |
| GET | `/api/v1/catalog/prices/:id` | Get price |
| PUT | `/api/v1/catalog/prices/:id` | Update price |
| GET | `/api/v1/catalog/price-lists/:id/prices` | List prices under a price list |
| GET | `/api/v1/catalog/product-variants/:id/prices` | List prices for a variant |

Only one current price can exist for a given:

```text
price_list + variant
```

Prices use exact decimal values rather than floating-point values.

Negative prices are rejected.

---

# API Integration Tests

PowerShell integration scripts are stored in:

```text
scripts/
```

The Catalog server must already be running before executing these scripts.

Example:

```powershell
go run ./cmd/server
```

Then open another PowerShell terminal.

Run one test:

```powershell
.\scripts\test-price-api.ps1
```

Run all integration scripts:

```powershell
Get-ChildItem .\scripts\test-*-api.ps1 | ForEach-Object {
    Write-Host "`nRUNNING $($_.Name)"

    try {
        & $_.FullName
        Write-Host "PASS: $($_.Name)"
    }
    catch {
        throw "FAILED: $($_.Name) - $($_.Exception.Message)"
    }
}
```

A successful regression should finish without any thrown errors.

---

## Error Handling

Catalog Service maps expected failures to safe HTTP responses.

Typical statuses include:

| Status | Meaning |
|---|---|
| `400 Bad Request` | Invalid input |
| `404 Not Found` | Resource does not exist |
| `409 Conflict` | Unique/conflicting resource |
| `422 Unprocessable Entity` | Invalid referenced resource or relationship |
| `500 Internal Server Error` | Unexpected internal failure |

Internal database errors and stack traces are not returned to API clients.

Unexpected errors are logged using Zerolog.

---

## Logging

The service uses structured logging with Zerolog.

HTTP requests include information such as:

```text
method
path
status
duration
```

Request and correlation IDs are included when provided.

Sensitive values such as passwords and database credentials must not be logged.

---

## Development Workflow

For normal Catalog development:

```powershell
docker run --rm -v "${PWD}:/src" -w /src sqlc/sqlc generate

gofmt -w .\cmd .\internal

go test ./...
```

Then start the server:

```powershell
go run ./cmd/server
```

In another terminal, verify:

```powershell
Invoke-RestMethod http://localhost:8081/health
```

Then run the integration regression scripts.

---

## Database Ownership Rule

Catalog Service owns only:

```text
catalog_db
```

Direct database access such as this is allowed:

```text
Catalog Service -> catalog_db
```

Direct database access such as this is not allowed:

```text
Catalog Service -> inventory_db
Catalog Service -> sales_db
Catalog Service -> auth_db
```

Communication with other services must happen through APIs or events rather than cross-service database queries.

---

## Schema Changes

All schema changes must be represented by SQL migration files.

For future schema changes, create a new sequential migration:

```text
db/migrations/
├── 000001_create_catalog_schema.up.sql
├── 000001_create_catalog_schema.down.sql
├── 000002_some_future_change.up.sql
└── 000002_some_future_change.down.sql
```

Do not rely on manually changing PostgreSQL through pgAdmin or `psql` without adding the equivalent migration to Git.

---

## Current Status

The Catalog Service currently has complete end-to-end implementation for:

```text
Category
Brand
Unit
Product
Product Variant
Attribute
Attribute Value
Variant Value
Product Image
Price List
Price
```

The service includes:

```text
SQL migrations
sqlc queries and generated database access
domain models
repository interfaces
PostgreSQL repositories
service/business logic
HTTP handlers
Gin routes
validation
structured error handling
Zerolog logging
unit tests
route-registration smoke testing
PowerShell API integration tests
health and readiness endpoints
graceful shutdown
```

---

## Repository

```text
github.com/Group-project-bridgeon/Catalog-Service
```
