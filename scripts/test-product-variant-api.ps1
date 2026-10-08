$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$categoryBaseUrl = "http://localhost:8081/api/v1/catalog/categories"
$unitBaseUrl = "http://localhost:8081/api/v1/catalog/units"
$productBaseUrl = "http://localhost:8081/api/v1/catalog/products"
$variantBaseUrl = "http://localhost:8081/api/v1/catalog/product-variants"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

Write-Host "0. CREATE DEPENDENCIES"

$category = Invoke-RestMethod `
    -Method Post `
    -Uri $categoryBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Variant Test Category $unique"
        code = "VCAT-$unique"
    } | ConvertTo-Json)

$unit = Invoke-RestMethod `
    -Method Post `
    -Uri $unitBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Piece $unique"
        code = "PCS-$unique"
    } | ConvertTo-Json)

$product = Invoke-RestMethod `
    -Method Post `
    -Uri $productBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        category_id = $category.id
        name = "Variant Test Product $unique"
        product_code = "VPROD-$unique"
    } | ConvertTo-Json)

Write-Host "PASS - Dependencies created"


Write-Host "1. CREATE PRODUCT VARIANTS"

$v1 = Invoke-RestMethod `
    -Method Post `
    -Uri $variantBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        sku = "SKU-$unique-1"
        default_selling_price = "499.99"
        unit_id = $unit.id
    } | ConvertTo-Json)

$v2 = Invoke-RestMethod `
    -Method Post `
    -Uri $variantBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        sku = "SKU-$unique-2"
        default_selling_price = "599.99"
        unit_id = $unit.id
    } | ConvertTo-Json)

if (-not $v1.id -or -not $v2.id) {
    throw "Create variants failed: missing IDs"
}

Write-Host "PASS - Created variants $($v1.id) and $($v2.id)"


Write-Host "2. GET PRODUCT VARIANT"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl/$($v1.id)"

if ($fetched.sku -ne "SKU-$unique-1") {
    throw "Get failed: wrong SKU"
}

Write-Host "PASS - Variant fetched"


Write-Host "3. LIST PRODUCT VARIANTS WITH FILTERS & PAGINATION"

$listAll = Invoke-RestMethod `
    -Method Get `
    -Uri $variantBaseUrl

if (-not ($listAll.items | Where-Object { $_.id -eq $v1.id })) {
    throw "List failed: variant 1 not found in items"
}

# Filter by product_id
$filteredByProduct = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl`?product_id=$($product.id)"

if ($filteredByProduct.items.Count -lt 2) {
    throw "Filter by product_id failed"
}

# Filter by unit_id
$filteredByUnit = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl`?unit_id=$($unit.id)"

if (-not ($filteredByUnit.items | Where-Object { $_.id -eq $v1.id })) {
    throw "Filter by unit_id failed"
}

# Filter by is_active
$filteredByActive = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl`?is_active=true"

if (-not ($filteredByActive.items | Where-Object { $_.id -eq $v1.id })) {
    throw "Filter by is_active failed"
}

# Test limit 100 accepted
$limit100 = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl`?limit=100"

if ($null -eq $limit100.items) {
    throw "Limit 100 test failed"
}

# Test limit 101 rejected
try {
    Invoke-RestMethod `
        -Method Get `
        -Uri "$variantBaseUrl`?limit=101"
    throw "Expected limit 101 to fail with 400"
} catch {
    if ($_.Exception.Response.StatusCode.value__ -ne 400) {
        throw "Expected 400 for limit 101"
    }
}

# Test limit 0 rejected
try {
    Invoke-RestMethod `
        -Method Get `
        -Uri "$variantBaseUrl`?limit=0"
    throw "Expected limit 0 to fail with 400"
} catch {
    if ($_.Exception.Response.StatusCode.value__ -ne 400) {
        throw "Expected 400 for limit 0"
    }
}

# Test invalid cursor rejected
try {
    Invoke-RestMethod `
        -Method Get `
        -Uri "$variantBaseUrl`?cursor=invalidcursor!!!"
    throw "Expected invalid cursor to fail with 400"
} catch {
    if ($_.Exception.Response.StatusCode.value__ -ne 400) {
        throw "Expected 400 for invalid cursor"
    }
}

Write-Host "PASS - Listing, pagination, and filters validated"


Write-Host "4. BATCH LOOKUP BY IDS"

$batchResult = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl`?ids=$($v1.id),$($v2.id)"

if ($batchResult.items.Count -ne 2) {
    throw "Batch lookup failed: expected 2 items, got $($batchResult.items.Count)"
}

# Batch lookup with invalid UUID
try {
    Invoke-RestMethod `
        -Method Get `
        -Uri "$variantBaseUrl`?ids=$($v1.id),not-a-uuid"
    throw "Expected invalid UUID in ids to fail with 400"
} catch {
    if ($_.Exception.Response.StatusCode.value__ -ne 400) {
        throw "Expected 400 for invalid UUID in batch lookup"
    }
}

# Batch lookup with 101 IDs
$manyIDs = (1..101 | ForEach-Object { [guid]::NewGuid().ToString() }) -join ","
try {
    Invoke-RestMethod `
        -Method Get `
        -Uri "$variantBaseUrl`?ids=$manyIDs"
    throw "Expected 101 IDs in batch lookup to fail with 400"
} catch {
    if ($_.Exception.Response.StatusCode.value__ -ne 400) {
        throw "Expected 400 for 101 IDs in batch lookup"
    }
}

Write-Host "PASS - Batch variant lookup validated"


Write-Host "5. UPDATE VARIANT"

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$variantBaseUrl/$($v1.id)" `
    -ContentType "application/json" `
    -Body (@{
        sku = "SKU-$unique-1-UPD"
        default_selling_price = "549.99"
        unit_id = $unit.id
    } | ConvertTo-Json)

if ($updated.sku -ne "SKU-$unique-1-UPD") {
    throw "Update failed: wrong SKU"
}

Write-Host "PASS - Variant updated"


Write-Host "6. DEACTIVATE VARIANT"

$deactivated = Invoke-RestMethod `
    -Method Patch `
    -Uri "$variantBaseUrl/$($v1.id)/status" `
    -ContentType "application/json" `
    -Body (@{
        is_active = $false
    } | ConvertTo-Json)

if ($deactivated.is_active -ne $false) {
    throw "Deactivate failed"
}

Write-Host "PASS - Variant deactivated"


Write-Host ""
Write-Host "=========================================="
Write-Host "ALL PRODUCT VARIANT API TESTS PASSED"
Write-Host "=========================================="
