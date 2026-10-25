#Requires -Version 7.4
[CmdletBinding()]
param()
$ErrorActionPreference='Stop'
$root=Split-Path $PSScriptRoot -Parent
$run='catalog-stock-test-'+[guid]::NewGuid().ToString('N')
$temporary=Join-Path $root ('.'+$run)
$containerID=$null
$attempted=$false
$names=@('POSTGRES_PASSWORD','PGPASSWORD','CATALOG_STOCK_TEST_DATABASE_URL','CATALOG_STOCK_TEST_RUN_ID','CATALOG_STOCK_TEST_PORT','GOTMPDIR')
$previous=@{}
foreach($name in $names){$previous[$name]=[Environment]::GetEnvironmentVariable($name,'Process')}
function FixtureDocker([string[]]$Arguments){
    $output=& docker @Arguments 2>&1
    if($LASTEXITCODE -ne 0){throw "Catalog fixture Docker operation failed ($($Arguments[0])); raw output suppressed"}
    return $output
}
function Ownership {
    if(!$containerID -or $containerID -notmatch '^[0-9a-f]{64}$'){throw 'Missing owned container ID'}
    $label=([string](FixtureDocker @('inspect','--format','{{index .Config.Labels "catalog.stock.fixture"}}',$containerID))).Trim()
    if($label -ne $run){throw 'Refusing non-owned container operation'}
}
function SQL([string]$query){Ownership;$result=FixtureDocker @('exec',$containerID,'psql','-X','-q','-t','-A','-v','ON_ERROR_STOP=1','-U','catalog_stock_fixture','-d','catalog_stock_test','-c',$query);return ([string]::Join("`n",[string[]]@($result))).Trim()}
function AssertSQL([string]$query,[string]$expected,[string]$label){if((SQL $query) -ne $expected){throw "Catalog fixture assertion failed: $label"};Write-Host "PASS: $label"}
function Migrate([string[]]$op){Ownership;$migrationPath=Join-Path $root 'db/migrations';FixtureDocker (@('run','--rm','--network',"container:$containerID",'-e','PGPASSWORD','-v',"${migrationPath}:/migrations:ro",'migrate/migrate:v4.18.3','-path=/migrations','-database=postgres://catalog_stock_fixture@127.0.0.1:5432/catalog_stock_test?sslmode=disable')+$op)|Out-Null}
try {
    $null=New-Item -ItemType Directory -Path $temporary
    $env:GOTMPDIR=$temporary
    $bytes=[byte[]]::new(32);[Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    $env:POSTGRES_PASSWORD=[Convert]::ToHexString($bytes).ToLowerInvariant();$env:PGPASSWORD=$env:POSTGRES_PASSWORD
    $attempted=$true
    $output=@(FixtureDocker @('run','--rm','-d','--name',$run,'--label',"catalog.stock.fixture=$run",'--tmpfs','/var/lib/postgresql/data','-e','POSTGRES_USER=catalog_stock_fixture','-e','POSTGRES_PASSWORD','-e','POSTGRES_DB=catalog_stock_test','-p','127.0.0.1::5432','postgres:17-alpine'))
    $ids=@($output|ForEach-Object{([string]$_).Trim()}|Where-Object{$_ -match '^[0-9a-f]{64}$'})
    if($ids.Count -ne 1){throw 'Invalid fixture identity output'};$containerID=$ids[0];Ownership
    $mapping=([string](FixtureDocker @('port',$containerID,'5432/tcp'))).Trim();if($mapping -notmatch '^127\.0\.0\.1:(\d+)$'){throw 'Invalid fixture binding'};$port=$Matches[1]
    $ready=$false
    for($i=0;$i -lt 60;$i++){try{if((SQL 'SELECT 1') -eq '1'){$ready=$true;break}}catch{};Start-Sleep -Seconds 1}
    if(!$ready){throw 'Catalog fixture readiness failed'}
    SQL "ALTER DATABASE catalog_stock_test SET catalog.fixture_id='$run'"|Out-Null
    Migrate @('up','2')
    AssertSQL "SELECT version::text||':'||dirty::text FROM schema_migrations" '2:false' 'legacy Catalog baseline clean'
    SQL "INSERT INTO catalog_categories(id,name,code) VALUES ('01935f00-0000-7000-8000-000000000001','Preserved category','SLICE4A'); INSERT INTO catalog_units(id,code,name) VALUES ('01935f00-0000-7000-8000-000000000002','LEGACYBOX','Legacy box, unclassified'); INSERT INTO catalog_products(id,category_id,name,product_code) VALUES ('01935f00-0000-7000-8000-000000000003','01935f00-0000-7000-8000-000000000001','Preserved legacy product','SLICE4A'); INSERT INTO catalog_product_variants(id,product_id,unit_id,sku) VALUES ('01935f00-0000-7000-8000-000000000004','01935f00-0000-7000-8000-000000000003','01935f00-0000-7000-8000-000000000002','LEGACY-SKU')"|Out-Null
    Migrate @('up','1')
    AssertSQL "SELECT version::text||':'||dirty::text FROM schema_migrations" '3:false' 'new migration clean'
    AssertSQL "SELECT (SELECT count(*) FROM catalog_piece_units)+(SELECT count(*) FROM catalog_stock_eligibility)+(SELECT count(*) FROM catalog_variant_packs)+(SELECT count(*) FROM catalog_pack_revisions)" '0' 'no fabricated eligibility or packs'
    Migrate @('down','1')
    AssertSQL "SELECT count(*) FROM catalog_product_variants WHERE sku='LEGACY-SKU'" '1' 'legacy row survives empty rollback'
    AssertSQL "SELECT count(*) FROM pg_proc WHERE pronamespace='public'::regnamespace AND (proname LIKE 'catalog_stock_%' OR proname LIKE 'catalog_pack_%')" '0' 'new functions removed'
    Migrate @('up','1')
    $env:CATALOG_STOCK_TEST_DATABASE_URL="postgres://catalog_stock_fixture:$($env:POSTGRES_PASSWORD)@127.0.0.1:$port/catalog_stock_test?sslmode=disable"
    $env:CATALOG_STOCK_TEST_RUN_ID=$run;$env:CATALOG_STOCK_TEST_PORT=$port
    & go -C $root test -count=1 -timeout=180s -v ./internal/repository/postgres -run '^TestCatalogStock'
    if($LASTEXITCODE -ne 0){throw 'Catalog stock tests failed; owned cleanup will run'}
    Migrate @('down','1')
    AssertSQL "SELECT version::text||':'||dirty::text FROM schema_migrations" '2:false' 'populated metadata rollback clean'
    AssertSQL "SELECT count(*) FROM catalog_product_variants v JOIN catalog_units u ON u.id=v.unit_id JOIN catalog_products p ON p.id=v.product_id WHERE v.sku='LEGACY-SKU' AND u.code='LEGACYBOX' AND p.name='Preserved legacy product'" '1' 'legacy identities/labels preserved'
    AssertSQL "SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('catalog_piece_units','catalog_stock_eligibility','catalog_variant_packs','catalog_pack_revisions')" '0' 'new tables removed'
    AssertSQL "SELECT count(*) FROM pg_proc WHERE pronamespace='public'::regnamespace AND (proname LIKE 'catalog_stock_%' OR proname LIKE 'catalog_pack_%')" '0' 'populated rollback function cleanup'
    Migrate @('up','1')
    AssertSQL "SELECT version::text||':'||dirty::text FROM schema_migrations" '3:false' 'reapplied metadata clean'
    AssertSQL "SELECT count(*) FROM catalog_stock_eligibility" '0' 'reapply remains opt-in'
    Write-Host 'PASS: isolated Catalog stock-reference verification; synthetic fixtures only.'
} finally {
    try {
        if($attempted -and !$containerID){$owned=@(FixtureDocker @('ps','-a','--no-trunc','--filter',"label=catalog.stock.fixture=$run",'--format','{{.ID}}'));if($owned.Count -gt 1){throw 'Ambiguous owned cleanup'};if($owned.Count -eq 1){$containerID=([string]$owned[0]).Trim()}}
        if($containerID){Ownership;FixtureDocker @('rm','-f',$containerID)|Out-Null;if(FixtureDocker @('ps','-a','--filter',"label=catalog.stock.fixture=$run",'--format','{{.ID}}')){throw 'Fixture cleanup incomplete'};Write-Host 'PASS: owned Catalog container removed; tmpfs discarded.'}
    } finally {
        foreach($name in $names){[Environment]::SetEnvironmentVariable($name,$previous[$name],'Process')}
        if(Test-Path -LiteralPath $temporary){$resolved=(Resolve-Path -LiteralPath $temporary).Path;$expected=[IO.Path]::GetFullPath((Join-Path $root ('.'+$run)));if($resolved -ne $expected -or !(Split-Path $resolved -Parent).Equals($root,[StringComparison]::OrdinalIgnoreCase)){throw 'Refusing unsafe fixture temp cleanup'};Remove-Item -LiteralPath $resolved -Recurse -Force}
        Write-Host 'PASS: owned temporary files removed; previous environment restored.'
    }
}
