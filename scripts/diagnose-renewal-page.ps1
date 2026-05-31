param(
  [string]$FrontendBaseUrl = "http://127.0.0.1",
  [string]$BackendBaseUrl = "http://127.0.0.1:8080",
  [string]$BackendPort = "8080",
  [string]$MySqlHost = "127.0.0.1",
  [int]$MySqlPort = 3306,
  [string]$LogPath = "",
  [int]$TimeoutSec = 5
)

$ErrorActionPreference = "Continue"

function Write-Section {
  param([string]$Title)
  Write-Host ""
  Write-Host "=== $Title ==="
}

function Redact-Dsn {
  param([string]$Value)
  if ([string]::IsNullOrWhiteSpace($Value)) {
    return ""
  }
  $redacted = $Value -replace ':(.*?)@tcp\(', ':***@tcp('
  $redacted = $redacted -replace 'password=[^;&\s]+', 'password=***'
  $redacted = $redacted -replace 'pwd=[^;&\s]+', 'pwd=***'
  return $redacted
}

function Test-Api {
  param(
    [string]$BaseUrl,
    [string]$Path
  )
  $url = $BaseUrl.TrimEnd("/") + $Path
  try {
    $resp = Invoke-WebRequest $url -UseBasicParsing -TimeoutSec $TimeoutSec
    $status = [int]$resp.StatusCode
    $body = [string]$resp.Content
    $apiCode = ""
    $apiMessage = ""
    try {
      $json = $body | ConvertFrom-Json
      if ($null -ne $json.code) {
        $apiCode = [string]$json.code
      }
      if ($null -ne $json.message) {
        $apiMessage = [string]$json.message
      }
    } catch {
      $apiMessage = ($body -replace '\s+', ' ').Trim()
      if ($apiMessage.Length -gt 120) {
        $apiMessage = $apiMessage.Substring(0, 120)
      }
    }
    return [PSCustomObject]@{
      Url = $url
      Ok = $true
      Status = $status
      ApiCode = $apiCode
      Message = $apiMessage
      Error = ""
    }
  } catch {
    $status = ""
    $message = $_.Exception.Message
    if ($_.Exception.Response) {
      try {
        $status = [int]$_.Exception.Response.StatusCode
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $body = $reader.ReadToEnd()
        if (![string]::IsNullOrWhiteSpace($body)) {
          $message = ($body -replace '\s+', ' ').Trim()
          if ($message.Length -gt 120) {
            $message = $message.Substring(0, 120)
          }
        }
      } catch {}
    }
    return [PSCustomObject]@{
      Url = $url
      Ok = $false
      Status = $status
      ApiCode = ""
      Message = ""
      Error = $message
    }
  }
}

function Print-Api-Result {
  param([object]$Result)
  $state = if ($Result.Ok -and ($Result.ApiCode -eq "" -or $Result.ApiCode -eq "0")) { "OK" } else { "FAIL" }
  $detail = if ($Result.Error) { $Result.Error } else { $Result.Message }
  Write-Host ("{0} status={1} api_code={2} path={3}" -f $state, $Result.Status, $Result.ApiCode, ([Uri]$Result.Url).PathAndQuery)
  if (![string]::IsNullOrWhiteSpace($detail) -and $detail -ne "ok") {
    Write-Host ("  detail: {0}" -f $detail)
  }
}

$paths = @(
  "/api/v1/renewals/settings",
  "/api/v1/renewals/plans",
  "/api/v1/servers",
  "/api/v1/host-packages",
  "/api/v1/renewals/unit-prices"
)

Write-Host "Computility Ops renewal page diagnostics"
Write-Host ("Time: {0}" -f (Get-Date -Format "yyyy-MM-dd HH:mm:ss zzz"))

Write-Section "Stage 1 - frontend proxied APIs"
$frontendResults = @()
foreach ($path in $paths) {
  $result = Test-Api -BaseUrl $FrontendBaseUrl -Path $path
  $frontendResults += $result
  Print-Api-Result $result
}

Write-Section "Stage 2 - backend direct APIs"
$backendResults = @()
foreach ($path in $paths) {
  $result = Test-Api -BaseUrl $BackendBaseUrl -Path $path
  $backendResults += $result
  Print-Api-Result $result
}

