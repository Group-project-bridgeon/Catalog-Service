# Catalog Stock Foundation slice 4a

Implementation contract, 2026-10-23. Authorized scope: Catalog-only internal
operations and coherent reference reads. Inventory is a read-only dependency;
Authentication and ERP-Platform are unchanged. This document records decisions
before implementation. Actual verification will be appended after execution.

## Ownership, legacy and immutable meaning

Catalog owns physical unit meaning and variant-specific pack definitions, not
stock quantities/history. Migration 000003 is a forward migration after 000002.
New side tables explicitly classify a unit as PIECE and make a variant eligible
with its existing unit_id. Empty side tables mean ALL existing variants remain
ineligible. No label/code inference, conversion/backfill or automatic packs.

ClassifyPieceUnit requires an existing active unit and an explicit trusted internal
management decision. EnableStock requires an existing active product/variant/unit,
explicit PIECE classification, and the exact expected existing unit binding.
Classification and eligibility are permanent: deactivate through existing active
flags, not by deleting/reinterpreting eligibility. A PIECE unit's physical meaning
cannot change under that UUID. Once eligible, a variant's product/unit/identity is
frozen globally, even before any warehouse has stock. Labels and active flags may
change. Existing variant update API already excludes product_id/unit_id; database
triggers and local FKs protect direct SQL/concurrent internal paths too.

Catalog never asks Inventory whether stock exists. A different physical meaning
requires a new identity or a separately approved migration. Legacy classification
requires operator review of actual meaning; BOX-like legacy assignments are not
silently reclassified because their labels were changed. These internal operations
have no HTTP/startup wiring and establish no management authorization policy.

## Versioned variant packs

Stable UUIDv7 pack ID, immutable normalized ASCII code (1-50 uppercase letters,
digits, underscore/hyphen), local eligible-variant ownership and active head.
Unique (variant_id,code); no universal conversion or recursive nesting.
Immutable revisions start at 1 and advance by 1. Each stores factor (positive
int64) and label (trimmed, 1-200 Unicode characters, no control characters).
Changing factor OR label appends a revision; never overwrite historical meaning.
RevisePack requires expected current revision; concurrent writers serialize on
the pack row and exactly one can win that expectation. SetPackActive uses the
expected head state version. All int64 factors/revisions/versions serialize as
canonical decimal strings, range 1..9223372036854775807.

Historical revisions remain readable with separate is_active/is_current flags.
Only an active head AND its current revision can be selected for new posting.
The resolver returns an explicit obsolete-revision outcome rather than treating
historical existence as current validity. Inventory's future adapter must reject
non-OK outcomes and fold current+active into its PackSnapshot.Active contract.

## Coherent references and revisions

ResolveBatch accepts 1..100 distinct variant requests, each with optional pack UUID
and exact decimal revision, plus optional expected stock-reference revision.
Zero UUIDs, duplicates, malformed/unsupported fields or oversized batches are
invalid input. Missing/mismatched/inactive/ineligible references are explicit
per-item outcomes; no implicit fallback factor or omitted requested entry.

One read-only REPEATABLE READ transaction covers observation time, all requested
variant/product/unit/eligibility rows and selected pack revisions. The entire batch
shares that snapshot/time, including missing results; related data is not read
from caches or independently visible states. Batch queries are bounded by IDs,
with at most one base query and one selected-pack query, plus observation time.
No Catalog read transaction survives return or waits for Inventory posting.

Each successful observation supplies product/variant/base-unit IDs, PIECE kind,
eligibility and active flags, product code/name, SKU, unit code/name, selected pack
ID/code/revision/factor/label and active/current flags, checked_at and opaque
stock_reference_revision. Inventory snapshot limits are respected: product code
50/name 200, SKU 100, base-unit code 30; pack label 200 fits Inventory TEXT.

Revision algorithm v1: SHA-256 over an explicit versioned ordered struct encoded
with encoding/json (no maps/floats), prefixed with cat-stock-v1:. Include every
returned posting-relevant identity, label, eligibility/semantic/active/current
field and monotonic entity stock versions. Product/variant/unit triggers advance
their counters only on contributing changes, including changes restored to prior
values (no ABA reuse). Pack revisions/head counters cover factor/label/status.
Observation time and caller expectations are excluded. Prices, unrelated
attributes/category data and generic updated_at are excluded. The revision is
neither a timestamp nor a PostgreSQL transaction ID and fits Inventory's 128 chars.

