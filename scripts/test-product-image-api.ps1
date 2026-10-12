param(
    [string]$BaseUrl = "http://localhost:8081"
)

$ErrorActionPreference = "Stop"

$BaseUrl = $BaseUrl.TrimEnd('/')
$CatalogBaseUrl = "$BaseUrl/api/v1/catalog"

$categoryBaseUrl = "$CatalogBaseUrl/categories"
$productBaseUrl = "$CatalogBaseUrl/products"
$unitBaseUrl = "$CatalogBaseUrl/units"
$variantBaseUrl = "$CatalogBaseUrl/product-variants"
$imageBaseUrl = "$CatalogBaseUrl/product-images"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

Write-Host "0. CREATE IMAGE DEPENDENCIES"

$category = Invoke-RestMethod `
    -Method Post `
    -Uri $categoryBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Image Category $unique"
        code = "IMG-CAT-$unique"
    } | ConvertTo-Json)

$unit = Invoke-RestMethod `
    -Method Post `
    -Uri $unitBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        code = "IU-$unique"
        name = "Image Unit $unique"
    } | ConvertTo-Json)

$product = Invoke-RestMethod `
    -Method Post `
    -Uri $productBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        category_id = $category.id
        name = "Image Product $unique"
        product_code = "IMG-PROD-$unique"
    } | ConvertTo-Json)

$variant = Invoke-RestMethod `
    -Method Post `
    -Uri $variantBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        sku = "IMG-SKU-$unique"
        unit_id = $unit.id
    } | ConvertTo-Json)

Write-Host "PASS - Dependencies created"


Write-Host "1. CREATE PRODUCT IMAGE"

$imageUrl = "https://example.com/products/$unique/main.jpg"

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $imageBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        variant_id = $variant.id
        image_url = $imageUrl
        sort_order = 1
    } | ConvertTo-Json)

$id = $created.id

if (-not $id) {
    throw "Create failed: product image ID missing"
}

if ($created.product_id -ne $product.id) {
    throw "Create failed: wrong product ID"
}

if ($created.variant_id -ne $variant.id) {
    throw "Create failed: wrong variant ID"
}

if ($created.image_url -ne $imageUrl) {
    throw "Create failed: wrong image URL"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET PRODUCT IMAGE"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$imageBaseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: wrong image ID"
}

Write-Host "PASS - Product image fetched"


Write-Host "3. LIST IMAGES BY PRODUCT"

$productImages = Invoke-RestMethod `
    -Method Get `
    -Uri "$productBaseUrl/$($product.id)/images"

$found = $productImages.items | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "Image not found under product"
}

# Also test query filter on regular endpoint
$regularList = Invoke-RestMethod `
    -Method Get `
    -Uri "$imageBaseUrl`?product_id=$($product.id)&variant_id=$($variant.id)"

if (-not ($regularList.items | Where-Object { $_.id -eq $id })) {
    throw "Regular product-images endpoint filter failed"
}

Write-Host "PASS - Product image found under product"


Write-Host "4. LIST IMAGES BY VARIANT"

$variantImages = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl/$($variant.id)/images"

$found = $variantImages.items | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "Image not found under variant"
}

Write-Host "PASS - Product image found under variant"


Write-Host "5. UPDATE PRODUCT IMAGE"

$updatedUrl = "https://example.com/products/$unique/updated.jpg"

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$imageBaseUrl/$id" `
    -ContentType "application/json" `
    -Body (@{
        variant_id = $null
        image_url = $updatedUrl
        sort_order = 5
    } | ConvertTo-Json)

if ($updated.image_url -ne $updatedUrl) {
    throw "Update failed: wrong URL"
}

if ($updated.sort_order -ne 5) {
    throw "Update failed: wrong sort order"
}

if ($null -ne $updated.variant_id) {
    throw "Update failed: variant should be null"
}

Write-Host "PASS - Product image updated"


Write-Host "6. VERIFY VARIANT MUST BELONG TO PRODUCT"

$otherProduct = Invoke-RestMethod `
    -Method Post `
    -Uri $productBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        category_id = $category.id
        name = "Other Image Product $unique"
        product_code = "IMG-OTHER-$unique"
    } | ConvertTo-Json)

$otherVariant = Invoke-RestMethod `
    -Method Post `
    -Uri $variantBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $otherProduct.id
        sku = "IMG-OTHER-SKU-$unique"
        unit_id = $unit.id
    } | ConvertTo-Json)

$rejected = $false

try {
    Invoke-RestMethod `
        -Method Post `
        -Uri $imageBaseUrl `
        -ContentType "application/json" `
        -Body (@{
            product_id = $product.id
            variant_id = $otherVariant.id
            image_url = "https://example.com/wrong.jpg"
            sort_order = 1
        } | ConvertTo-Json)

} catch {
    $statusCode = [int]$_.Exception.Response.StatusCode

    if ($statusCode -eq 422) {
        $rejected = $true
    }
}

if (-not $rejected) {
    throw "Expected mismatched product/variant to return 422"
}

Write-Host "PASS - Mismatched product/variant rejected"


Write-Host ""
Write-Host "======================================"
Write-Host "ALL PRODUCT IMAGE API TESTS PASSED"
Write-Host "======================================"