Write-Section "Stage 3 - backend listening port"
try {
  $listeners = Get-NetTCPConnection -LocalPort ([int]$BackendPort) -State Listen -ErrorAction Stop
  if ($listeners.Count -gt 0) {
    foreach ($item in $listeners) {
      Write-Host ("OK LISTEN local={0}:{1} pid={2}" -f $item.LocalAddress, $item.LocalPort, $item.OwningProcess)
    }
  } else {
    Write-Host ("FAIL no LISTEN socket on port {0}" -f $BackendPort)
  }
} catch {
  Write-Host ("FAIL cannot query port {0}: {1}" -f $BackendPort, $_.Exception.Message)
}

Write-Section "Stage 4 - MySQL TCP"
try {
  $mysql = Test-NetConnection $MySqlHost -Port $MySqlPort -WarningAction SilentlyContinue
  Write-Host ("TcpTestSucceeded={0} target={1}:{2}" -f $mysql.TcpTestSucceeded, $MySqlHost, $MySqlPort)
} catch {
  Write-Host ("FAIL MySQL TCP test error: {0}" -f $_.Exception.Message)
}

Write-Section "Stage 5 - related processes"
try {
  $procs = Get-Process | Where-Object { $_.ProcessName -match "computility|server|backend|go|nginx|caddy|iisexpress|mysql" }
  if ($procs.Count -gt 0) {
    $procs | Select-Object Id, ProcessName, Path | Format-Table -AutoSize
  } else {
    Write-Host "WARN no obvious backend/frontend/mysql process matched common names"
  }
} catch {
  Write-Host ("WARN process query failed: {0}" -f $_.Exception.Message)
}

Write-Section "Stage 6 - environment hints"
Write-Host ("STORAGE_DRIVER={0}" -f $env:STORAGE_DRIVER)
Write-Host ("MYSQL_DSN={0}" -f (Redact-Dsn $env:MYSQL_DSN))

if (![string]::IsNullOrWhiteSpace($LogPath)) {
  Write-Section "Stage 7 - log keyword scan"
  if (Test-Path $LogPath) {
    $keywords = "panic|mysql|Access denied|connection refused|Unknown database|no such host|bind|listen|502|500"
    try {
      Get-ChildItem $LogPath -File -Recurse -ErrorAction SilentlyContinue |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 10 |
        ForEach-Object {
          $matches = Select-String -Path $_.FullName -Pattern $keywords -CaseSensitive:$false -ErrorAction SilentlyContinue | Select-Object -Last 5
          if ($matches) {
            Write-Host ("LOG {0}" -f $_.FullName)
            $matches | ForEach-Object {
              $line = ($_.Line -replace '\s+', ' ').Trim()
              if ($line.Length -gt 180) {
                $line = $line.Substring(0, 180)
              }
              Write-Host ("  {0}: {1}" -f $_.LineNumber, $line)
            }
          }
        }
    } catch {
      Write-Host ("WARN log scan failed: {0}" -f $_.Exception.Message)
    }
  } else {
    Write-Host ("WARN log path not found: {0}" -f $LogPath)
  }
}

Write-Section "Summary"
$frontendFail = @($frontendResults | Where-Object { -not $_.Ok -or ($_.ApiCode -ne "" -and $_.ApiCode -ne "0") }).Count
$backendFail = @($backendResults | Where-Object { -not $_.Ok -or ($_.ApiCode -ne "" -and $_.ApiCode -ne "0") }).Count
if ($backendFail -eq 0 -and $frontendFail -gt 0) {
  Write-Host "LIKELY frontend proxy/static deployment problem: backend direct APIs are OK, proxied APIs failed."
} elseif ($backendFail -gt 0) {
  Write-Host "LIKELY backend or database problem: backend direct APIs failed."
} elseif ($frontendFail -eq 0 -and $backendFail -eq 0) {
  Write-Host "APIs look OK from this host. If browser still fails, check browser Network tab for actual host/base URL."
} else {
  Write-Host "Mixed result. Share Stage 1/2 status lines only."
}
