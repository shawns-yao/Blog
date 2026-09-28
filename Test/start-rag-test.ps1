param([string]$Database = 'rag_test', [switch]$RerankUnavailable)
$ErrorActionPreference = 'Stop'
if ($Database -notmatch '^rag_test(_[a-z0-9]+)?$') { throw 'Invalid test database name' }
$taskRoot = Split-Path -Parent $PSScriptRoot
$runtime = Join-Path $taskRoot 'Temp/rag-runtime'
New-Item -ItemType Directory -Force -Path $runtime | Out-Null
$env:APP_ENV = 'test'
$env:APP_PORT = '18089'
$env:DB_DRIVER = 'postgres'
$env:DB_DSN = 'postgres://postgres@127.0.0.1:15439/' + $Database + '?sslmode=disable'
$env:REDIS_ADDR = '127.0.0.1:16389'
$env:REDIS_PASSWORD = ''
$env:REDIS_DB = '0'
$env:REDIS_PREFIX = 'rag-test:'
$env:AUTH_SECRET = [guid]::NewGuid().ToString('N') + [guid]::NewGuid().ToString('N')
$env:HTMLSNAPSHOT_BASE_URL = 'http://127.0.0.1:1'
$env:GEOIP_DB_URL = 'http://127.0.0.1:1'
$env:GEOIP_ASN_URL = 'http://127.0.0.1:1'
$env:APP_UPDATE_CHECK_ENABLED = 'false'
$env:RAG_ENABLED = 'true'
# Exercise a real connection failure and the configured official fallback.
$env:RAG_CHAT_PRIMARY_BASE_URL = 'http://127.0.0.1:1/v1'
$env:RAG_CHAT_PRIMARY_API_KEY = 'rag-test-unavailable'
$env:RAG_CHAT_PRIMARY_TIMEOUT = '1s'
if ($RerankUnavailable) {
  $env:RAG_RERANK_BASE_URL = 'http://127.0.0.1:1/v1'
  $env:RAG_RERANK_API_KEY = 'rag-test-unavailable'
  $env:RAG_RERANK_TIMEOUT = '1s'
}
Get-ChildItem Env: | Where-Object { $_.Name -match '^(MEDIA_R2_|S3_API_|LEOSTUDIO_R2_|Access_Key_ID$|Secret_Access_Key$)' } | ForEach-Object {
  [Environment]::SetEnvironmentVariable($_.Name, '', 'Process')
}
$process = Start-Process -FilePath (Join-Path $taskRoot 'Temp/rag-test-api-next.exe') -WorkingDirectory $runtime -WindowStyle Hidden -RedirectStandardOutput (Join-Path $runtime 'stdout.log') -RedirectStandardError (Join-Path $runtime 'stderr.log') -PassThru
$process.Id | Set-Content -LiteralPath (Join-Path $taskRoot 'Temp/rag-test-api.pid')
Write-Output ('Test server PID: ' + $process.Id)