R1: active at the coherent Catalog check is required. Deactivation after this
observation but before Inventory commit is accepted; there is no distributed
active-at-commit guarantee. No approved freshness duration is invented. The later
authenticated Inventory adapter must enforce the separately approved freshness
policy and reject non-OK results, stale expected revisions and obsolete packs.

## Internal API and transport boundary

Internal service operations: ClassifyPieceUnit, EnableStock, CreatePack,
RevisePack, SetPackActive, GetPackRevision and ResolveBatch. Repositories own local
transactions, lock timeouts and bounded independent cleanup. Pure domain helpers
return recognizable errors; owning boundaries log safe classified diagnostics once.

Proposed, UNREGISTERED transport: POST /api/v1/catalog/stock-references/batch,
request {items:[{variant_id,pack_id?,pack_revision?,expected_revision?}]}, response
{checked_at,items:[{variant_id,status,reference?}]}. Reference fields use snake_case;
int64 values use strings. Successful responses must be no-store, no stale cache.
Proposed dedicated GLOBAL permission catalog.stock-references:read and a dedicated
Catalog service audience (candidate: catalog-stock-references) must be finalized
in slice 4b; the current erp-api human
token audience is not silently sufficient. No permission definition/grants, static
secret, fake identity, admin-token fallback or route is added in slice 4a.
Management HTTP APIs also remain unregistered because no approved policy/routes
currently define eligibility/pack management. Internal method visibility is not
authentication; future trusted adapters must authorize before invoking them.

## Schema, rollback and verification

Migration adds explicit PIECE/eligibility side tables, pack heads/revisions,
named local FKs/checks/uniques, relevant-field version triggers and immutable-fact
guards. Existing normal CRUD remains compatible; new columns are internal counters,
not changes to existing HTTP DTOs. New FK and pack PK/unique indexes support actual
lookup/lock queries; no cross-service FKs or normal DB checks.

Down removes the new metadata, guards and counters while preserving legacy Catalog
rows. It is destructive to classification/pack/revision evidence and must not be
used casually once other systems rely on it. Normal rollback favors compatible
application rollback/backup recovery and separately reviewed deployment; only
disposable fixtures exercise down here. Existing migrations are never rewritten.

Verification: pinned sqlc 1.31.1 generation twice with hash comparison; guarded
uniquely labelled tmpfs PostgreSQL fixtures with generated credentials, isolated
database/role/port/marker checks and run-owned cleanup; migration up/down/reapply
with sentinel Catalog rows; unit/service/repository and concurrency/invariant/
revision/coherence compatibility tests; existing Catalog route/security regression
suite, gofmt, vet/build and whitespace. Workers return errors through bounded
channels, never fatal assertions or correctness sleeps. No normal .env, database,
volume, service restart, deployment, commit/push/merge or sibling edits.

## Exact internal contracts and canonical v1

Actual entry points: service.NewStockReferenceService and
postgres.NewStockReferenceRepository. They are not constructed in app startup.
ResolveBatch's output is domain.StockReferenceBatch; management returns
domain.VariantPack or a recognizable domain error. Existing Unit/ProductVariant
HTTP DTOs and permission policies do not expose the new side-table metadata.

Per-item statuses: ok, variant_not_found, pack_not_found,
pack_variant_mismatch, pack_revision_not_found, pack_revision_obsolete,
not_stock_eligible, reference_inactive, reference_changed and invalid_metadata.
Missing Catalog parents/units cannot exist under normal local FKs. A missing
variant still gets one explicit item. Legacy invalid IDs/labels return
invalid_metadata; zero-ID parents cannot become stock eligible. Only status=ok
is usable by the future Inventory adapter. For invalid pack requests, any returned
base reference is diagnostic metadata, never permission to fall back to pieces.
Pack mismatch/missing/obsolete outcomes precede active/eligibility checks;
expected_revision is compared only for otherwise-OK observations.

