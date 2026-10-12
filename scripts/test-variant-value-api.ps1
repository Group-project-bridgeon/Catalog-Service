param(
    [string]$BaseUrl = "http://localhost:8081"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$BaseUrl = $BaseUrl.TrimEnd('/')
$CatalogBaseUrl = "$BaseUrl/api/v1/catalog"

$categoryBaseUrl = "$CatalogBaseUrl/categories"
$productBaseUrl = "$CatalogBaseUrl/products"
$unitBaseUrl = "$CatalogBaseUrl/units"
$variantBaseUrl = "$CatalogBaseUrl/product-variants"
$attributeBaseUrl = "$CatalogBaseUrl/attributes"
$attributeValueBaseUrl = "$CatalogBaseUrl/attribute-values"
$variantValueBaseUrl = "$CatalogBaseUrl/variant-values"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

Write-Host "0. CREATE DEPENDENCIES"

$category = Invoke-RestMethod `
    -Method Post `
    -Uri $categoryBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "VV Category $unique"
        code = "VV-CAT-$unique"
    } | ConvertTo-Json)

$unit = Invoke-RestMethod `
    -Method Post `
    -Uri $unitBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        code = "VVU-$unique"
        name = "VV Unit $unique"
    } | ConvertTo-Json)

$product = Invoke-RestMethod `
    -Method Post `
    -Uri $productBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        category_id = $category.id
        name = "VV Product $unique"
        product_code = "VV-PROD-$unique"
    } | ConvertTo-Json)

$variant = Invoke-RestMethod `
    -Method Post `
    -Uri $variantBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        sku = "VV-SKU-$unique"
        unit_id = $unit.id
    } | ConvertTo-Json)

$attribute = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "VV Color $unique"
        code = "VV-COLOR-$unique"
    } | ConvertTo-Json)

$blue = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeValueBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        attribute_id = $attribute.id
        value = "Blue $unique"
    } | ConvertTo-Json)

$red = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeValueBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        attribute_id = $attribute.id
        value = "Red $unique"
    } | ConvertTo-Json)

Write-Host "PASS - Dependencies created"


Write-Host "1. CREATE VARIANT VALUE"

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $variantValueBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        variant_id = $variant.id
        attribute_id = $attribute.id
        attribute_value_id = $blue.id
    } | ConvertTo-Json)

if ($created.variant_id -ne $variant.id) {
    throw "Create failed: wrong variant ID"
}

if ($created.attribute_value_id -ne $blue.id) {
    throw "Create failed: wrong attribute value ID"
}

Write-Host "PASS - Variant value created"


Write-Host "2. GET VARIANT VALUE"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantValueBaseUrl/$($variant.id)/$($attribute.id)"

if ($fetched.attribute_value_id -ne $blue.id) {
    throw "Get failed"
}

Write-Host "PASS - Variant value fetched"


Write-Host "3. LIST VALUES BY VARIANT"

$values = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantBaseUrl/$($variant.id)/values"

$found = $values.items | Where-Object {
    $_.attribute_id -eq $attribute.id
}

if (-not $found) {
    throw "Variant value not found in list"
}

# Test regular endpoint with filter
$regularValues = Invoke-RestMethod `
    -Method Get `
    -Uri "$variantValueBaseUrl`?variant_id=$($variant.id)"

if (-not ($regularValues.items | Where-Object { $_.attribute_id -eq $attribute.id })) {
    throw "Regular variant-values endpoint filter failed"
}

Write-Host "PASS - Variant value found in list"


Write-Host "4. UPDATE VARIANT VALUE"

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$variantValueBaseUrl/$($variant.id)/$($attribute.id)" `
    -ContentType "application/json" `
    -Body (@{
        attribute_value_id = $red.id
    } | ConvertTo-Json)

if ($updated.attribute_value_id -ne $red.id) {
    throw "Update failed"
}

Write-Host "PASS - Variant value updated"


Write-Host "5. VERIFY ATTRIBUTE/VALUE FK"

$otherAttribute = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "VV Size $unique"
        code = "VV-SIZE-$unique"
    } | ConvertTo-Json)

$large = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeValueBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        attribute_id = $otherAttribute.id
        value = "Large $unique"
    } | ConvertTo-Json)

$invalidReferenceRejected = $false

try {
    Invoke-RestMethod `
        -Method Put `
        -Uri "$variantValueBaseUrl/$($variant.id)/$($attribute.id)" `
        -ContentType "application/json" `
        -Body (@{
            attribute_value_id = $large.id
        } | ConvertTo-Json)

} catch {
    $statusCode = [int]$_.Exception.Response.StatusCode

    if ($statusCode -eq 422) {
        $invalidReferenceRejected = $true
    }
}

if (-not $invalidReferenceRejected) {
    throw "Expected mismatched attribute value to return 422"
}

Write-Host "PASS - Invalid attribute/value relationship rejected"


Write-Host ""
Write-Host "======================================"
Write-Host "ALL VARIANT VALUE API TESTS PASSED"
Write-Host "======================================"
