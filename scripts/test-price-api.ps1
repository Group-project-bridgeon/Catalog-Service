$categoryBaseUrl = "http://localhost:8081/api/v1/catalog/categories"
$productBaseUrl = "http://localhost:8081/api/v1/catalog/products"
$unitBaseUrl = "http://localhost:8081/api/v1/catalog/units"
$variantBaseUrl = "http://localhost:8081/api/v1/catalog/product-variants"
$priceListBaseUrl = "http://localhost:8081/api/v1/catalog/price-lists"
$priceBaseUrl = "http://localhost:8081/api/v1/catalog/prices"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

Write-Host "0. CREATE PRICE DEPENDENCIES"

$category = Invoke-RestMethod `
    -Method Post `
    -Uri $categoryBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Price Category $unique"
        code = "PRICE-CAT-$unique"
    } | ConvertTo-Json)

$unit = Invoke-RestMethod `
    -Method Post `
    -Uri $unitBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        code = "PU-$unique"
        name = "Price Unit $unique"
    } | ConvertTo-Json)

$product = Invoke-RestMethod `
    -Method Post `
    -Uri $productBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        category_id = $category.id
        name = "Price Product $unique"
        product_code = "PRICE-PROD-$unique"
    } | ConvertTo-Json)

$variant = Invoke-RestMethod `
    -Method Post `
    -Uri $variantBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        sku = "PRICE-SKU-$unique"
        unit_id = $unit.id
    } | ConvertTo-Json)

$priceList = Invoke-RestMethod `
    -Method Post `
    -Uri $priceListBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Price Test List $unique"
        code = "PRICE-LIST-$unique"
        currency_code = "INR"
    } | ConvertTo-Json)

Write-Host "PASS - Dependencies created"


Write-Host "1. CREATE PRICE"

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $priceBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        price_list_id = $priceList.id
        variant_id = $variant.id
        unit_price = "850.50"
    } | ConvertTo-Json)

$id = $created.id

if (-not $id) {
    throw "Create failed: price ID missing"
}

if ($created.price_list_id -ne $priceList.id) {
    throw "Create failed: wrong price list ID"
}

if ($created.variant_id -ne $variant.id) {
    throw "Create failed: wrong variant ID"
}

if ([decimal]$created.unit_price -ne [decimal]"850.50") {
    throw "Create failed: wrong unit price"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET PRICE"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$priceBaseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: wrong price ID"
}

Write-Host "PASS - Price fetched"


Write-Host "3. LIST ALL PRICES"

$prices = Invoke-RestMethod `
    -Method Get `
    -Uri $priceBaseUrl

$found = $prices | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "List failed: price not found"
}

Write-Host "PASS - Price found in list"


Write-Host "4. LIST PRICES BY PRICE LIST"

$priceListPrices = Invoke-RestMethod `
    -Method Get `
    -Uri "$priceListBaseUrl/$($priceList.id)/prices"

$found = $priceListPrices | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "Price not found under price list"
}

Write-Host "PASS - Price found under price list"


Write-Host "5. LIST PRICES BY VARIANT"

$variantPrices = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl/$($variant.id)/prices"

$found = $variantPrices | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "Price not found under variant"
}

Write-Host "PASS - Price found under variant"


Write-Host "6. UPDATE PRICE"

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$priceBaseUrl/$id" `
    -ContentType "application/json" `
    -Body (@{
        unit_price = "999.75"
    } | ConvertTo-Json)

if ([decimal]$updated.unit_price -ne [decimal]"999.75") {
    throw "Update failed: wrong unit price"
}

if ($updated.price_list_id -ne $priceList.id) {
    throw "Update failed: price list changed"
}

if ($updated.variant_id -ne $variant.id) {
    throw "Update failed: variant changed"
}

Write-Host "PASS - Price updated"


Write-Host "7. VERIFY DUPLICATE PRICE CONFLICT"

$duplicateRejected = $false

try {
    Invoke-RestMethod `
        -Method Post `
        -Uri $priceBaseUrl `
        -ContentType "application/json" `
        -Body (@{
            price_list_id = $priceList.id
            variant_id = $variant.id
            unit_price = "1050.00"
        } | ConvertTo-Json)

} catch {
    $statusCode = [int]$_.Exception.Response.StatusCode

    if ($statusCode -eq 409) {
        $duplicateRejected = $true
    }
}

if (-not $duplicateRejected) {
    throw "Expected duplicate price to return 409"
}

Write-Host "PASS - Duplicate price rejected"


Write-Host "8. VERIFY NEGATIVE PRICE REJECTED"

$negativeRejected = $false

try {
    Invoke-RestMethod `
        -Method Post `
        -Uri $priceBaseUrl `
        -ContentType "application/json" `
        -Body (@{
            price_list_id = $priceList.id
            variant_id = $variant.id
            unit_price = "-1.00"
        } | ConvertTo-Json)

} catch {
    $statusCode = [int]$_.Exception.Response.StatusCode

    if ($statusCode -eq 400) {
        $negativeRejected = $true
    }
}

if (-not $negativeRejected) {
    throw "Expected negative price to return 400"
}

Write-Host "PASS - Negative price rejected"


Write-Host "9. VERIFY INVALID REFERENCE"

$invalidReferenceRejected = $false

try {
    Invoke-RestMethod `
        -Method Post `
        -Uri $priceBaseUrl `
        -ContentType "application/json" `
        -Body (@{
            price_list_id = ([guid]::NewGuid().ToString())
            variant_id = $variant.id
            unit_price = "500.00"
        } | ConvertTo-Json)

} catch {
    $statusCode = [int]$_.Exception.Response.StatusCode

    if ($statusCode -eq 422) {
        $invalidReferenceRejected = $true
    }
}

if (-not $invalidReferenceRejected) {
    throw "Expected invalid reference to return 422"
}

Write-Host "PASS - Invalid reference rejected"


Write-Host ""
Write-Host "==============================="
Write-Host "ALL PRICE API TESTS PASSED"
Write-Host "==============================="
