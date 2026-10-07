$baseUrl = "http://localhost:8081/api/v1/catalog/attributes"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

$name = "Attribute Integration $unique"
$code = "ATTR-$unique"

Write-Host "1. CREATE ATTRIBUTE"

$createBody = @{
    name = $name
    code = $code
} | ConvertTo-Json

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $baseUrl `
    -ContentType "application/json" `
    -Body $createBody

$id = $created.id

if (-not $id) {
    throw "Create failed: attribute ID missing"
}

if ($created.name -ne $name) {
    throw "Create failed: wrong name"
}

if ($created.code -ne $code) {
    throw "Create failed: wrong code"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET ATTRIBUTE"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: wrong attribute ID"
}

if ($fetched.code -ne $code) {
    throw "Get failed: wrong attribute code"
}

Write-Host "PASS - Attribute fetched"


Write-Host "3. LIST ATTRIBUTES"

$attributes = Invoke-RestMethod `
    -Method Get `
    -Uri $baseUrl

$found = $attributes | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "List failed: created attribute not found"
}

Write-Host "PASS - Attribute found in list"


Write-Host "4. UPDATE ATTRIBUTE"

$updatedName = "Updated Attribute $unique"
$updatedCode = "ATTR-UPD-$unique"

$updateBody = @{
    name = $updatedName
    code = $updatedCode
} | ConvertTo-Json

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$baseUrl/$id" `
    -ContentType "application/json" `
    -Body $updateBody

if ($updated.name -ne $updatedName) {
    throw "Update failed: wrong name"
}

if ($updated.code -ne $updatedCode) {
    throw "Update failed: wrong code"
}

Write-Host "PASS - Attribute updated"


Write-Host ""
Write-Host "=================================="
Write-Host "ALL ATTRIBUTE API TESTS PASSED"
Write-Host "=================================="