Canonical v1 top-level JSON key order and names (case-sensitive): Version,
ProductID, VariantID, BaseUnitID, Kind, Eligible, ProductActive, VariantActive,
UnitActive, ProductCode, ProductName, SKU, UnitCode, UnitName, ProductVersion,
VariantVersion, UnitVersion, Pack. Version is catalog.stock-reference.v1.
UUIDs use Go uuid's lowercase canonical string encoding; booleans are JSON
booleans, StockCount values are positive decimal strings. Pack is null for pieces,
otherwise its keys are id, variant_id, code, revision, current_revision, version,
pieces_per_pack, label, is_active, is_current in that order. encoding/json's
standard string escaping is part of this v1 format. Any future encoding/meaning
change requires a new revision algorithm version, not silent v1 modification.
Hash the UTF-8 canonical bytes with SHA-256, lower-case hex, prefix cat-stock-v1:.
The resulting revision is exactly 77 ASCII characters. The fixed fixture with
UUID suffixes 1/2/3, labels P/Product/SKU/PC/Piece, all flags true, all versions 1
and no pack has hash:

    cat-stock-v1:dee7899214f51b1d0441c493ad3592db2cbb78740b081bbfcde3710fd73dffcd

The golden test pins that independently computed value. Time and caller
expectations do not enter the canonical object. A change to an unselected pack
does not alter a direct-piece operation's reference; only contributing state is
versioned. Counters prevent hash reuse when a contributing value changes back.

Example preparation request (proposed transport, NOT a registered route):

```json
{"items":[{"variant_id":"01935f00-0000-7000-8000-000000000002"}]}
```

Illustrative response using the pinned synthetic canonical fixture (observation
time is explanatory, not a claimed live response):

```json
{
  "checked_at": "2026-10-23T17:00:00Z",
  "items": [{
    "variant_id": "01935f00-0000-7000-8000-000000000002",
    "status": "ok",
    "reference": {
      "product_id": "01935f00-0000-7000-8000-000000000001",
      "variant_id": "01935f00-0000-7000-8000-000000000002",
      "base_unit_id": "01935f00-0000-7000-8000-000000000003",
      "quantity_kind": "PIECE", "eligible": true,
      "product_active": true, "variant_active": true, "base_unit_active": true,
      "product_code": "P", "product_name": "Product", "sku": "SKU",
      "base_unit_code": "PC", "base_unit_name": "Piece",
      "product_version": "1", "variant_version": "1", "unit_version": "1",
      "stock_reference_revision": "cat-stock-v1:dee7899214f51b1d0441c493ad3592db2cbb78740b081bbfcde3710fd73dffcd",
      "checked_at": "2026-10-23T17:00:00Z"
    }
  }]
}
```

Pack data has this shape (synthetic example, factor/revision are strings):

```json
{
  "id": "01935f00-0000-7000-8000-000000000004",
  "variant_id": "01935f00-0000-7000-8000-000000000002",
  "code": "BOX6", "revision": "1", "current_revision": "1", "version": "1",
  "pieces_per_pack": "6", "label": "Box of six", "is_active": true, "is_current": true
}
```

Internal operation budgets: 5s normal work, transaction-local 500ms lock wait/5s
statement timeout, bounded independent 2s rollback. Caller cancellation can shorten
normal work; cleanup may extend wall time. These engineering limits do not approve
a Catalog observation age. No automatic mutation retries are added. A transport
failure at management Commit returns ErrStockOutcomeUnknown; callers reconcile
the generated pack ID/current revision before retrying. CreatePack preserves its
generated ID alongside errors for that purpose. Revision/state changes use CAS;
future management HTTP idempotency requires its own reviewed transport contract.
No stock-posting or exactly-once management API claim is made.

Safe diagnostic errors expose StockFailure's component, operation, fixed reason
and SQLSTATE, with raw cause hidden from Error(). Owning future adapters log once
with request_id. Pure validation does not log; no authenticated adapter is wired.
Normal Catalog runtime roles currently own objects; administrators/owners can
disable guards. This is database-backed application integrity, not WORM storage.

## Inventory compatibility and deployment follow-up

