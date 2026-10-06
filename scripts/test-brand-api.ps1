$baseUrl = "http://localhost:8081/api/v1/catalog/brands"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$code = "BRAND-INT-$unique"

Write-Host "1. CREATE BRAND"

$createBody = @{
    name = "Integration Brand $unique"
    code = $code
} | ConvertTo-Json

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $baseUrl `
    -ContentType "application/json" `
    -Body $createBody

$id = $created.id

if (-not $id) {
    throw "Create failed: no brand ID returned"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET BRAND"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: returned wrong ID"
}

Write-Host "PASS - Brand fetched"


Write-Host "3. LIST BRANDS"

$brands = Invoke-RestMethod `
    -Method Get `
    -Uri $baseUrl

$found = $brands | Where-Object { $_.id -eq $id }

if (-not $found) {
    throw "List failed: created brand not found"
}

Write-Host "PASS - Brand found in list"


Write-Host "4. UPDATE BRAND"

$updateBody = @{
    name = "Updated Integration Brand $unique"
    code = "$code-UPD"
} | ConvertTo-Json

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$baseUrl/$id" `
    -ContentType "application/json" `
    -Body $updateBody

if ($updated.name -ne "Updated Integration Brand $unique") {
    throw "Update failed"
}

Write-Host "PASS - Brand updated"


Write-Host "5. DEACTIVATE BRAND"

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

Write-Host "PASS - Brand deactivated"


Write-Host ""
Write-Host "==============================="
Write-Host "ALL BRAND API TESTS PASSED"
Write-Host "==============================="