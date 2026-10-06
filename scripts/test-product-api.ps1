$productBaseUrl = "http://localhost:8081/api/v1/catalog/products"
$categoryBaseUrl = "http://localhost:8081/api/v1/catalog/categories"
$brandBaseUrl = "http://localhost:8081/api/v1/catalog/brands"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

$categoryCode = "PROD-CAT-$unique"
$brandCode = "PROD-BRAND-$unique"
$productCode = "PROD-INT-$unique"

Write-Host "0. CREATE PRODUCT DEPENDENCIES"

$categoryBody = @{
    name = "Product Test Category $unique"
    code = $categoryCode
    description = "Category created for product integration test"
} | ConvertTo-Json

$category = Invoke-RestMethod `
    -Method Post `
    -Uri $categoryBaseUrl `
    -ContentType "application/json" `
    -Body $categoryBody

$categoryId = $category.id

if (-not $categoryId) {
    throw "Dependency setup failed: category ID missing"
}

$brandBody = @{
    name = "Product Test Brand $unique"
    code = $brandCode
} | ConvertTo-Json

$brand = Invoke-RestMethod `
    -Method Post `
    -Uri $brandBaseUrl `
    -ContentType "application/json" `
    -Body $brandBody

$brandId = $brand.id

if (-not $brandId) {
    throw "Dependency setup failed: brand ID missing"
}

Write-Host "PASS - Category and Brand created"


Write-Host "1. CREATE PRODUCT"

$description = "Product integration test"

$createBody = @{
    category_id = $categoryId
    brand_id = $brandId
    name = "Integration Product $unique"
    product_code = $productCode
    description = $description
} | ConvertTo-Json

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $productBaseUrl `
    -ContentType "application/json" `
    -Body $createBody

$id = $created.id

if (-not $id) {
    throw "Create failed: product ID missing"
}

if ($created.category_id -ne $categoryId) {
    throw "Create failed: wrong category ID"
}

if ($created.brand_id -ne $brandId) {
    throw "Create failed: wrong brand ID"
}

if ($created.is_active -ne $true) {
    throw "Create failed: product should be active"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET PRODUCT"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$productBaseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: returned wrong product ID"
}

if ($fetched.product_code -ne $productCode) {
    throw "Get failed: returned wrong product code"
}

Write-Host "PASS - Product fetched"


Write-Host "3. LIST PRODUCTS"

$products = Invoke-RestMethod `
    -Method Get `
    -Uri $productBaseUrl

$found = $products | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "List failed: created product not found"
}

Write-Host "PASS - Product found in list"


Write-Host "4. UPDATE PRODUCT"

$updatedCode = "$productCode-UPD"
$updatedName = "Updated Integration Product $unique"

$updateBody = @{
    category_id = $categoryId
    brand_id = $brandId
    name = $updatedName
    product_code = $updatedCode
    description = "Updated product integration test"
} | ConvertTo-Json

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$productBaseUrl/$id" `
    -ContentType "application/json" `
    -Body $updateBody

if ($updated.name -ne $updatedName) {
    throw "Update failed: wrong name"
}

if ($updated.product_code -ne $updatedCode) {
    throw "Update failed: wrong product code"
}

Write-Host "PASS - Product updated"


Write-Host "5. DEACTIVATE PRODUCT"

$statusBody = @{
    is_active = $false
} | ConvertTo-Json

$deactivated = Invoke-RestMethod `
    -Method Patch `
    -Uri "$productBaseUrl/$id/status" `
    -ContentType "application/json" `
    -Body $statusBody

if ($deactivated.is_active -ne $false) {
    throw "Deactivate failed"
}

Write-Host "PASS - Product deactivated"


Write-Host ""
Write-Host "================================="
Write-Host "ALL PRODUCT API TESTS PASSED"
Write-Host "================================="