| Catalog observation | Later Inventory adapter responsibility |
| --- | --- |
| status=ok, eligible=true, quantity_kind=PIECE | Require all three; Inventory's pure ReferenceSnapshot alone does not establish Catalog eligibility/provenance |
| product/variant/base-unit IDs and active flags | Map into ReferenceSnapshot and validate against the command variant; no cross-service FK/query |
| stock_reference_revision, checked_at | Retain exact revision/time; compare the caller's expectation and enforce a separately approved age policy |
| pack ID/revision/pieces_per_pack | Map exact positive int64 factor and selected revision; use overflow-safe count conversion in Inventory |
| pack is_active and is_current | Reject any non-OK result; only active AND current can map to Inventory PackSnapshot.Active=true |
| product code/name, SKU, unit code, optional pack label | Preserve historical movement labels within Inventory's existing limits; no price/float conversion |

These are inspected contract mappings and synthetic tests, not a live Inventory
adapter or authenticated production communication. Inventory application code and
its specifications remain unchanged by this Catalog-only slice.

Future deployment (not executed): obtain migration/operational approval and backup;
audit legacy physical meanings without inferring from codes; apply 000003 before
deploying binaries whose generated queries select stock_version; deploy a compatible
Catalog build with new transports still unregistered; separately authorize the
management process. Classify only genuinely single-piece units and explicitly enable
each matching variant. Genuine legacy box assignments need new identities or a
separately approved migration, not silent rebinding. Add only explicitly reviewed
variant packs. No real legacy rows were classified in this work.

Application rollback must remain schema-compatible. Down discards pack/eligibility
history and removes permanence guards, so it is unsuitable once consumers depend
on those definitions without coordinated recovery/review. The down migration fails
on unrelated external dependencies instead of broadly cascading their deletion.

Next roadmap slice 4b: approve and implement actual machine identity/token issuance,
dedicated audience, least-privilege stock-reference permission, revocation/rotation
and verifier compatibility. No new normal-user grants are implied. Then slice 4c:
register the reviewed secured reference transport, implement Inventory's authenticated
adapter, align budgets/body bounds and finalize/enforce freshness. Catalog's internal
5s ceiling is not an Inventory adapter latency promise. Real disposable cross-service
contract acceptance is required before production posting, including direct pieces.
Eligibility/pack management HTTP routes/authorization and operational audit/recovery
remain separately specified follow-up; existing CRUD grants are not silently reused.

## Implemented and verified checkpoint (2026-10-23)

Verdict: SLICE 4a COMPLETE for the authorized Catalog internal scope. This updates
the implementation-contract checkpoint above without rewriting Inventory's earlier
slice history. The Catalog README links this contract. It is not a claim that
Inventory-Catalog production communication or the whole Stock Foundation works.

- Forward/down migration 000003, domain/repository/service support, coherent batch
  reads, permanent physical definitions and immutable pack history are implemented.
  No existing applied migration or handwritten Catalog CRUD/security implementation
  changed. Generated legacy query scans include the new internal counters; existing
  HTTP DTOs remain unchanged.
- Two final sqlc/sqlc:1.31.1 generations produced identical hashes across all 15
  generated files. Nine logically unchanged generated files were restored to their
  original byte/line-ending representation; no generated logic was hand-edited.
- The final disposable harness passed six PostgreSQL integration groups plus the
  pure safe-error-mapping check (package time 3.909s). It verified empty/populated
  down/up/reapply, preserved sentinel legacy identities/labels, and confirmed no
  automatic eligibility or packs. Both head/revision and direct SQL bypass guards,
  zero-ID legacy safety, active/historical semantics, counter/hash changes, maximum
  batch size, concurrent revision CAS and concurrent snapshot coherence passed.
- Focused domain/service/router tests passed; the domain golden/serialization
  test independently pins the canonical v1 hash. Full serial Catalog regression
  `go test -p 1 -count=1 -timeout=120s ./...` passed, including existing handler,
  router/security and cache tests. Ungated PostgreSQL tests skip; runtime evidence
  comes from the explicitly gated disposable harness, not those skips.
- `go vet ./...`, `go build ./...`, final focused repository vet/tests, formatting,
  tracked/untracked whitespace and PowerShell parser checks passed. The sandbox
  build emitted a nonfatal module-cache stat-write warning and exited 0. An initial
  sqlc multi-array unnest parser error was corrected with supported typed syntax;
  a diff command initially used the workspace parent and was rerun correctly in
  Catalog. No application/test failures remain from those tooling issues.
- Fixtures used generated credentials, loopback random ports, marker/role/database
  gates and uniquely labelled tmpfs PostgreSQL. Containers/temp files were removed
  and prior environment restored; no normal DB, volume or service was touched.
