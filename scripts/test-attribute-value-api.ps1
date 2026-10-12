param(
    [string]$BaseUrl = "http://localhost:8081"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$BaseUrl = $BaseUrl.TrimEnd('/')
$CatalogBaseUrl = "$BaseUrl/api/v1/catalog"

$attributeBaseUrl = "$CatalogBaseUrl/attributes"
$valueBaseUrl = "$CatalogBaseUrl/attribute-values"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

Write-Host "0. CREATE ATTRIBUTE DEPENDENCY"

$attributeBody = @{
    name = "Value Attribute $unique"
    code = "VAL-ATTR-$unique"
} | ConvertTo-Json

$attribute = Invoke-RestMethod `
    -Method Post `
    -Uri $attributeBaseUrl `
    -ContentType "application/json" `
    -Body $attributeBody

$attributeId = $attribute.id

if (-not $attributeId) {
    throw "Attribute dependency creation failed"
}

Write-Host "PASS - Attribute created"


Write-Host "1. CREATE ATTRIBUTE VALUE"

$initialValue = "Blue $unique"

$createBody = @{
    attribute_id = $attributeId
    value = $initialValue
} | ConvertTo-Json

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $valueBaseUrl `
    -ContentType "application/json" `
    -Body $createBody

$id = $created.id

if (-not $id) {
    throw "Create failed: attribute value ID missing"
}

if ($created.attribute_id -ne $attributeId) {
    throw "Create failed: wrong attribute ID"
}

if ($created.value -ne $initialValue) {
    throw "Create failed: wrong value"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET ATTRIBUTE VALUE"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$valueBaseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: wrong ID"
}

if ($fetched.value -ne $initialValue) {
    throw "Get failed: wrong value"
}

Write-Host "PASS - Attribute value fetched"


Write-Host "3. LIST ATTRIBUTE VALUES"

$values = Invoke-RestMethod `
    -Method Get `
    -Uri $valueBaseUrl

$found = $values | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "List failed: attribute value not found"
}

Write-Host "PASS - Attribute value found in list"


Write-Host "4. LIST VALUES BY ATTRIBUTE"

$attributeValues = Invoke-RestMethod `
    -Method Get `
    -Uri "$attributeBaseUrl/$attributeId/values"

$attributeValueFound = $attributeValues | Where-Object {
    $_.id -eq $id
}

if (-not $attributeValueFound) {
    throw "Attribute value not found under attribute"
}

Write-Host "PASS - Attribute value found under attribute"


Write-Host "5. UPDATE ATTRIBUTE VALUE"

$updatedValue = "Navy Blue $unique"

$updateBody = @{
    value = $updatedValue
} | ConvertTo-Json

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$valueBaseUrl/$id" `
    -ContentType "application/json" `
    -Body $updateBody

if ($updated.value -ne $updatedValue) {
    throw "Update failed: wrong value"
}

if ($updated.attribute_id -ne $attributeId) {
    throw "Update failed: attribute changed unexpectedly"
}

Write-Host "PASS - Attribute value updated"


Write-Host ""
Write-Host "========================================"
Write-Host "ALL ATTRIBUTE VALUE API TESTS PASSED"
Write-Host "========================================"
