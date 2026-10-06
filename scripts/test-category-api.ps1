
# Integration test script
$baseUrl = "http://localhost:8081/api/v1/catalog/categories"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$code = "CAT-INT-$unique"

Write-Host "1. CREATE CATEGORY"

$createBody = @{
    name = "Integration Category"
    code = $code
    description = "Automated integration test"
} | ConvertTo-Json

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $baseUrl `
    -ContentType "application/json" `
    -Body $createBody

$id = $created.id

if (-not $id) {
    throw "Create failed: no category ID returned"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET CATEGORY"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: returned wrong ID"
}

Write-Host "PASS - Category fetched"


Write-Host "3. LIST CATEGORIES"

$categories = Invoke-RestMethod `
    -Method Get `
    -Uri $baseUrl

$found = $categories | Where-Object { $_.id -eq $id }

if (-not $found) {
    throw "List failed: created category not found"
}

Write-Host "PASS - Category found in list"


Write-Host "4. UPDATE CATEGORY"

$updateBody = @{
    name = "Updated Integration Category"
    code = "$code-UPD"
    description = "Updated by integration test"
} | ConvertTo-Json

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$baseUrl/$id" `
    -ContentType "application/json" `
    -Body $updateBody

if ($updated.name -ne "Updated Integration Category") {
    throw "Update failed"
}

Write-Host "PASS - Category updated"


Write-Host "5. DEACTIVATE CATEGORY"

$statusBody = @{
    is_active = $false
} | ConvertTo-Json

$deactivated = Invoke-RestMethod `
    -Method Patch `
    -Uri "$baseUrl/$id/status" `
    -ContentType "application/json" `
    -Body $statusBody

if ($deactivated.is_active -ne $false) {
    throw "Deactivate failed"
}

Write-Host "PASS - Category deactivated"


Write-Host ""
Write-Host "================================"
Write-Host "ALL CATEGORY API TESTS PASSED"
Write-Host "================================"