- Branch feat/catalog-stock-foundation starts from b60cb583f5009fe0272013e80ac8704500aa77fb.
  Catalog was initially clean. Only this slice's source/docs/migration/generated
  files changed; all Inventory/Authentication/ERP-Platform baseline file hashes
  match (83/85/23 files respectively). No commit/push/PR/merge/deployment occurred.

Evidence and detailed file list: workspace review-artifacts/
catalog-stock-slice4a-20261023T165542Z-538cd7a2. Tests use real disposable PostgreSQL
with synthetic catalog records and deterministic barriers; no machine tokens,
real cross-service HTTP, network disruptions or normal-data classification were
tested or claimed. Legacy production data was not audited. Authentication,
transport, approved freshness and management authorization are still prerequisites.

## Slice 4b machine boundary (2026-10-24)

The approved machine contract is implemented separately in security/machine.go.
It uses the configured Auth issuer and RSA public key, RS256 only, dedicated
audience catalog-stock-references, canonical machine:<UUID> subject/client_id,
token_kind=machine, UUID jti, and exact catalog.stock-references:read permission.
The strict claim decoder rejects duplicate, absent, null, malformed and unknown
claims. Integer iat/nbf/exp are mandatory; nbf equals iat; exp-iat is independently
limited to 300 seconds. A maximum 30-second clock tolerance does not extend the
permitted issued lifetime. Missing authority is denied by the dedicated middleware.

LoadMachine reuses public-key loading validation. It is not installed in the human
router or startup path; slice 4c will explicitly load/register this separate
boundary. Human erp-api tokens do not authenticate here, and machine tokens lack
the human identity/audience contract. The reference batch route remains unregistered.

Auth owns separate identities, hashed random credentials, grants and issuance.
See [the Auth contract](../../Authentication/docs/machine-identity-spec.md). Credential expiry,
revocation, machine disable and grant removal stop future issuance; locally verified
tokens may remain accepted until expiry plus 30 seconds. There is no online JWT
revocation. Production communication, transport registration, Inventory adapter,
freshness enforcement and normal provisioning remain outside this slice.

## Slice 4c transport checkpoint (2026-10-24)

This supersedes the slice 4b unregistered-route/startup status above; the historical
4b boundary remains documented. Optional CATALOG_STOCK_REFERENCES_ENABLED=true
loads the dedicated machine verifier and registers exactly
POST /api/v1/catalog/stock-references/batch against the coherent service/repository.
It requires paired TLS_CERT_PATH/TLS_KEY_PATH; direct TLS is checked at the handler,
and forwarded headers do not establish trust. The disabled default preserves the
existing listener/routing behavior. Existing human business routes do not accept
machine JWTs, and human JWTs do not authenticate the reference route.

The request is application/json without media parameters, no query, at most 64 KiB
and 100 unique variants. Unknown/duplicate/case-ambiguous fields, null values,
invalid UTF-8, trailing JSON and malformed identifiers/revisions reject. The exact
method is the only registered reference method. Responses preserve one explicit
status per requested variant and canonical decimal-string int64 values, with
Cache-Control: no-store. Cancellation reaches the coherent repository. Transport
errors are generic; diagnostics carry component/operation/reason/request_id and
safe SQLSTATE where relevant, without duplicate generic URL failure logs.

Inventory now consumes these references over verified HTTPS using Auth machine
issuance, without a reference cache. The user approved maximum observation age 5s
and future-clock allowance 1s. Age checks include response acceptance and the point
after Inventory pool acquisition immediately before each new BEGIN. R1 permits
deactivation after the coherent check and before Inventory commits; no
active-at-commit guarantee is created. JWT tolerance remains independently at
most 30s, with issued lifetime at most 300s.

Actual temporary-trust HTTPS/three-database posting passed with synthetic fixture
data, along with Auth lifecycle checks and human/machine router regressions.
See [Inventory integration documentation](../../Inventory-Service/docs/catalog-reference-integration.md)
and workspace review-artifacts/stock-slice4c-20261024/report.md for configuration,
coverage, deployment order and limitations. Normal TLS/upstream/probe changes,
machine/grant provisioning, credential distribution, normal Catalog classification
and production stock workflow authorization/wiring remain separately required.
No normal migration, service restart, Kong change or deployment occurred here.
