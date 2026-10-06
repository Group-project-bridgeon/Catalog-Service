$baseUrl = "http://localhost:8081/api/v1/catalog/units"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$code = "UNIT-$unique"

Write-Host "1. CREATE UNIT"

$createBody = @{
    code = $code
    name = "Integration Unit $unique"
} | ConvertTo-Json

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $baseUrl `
    -ContentType "application/json" `
    -Body $createBody

$id = $created.id

if (-not $id) {
    throw "Create failed: no unit ID returned"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET UNIT"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: returned wrong ID"
}

Write-Host "PASS - Unit fetched"


Write-Host "3. LIST UNITS"

$units = Invoke-RestMethod `
    -Method Get `
    -Uri $baseUrl

$found = $units | Where-Object { $_.id -eq $id }

if (-not $found) {
    throw "List failed: created unit not found"
}

Write-Host "PASS - Unit found in list"


Write-Host "4. UPDATE UNIT"

$updateBody = @{
    code = "$code-U"
    name = "Updated Unit $unique"
} | ConvertTo-Json

$updated = Invoke-RestMethod `
    -Method Put `
    -Uri "$baseUrl/$id" `
    -ContentType "application/json" `
    -Body $updateBody

if ($updated.name -ne "Updated Unit $unique") {
    throw "Update failed"
}

Write-Host "PASS - Unit updated"


Write-Host "5. DEACTIVATE UNIT"

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

Write-Host "PASS - Unit deactivated"


Write-Host ""
Write-Host "==============================="
Write-Host "ALL UNIT API TESTS PASSED"
Write-Host "==============================="
