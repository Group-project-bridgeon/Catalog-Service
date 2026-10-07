$baseUrl = "http://localhost:8081/api/v1/catalog/price-lists"

$unique = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()

$name = "Wholesale Price List $unique"
$code = "WHOLESALE-$unique"

Write-Host "1. CREATE PRICE LIST"

$createBody = @{
    name = $name
    code = $code
    currency_code = "INR"
} | ConvertTo-Json

$created = Invoke-RestMethod `
    -Method Post `
    -Uri $baseUrl `
    -ContentType "application/json" `
    -Body $createBody

$id = $created.id

if (-not $id) {
    throw "Create failed: price list ID missing"
}

if ($created.name -ne $name) {
    throw "Create failed: wrong name"
}

if ($created.code -ne $code) {
    throw "Create failed: wrong code"
}

if ($created.currency_code -ne "INR") {
    throw "Create failed: wrong currency code"
}

if ($created.is_active -ne $true) {
    throw "Create failed: price list should be active"
}

Write-Host "PASS - Created $id"


Write-Host "2. GET PRICE LIST"

$fetched = Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/$id"

if ($fetched.id -ne $id) {
    throw "Get failed: wrong price list ID"
}

Write-Host "PASS - Price list fetched"


Write-Host "3. LIST PRICE LISTS"

$priceLists = Invoke-RestMethod `
    -Method Get `
    -Uri $baseUrl

$found = $priceLists | Where-Object {
    $_.id -eq $id
}

if (-not $found) {
    throw "List failed: price list not found"
}

Write-Host "PASS - Price list found in list"


Write-Host "4. UPDATE PRICE LIST"

$updatedName = "Updated Wholesale $unique"
$updatedCode = "WHOLESALE-UPD-$unique"

$updateBody = @{
    name = $updatedName
    code = $updatedCode
    currency_code = "USD"
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

if ($updated.currency_code -ne "USD") {
    throw "Update failed: wrong currency code"
}

Write-Host "PASS - Price list updated"


Write-Host "5. DEACTIVATE PRICE LIST"

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

Write-Host "PASS - Price list deactivated"


Write-Host "6. VERIFY DUPLICATE CODE CONFLICT"

$duplicateRejected = $false

try {
    Invoke-RestMethod `
        -Method Post `
        -Uri $baseUrl `
        -ContentType "application/json" `
        -Body (@{
            name = "Duplicate Price List $unique"
            code = $updatedCode
            currency_code = "INR"
        } | ConvertTo-Json)

} catch {
    $statusCode = [int]$_.Exception.Response.StatusCode

    if ($statusCode -eq 409) {
        $duplicateRejected = $true
    }
}

if (-not $duplicateRejected) {
    throw "Expected duplicate price list code to return 409"
}

Write-Host "PASS - Duplicate code rejected"


Write-Host ""
Write-Host "===================================="
Write-Host "ALL PRICE LIST API TESTS PASSED"
Write-Host "===================================="
