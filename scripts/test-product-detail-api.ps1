$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$categoryBaseUrl = "http://localhost:8081/api/v1/catalog/categories"
$unitBaseUrl = "http://localhost:8081/api/v1/catalog/units"
$productBaseUrl = "http://localhost:8081/api/v1/catalog/products"
$variantBaseUrl = "http://localhost:8081/api/v1/catalog/product-variants"
$attributeBaseUrl = "http://localhost:8081/api/v1/catalog/attributes"
$attributeValueBaseUrl = "http://localhost:8081/api/v1/catalog/attribute-values"
$variantValueBaseUrl = "http://localhost:8081/api/v1/catalog/variant-values"
$imageBaseUrl = "http://localhost:8081/api/v1/catalog/product-images"
$priceListBaseUrl = "http://localhost:8081/api/v1/catalog/price-lists"
$priceBaseUrl = "http://localhost:8081/api/v1/catalog/prices"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

Write-Host "0. CREATE PRODUCT DETAIL DEPENDENCIES"

# Category
$category = Invoke-RestMethod `
    -Method Post `
    -Uri $categoryBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Detail Cat $unique"
        code = "DCAT-$unique"
    } | ConvertTo-Json)

# Unit
$unit = Invoke-RestMethod `
    -Method Post `
    -Uri $unitBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Piece $unique"
        code = "DUNT-$unique"
    } | ConvertTo-Json)

# Product
$product = Invoke-RestMethod `
    -Method Post `
    -Uri $productBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        category_id = $category.id
        name = "Detail Product $unique"
        product_code = "DPROD-$unique"
    } | ConvertTo-Json)

# 2 Product Variants
$v1 = Invoke-RestMethod `
    -Method Post `
    -Uri $variantBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        sku = "DSKU-$unique-1"
        default_selling_price = "100.00"
        unit_id = $unit.id
    } | ConvertTo-Json)

$v2 = Invoke-RestMethod `
    -Method Post `
    -Uri $variantBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        sku = "DSKU-$unique-2"
        default_selling_price = "200.00"
        unit_id = $unit.id
    } | ConvertTo-Json)

# Attribute & Attribute Values
$attribute = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Detail Color $unique"
        code = "DCOL-$unique"
    } | ConvertTo-Json)

$valRed = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeValueBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        attribute_id = $attribute.id
        value = "Red $unique"
    } | ConvertTo-Json)

$valBlue = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeValueBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        attribute_id = $attribute.id
        value = "Blue $unique"
    } | ConvertTo-Json)

# Variant Values
$vv1 = Invoke-RestMethod `
    -Method Post `
    -Uri $variantValueBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        variant_id = $v1.id
        attribute_id = $attribute.id
        attribute_value_id = $valRed.id
    } | ConvertTo-Json)

$vv2 = Invoke-RestMethod `
    -Method Post `
    -Uri $variantValueBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        variant_id = $v2.id
        attribute_id = $attribute.id
        attribute_value_id = $valBlue.id
    } | ConvertTo-Json)

# Product Images
$img1 = Invoke-RestMethod `
    -Method Post `
    -Uri $imageBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        variant_id = $v1.id
        image_url = "https://example.com/d1.jpg"
        sort_order = 1
    } | ConvertTo-Json)

$img2 = Invoke-RestMethod `
    -Method Post `
    -Uri $imageBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        product_id = $product.id
        variant_id = $v2.id
        image_url = "https://example.com/d2.jpg"
        sort_order = 2
    } | ConvertTo-Json)

# Price List & Prices
$priceList = Invoke-RestMethod `
    -Method Post `
    -Uri $priceListBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        name = "Detail Price List $unique"
        code = "DPL-$unique"
        currency_code = "INR"
    } | ConvertTo-Json)

$price1 = Invoke-RestMethod `
    -Method Post `
    -Uri $priceBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        price_list_id = $priceList.id
        variant_id = $v1.id
        unit_price = "95.00"
    } | ConvertTo-Json)

$price2 = Invoke-RestMethod `
    -Method Post `
    -Uri $priceBaseUrl `
    -ContentType "application/json" `
    -Body (@{
        price_list_id = $priceList.id
        variant_id = $v2.id
        unit_price = "190.00"
    } | ConvertTo-Json)

Write-Host "PASS - All dependencies and associated records created"


Write-Host "1. GET PRODUCT DETAIL"

$detail = Invoke-RestMethod `
    -Method Get `
    -Uri "$productBaseUrl/$($product.id)/detail"

if ($detail.product.id -ne $product.id) {
    throw "Product detail verification failed: wrong product.id"
}

if ($detail.variants.Count -ne 2) {
    throw "Product detail verification failed: expected 2 variants, got $($detail.variants.Count)"
}

if ($detail.variant_values.Count -ne 2) {
    throw "Product detail verification failed: expected 2 variant values, got $($detail.variant_values.Count)"
}

if ($detail.images.Count -ne 2) {
    throw "Product detail verification failed: expected 2 images, got $($detail.images.Count)"
}

if ($detail.prices.Count -ne 2) {
    throw "Product detail verification failed: expected 2 prices, got $($detail.prices.Count)"
}

Write-Host "PASS - Product detail structure verified"


Write-Host ""
Write-Host "=========================================="
Write-Host "ALL PRODUCT DETAIL API TESTS PASSED"
Write-Host "=========================================="
