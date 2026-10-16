# Catalog security contract

Every business route requires a locally verified RS256 JWT and the exact GLOBAL
permission(s) in the endpoint matrix below. GET /health and GET /ready are public.
Readiness still depends on the Catalog database; Kong does not publish these
root-level probes through its existing API routes.

## Configuration and identity

Set JWT_PUBLIC_KEY_PATH, JWT_ISSUER=erp-auth and JWT_AUDIENCE=erp-api. The key
must be one PEM public-key block (PKIX PUBLIC KEY or PKCS#1 RSA PUBLIC KEY), RSA
2048 bits or stronger. Private keys, certificates, weak keys, invalid files and
blank issuer/audience fail startup before database connection. Catalog holds no
private signing key and performs no Auth HTTP request or Auth database query.

Verification requires RS256, a matching issuer/audience, exp, nbf, iat, a nonzero
UUID user_id, matching canonical UUID sub, a present nonnegative int32
token_version, and a non-null array of well-formed resource:action permissions.
Timestamps must be JSON numbers; exp > iat, nbf >= iat, exp > nbf, and
exp - iat <= 900 seconds. Future iat/nbf and expired exp are rejected with no
clock leeway. Synchronize server clocks. Email is informational, not an identity
or authorization key; Auth's existing validator does not require it.

Missing/malformed/invalid credentials return 401 with WWW-Authenticate: Bearer.
Valid credentials without every required exact permission return 403. Empty
permission arrays authenticate but cannot access business operations. Wildcards,
administrator labels, case changes and proxy-supplied identity headers confer
no implicit access. Verified claims are available through security.Identity(c).
Unmapped routes are denied, and route-coverage tests detect added or stale routes.

Auth's GetUserGlobalPermissions query includes only GLOBAL grants with NULL
scope_id. WAREHOUSE grants are deliberately excluded: Catalog has no scoped
warehouse authorization contract. Do not flatten those grants into this array.

Local verification does not immediately detect user deactivation, role/permission
changes or token_version increments. Existing tokens can remain usable until
expiry (at most 15 minutes after issuance). Auth /me performs its own database
revocation check; Catalog intentionally does not. After granting permissions,
log in again. Removing grants alone does not revoke already-issued Catalog tokens.

## Permission ownership and deployment

Authentication/db/migrations/000005_catalog_permissions.up.sql defines 33 grants
(11 resources x read/create/update), with no automatic role or user assignments.
Old migrations are unchanged. Broad catalog:read and catalog:write strings seen
in old Auth test fixtures are not Catalog endpoint grants. No existing seed data
defined an endpoint contract. This resource-level mapping lets a brand editor
operate without receiving price or product mutation permissions.

Status PATCH uses the resource's update grant. Nested collections require the
returned resource's read grant. Product detail returns products, variants,
variant values, images and prices, so it requires all five corresponding reads.

Back up Auth RBAC data before deploying migration 000005. The additive migration
uses ON CONFLICT DO NOTHING and preserves existing grants. Its down migration
intentionally retains definitions: deleting them could remove existing role
dependencies or pre-existing permissions. No business schema/data is changed.
Reverting to the old unauthenticated Catalog binary reopens access; keep the new
enforcement in place while correcting configuration or assignments.

Use ERP-Platform/scripts/test-catalog-security.ps1 for repeatable isolated test
assignments. It creates only fresh test identities and roles: brand read/create/
update GLOBAL, brand read-only GLOBAL, no grants, and WAREHOUSE-only. It never
selects or modifies existing users. New tests use temporary keys and passwords.

For a real development account, an authorized Auth administrator must select
the exact user, role and permissions explicitly, then use the existing Auth RBAC
tables/queries to add identity_role_permissions and a GLOBAL identity_user_roles
assignment with scope_id NULL. Do not grant every permission or automatically
grant all authenticated users. The user reported a controlled read-only test role assignment; the finalizing agent did not change accounts or permissions.

Deploy the Auth RS256 contract, Catalog verifier and Platform key/configuration changes together. Missing runtime key configuration can prevent startup; mismatched keys or old HS256 tokens prevent access. Migration 000005 must be applied, but explicit role assignments are still required.

## Verification evidence

The user reported the isolated Docker/Kong suite passing (exit 0), plus normal-stack anonymous 401, missing-read 403, fresh-login read 200, read-only POST 403, and direct Docker-network anonymous 401 checks. These were not rerun by the agent during Git finalization. They do not establish live CRUD coverage for every endpoint. See ERP-Platform/docs/catalog-security-results.md for attribution and limitations.

## Local checks

From Catalog-Service:

```powershell
go test -count=1 ./...
go vet ./...
go build ./...
go mod verify
```

The tests cover authentication failures, exact permission behavior, all 58
business endpoints, fail-closed unknown routes, public probes and real brand
HTTP handlers over an isolated in-memory fixture. Those tests are not evidence
of PostgreSQL persistence or Kong connectivity; use the platform runtime script.

Existing scripts/test-*-api.ps1 predate authentication and do not send JWTs.
Do not treat their unauthenticated failures as CRUD regressions. Use the new
security harness or explicitly authenticated requests for runtime checks.

## Gotchas

| Mistake | Cause | Prevention | Verification |
|---|---|---|---|
| New endpoint bypasses security | Route omitted from policy | Global middleware denies unmapped routes | TestEveryBusinessRouteEnforcesMatrix |
| Product detail leaks prices | Only checking product permission | Require all five component reads | Remove each grant in route tests |
| Warehouse user gains global access | Scope grants flattened at issuance | Preserve GLOBAL-only Auth query | Auth scope integration test and isolated WAREHOUSE login |
| Missing PEM becomes a directory | Short bind syntax creates host paths | create_host_path: false, public-file validation | Compose config and isolated non-root readability check |
| Revoked user retains access | Stateless JWT validation | Short token lifetime; document revocation window | Expiry tests; fresh login after grants |
| Removing port assumed sufficient | Requests can still arrive inside the network | Enforce JWT and permissions in Catalog | 401/403 route tests and Kong harness |

## Endpoint-to-permission matrix

Paths below are complete Catalog routes. All listed permissions are required.

| Method | Path | Required permissions |
|---|---|---|
| GET | /api/v1/catalog/categories | catalog.categories:read |
| GET | /api/v1/catalog/categories/:id | catalog.categories:read |
| POST | /api/v1/catalog/categories | catalog.categories:create |
| PUT | /api/v1/catalog/categories/:id | catalog.categories:update |
| GET | /api/v1/catalog/brands | catalog.brands:read |
| GET | /api/v1/catalog/brands/:id | catalog.brands:read |
| POST | /api/v1/catalog/brands | catalog.brands:create |
| PUT | /api/v1/catalog/brands/:id | catalog.brands:update |
| GET | /api/v1/catalog/units | catalog.units:read |
| GET | /api/v1/catalog/units/:id | catalog.units:read |
| POST | /api/v1/catalog/units | catalog.units:create |
| PUT | /api/v1/catalog/units/:id | catalog.units:update |
| GET | /api/v1/catalog/products | catalog.products:read |
| GET | /api/v1/catalog/products/:id | catalog.products:read |
| POST | /api/v1/catalog/products | catalog.products:create |
| PUT | /api/v1/catalog/products/:id | catalog.products:update |
| GET | /api/v1/catalog/product-variants | catalog.product-variants:read |
| GET | /api/v1/catalog/product-variants/:id | catalog.product-variants:read |
| POST | /api/v1/catalog/product-variants | catalog.product-variants:create |
| PUT | /api/v1/catalog/product-variants/:id | catalog.product-variants:update |
| GET | /api/v1/catalog/attributes | catalog.attributes:read |
| GET | /api/v1/catalog/attributes/:id | catalog.attributes:read |
| POST | /api/v1/catalog/attributes | catalog.attributes:create |
| PUT | /api/v1/catalog/attributes/:id | catalog.attributes:update |
| GET | /api/v1/catalog/attribute-values | catalog.attribute-values:read |
| GET | /api/v1/catalog/attribute-values/:id | catalog.attribute-values:read |
| POST | /api/v1/catalog/attribute-values | catalog.attribute-values:create |
| PUT | /api/v1/catalog/attribute-values/:id | catalog.attribute-values:update |
| GET | /api/v1/catalog/variant-values | catalog.variant-values:read |
| GET | /api/v1/catalog/variant-values/:variant_id/:attribute_id | catalog.variant-values:read |
| POST | /api/v1/catalog/variant-values | catalog.variant-values:create |
| PUT | /api/v1/catalog/variant-values/:variant_id/:attribute_id | catalog.variant-values:update |
| GET | /api/v1/catalog/product-images | catalog.product-images:read |
| GET | /api/v1/catalog/product-images/:id | catalog.product-images:read |
| POST | /api/v1/catalog/product-images | catalog.product-images:create |
| PUT | /api/v1/catalog/product-images/:id | catalog.product-images:update |
| GET | /api/v1/catalog/price-lists | catalog.price-lists:read |
| GET | /api/v1/catalog/price-lists/:id | catalog.price-lists:read |
| POST | /api/v1/catalog/price-lists | catalog.price-lists:create |
| PUT | /api/v1/catalog/price-lists/:id | catalog.price-lists:update |
| GET | /api/v1/catalog/prices | catalog.prices:read |
| GET | /api/v1/catalog/prices/:id | catalog.prices:read |
| POST | /api/v1/catalog/prices | catalog.prices:create |
| PUT | /api/v1/catalog/prices/:id | catalog.prices:update |
| PATCH | /api/v1/catalog/categories/:id/status | catalog.categories:update |
| PATCH | /api/v1/catalog/brands/:id/status | catalog.brands:update |
| PATCH | /api/v1/catalog/units/:id/status | catalog.units:update |
| PATCH | /api/v1/catalog/products/:id/status | catalog.products:update |
| PATCH | /api/v1/catalog/product-variants/:id/status | catalog.product-variants:update |
| PATCH | /api/v1/catalog/price-lists/:id/status | catalog.price-lists:update |
| GET | /api/v1/catalog/attributes/:id/values | catalog.attribute-values:read |
| GET | /api/v1/catalog/products/:id/variants | catalog.product-variants:read |
| GET | /api/v1/catalog/product-variants/:id/values | catalog.variant-values:read |
| GET | /api/v1/catalog/products/:id/images | catalog.product-images:read |
| GET | /api/v1/catalog/product-variants/:id/images | catalog.product-images:read |
| GET | /api/v1/catalog/price-lists/:id/prices | catalog.prices:read |
| GET | /api/v1/catalog/product-variants/:id/prices | catalog.prices:read |
| GET | /api/v1/catalog/products/:id/detail | catalog.products:read, catalog.product-variants:read, catalog.variant-values:read, catalog.product-images:read, catalog.prices:read |
