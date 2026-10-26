# Local stock-definition administration

Calls the existing StockReferenceService for an explicit local selection:

```text
catalog-local-admin --local-host postgres --local-database catalog_db --selection-file /protected/catalog-selection.json
```

DATABASE_URL comes from the owning service's environment. The explicit host/database
must match configuration; only postgres/loopback hosts and matching fallback hosts
are accepted. Selection file paths must be absolute. Use selection.example.json as
a shape template; its UUID placeholders intentionally cannot be executed unchanged.

Inspect every variant sharing the unit before confirming that one base unit is
one physical piece for all of them. The confirmation is unit-wide; eligibility is
added only for the explicit unique variants. Input is limited to 64 KiB, 1–100
variants, strict known fields and unambiguous JSON. Missing/null/duplicate IDs,
duplicate JSON fields, trailing data and absent physical confirmation reject.

No public provisioning route, packs, rebinding, stock posting or broad grant is
created. Existing local service transactions remain the write boundary; a plan
with multiple variants can partially complete. On failure, inspect classification
and eligibility by their IDs before an authorized continuation. Do not infer
rollback from a timeout, delete definitions or blindly rerun the plan.

See ERP-Platform/docs/opening-stock-local-operations.md in the adjacent checkout.
