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
    $dataTotal = ""
    $dataListCount = ""
    $contentPreview = ($body -replace '\s+', ' ').Trim()
    if ($contentPreview.Length -gt 160) {
      $contentPreview = $contentPreview.Substring(0, 160)
    }
    try {
      $json = $body | ConvertFrom-Json
      if ($null -ne $json.code) {
        $apiCode = [string]$json.code
      }
      if ($null -ne $json.message) {
        $apiMessage = [string]$json.message
      }
      if ($null -ne $json.data -and $null -ne $json.data.total) {
        $dataTotal = [string]$json.data.total
      }
      if ($null -ne $json.data -and $null -ne $json.data.list) {
        $dataListCount = [string]@($json.data.list).Count
      }
    } catch {
      $apiMessage = $contentPreview
    }
    return [PSCustomObject]@{
      Url = $url
      Ok = $true
      Status = $status
      ApiCode = $apiCode
      Message = $apiMessage
      DataTotal = $dataTotal
      DataListCount = $dataListCount
      ContentPreview = $contentPreview
      Error = ""
    }
  } catch {
    $status = ""
    $message = $_.Exception.Message
    $contentPreview = ""
    if ($_.Exception.Response) {
      try {
        $status = [int]$_.Exception.Response.StatusCode
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $body = $reader.ReadToEnd()
        if (![string]::IsNullOrWhiteSpace($body)) {
          $contentPreview = ($body -replace '\s+', ' ').Trim()
          if ($contentPreview.Length -gt 160) {
            $contentPreview = $contentPreview.Substring(0, 160)
          }
          $message = $contentPreview
        }
      } catch {}
    }
    return [PSCustomObject]@{
      Url = $url
      Ok = $false
      Status = $status
      ApiCode = ""
      Message = ""
      DataTotal = ""
      DataListCount = ""
      ContentPreview = $contentPreview
      Error = $message
    }
  }
}

function Print-Api-Result {
  param([object]$Result)
  $state = if ($Result.Ok -and ($Result.ApiCode -eq "" -or $Result.ApiCode -eq "0")) { "OK" } else { "FAIL" }
  $detail = if ($Result.Error) { $Result.Error } else { $Result.Message }
  $apiMsg = if ([string]::IsNullOrWhiteSpace($Result.Message)) { "-" } else { $Result.Message }
  $listInfo = ""
  if (![string]::IsNullOrWhiteSpace($Result.DataListCount) -or ![string]::IsNullOrWhiteSpace($Result.DataTotal)) {
    $listInfo = " list_count=$($Result.DataListCount) total=$($Result.DataTotal)"
  }
  Write-Host ("{0} status={1} api_code={2} api_msg={3}{4} path={5}" -f $state, $Result.Status, $Result.ApiCode, $apiMsg, $listInfo, ([Uri]$Result.Url).PathAndQuery)
  if ($state -eq "FAIL" -and ![string]::IsNullOrWhiteSpace($detail) -and $detail -ne "ok") {
    Write-Host ("  detail: {0}" -f $detail)
  }
}

function Find-ResultByPath {
  param(
    [object[]]$Results,
    [string]$Path
  )
  foreach ($result in $Results) {
    if (([Uri]$result.Url).PathAndQuery -eq $Path) {
      return $result
    }
  }
  return $null
}

function Is-Api-Ok {
  param([object]$Result)
  return ($null -ne $Result -and $Result.Ok -and ($Result.ApiCode -eq "" -or $Result.ApiCode -eq "0"))
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
$frontendPlans = Find-ResultByPath -Results $frontendResults -Path "/api/v1/renewals/plans"
$backendPlans = Find-ResultByPath -Results $backendResults -Path "/api/v1/renewals/plans"
if ($backendFail -eq 0 -and $frontendFail -gt 0) {
  Write-Host "LIKELY frontend proxy/static deployment problem: backend direct APIs are OK, proxied APIs failed."
} elseif ($backendFail -gt 0) {
  Write-Host "LIKELY backend or database problem: backend direct APIs failed."
} elseif ($frontendFail -eq 0 -and $backendFail -eq 0) {
  Write-Host "APIs look OK from this host. If browser still fails, check browser Network tab for actual host/base URL."
} else {
  Write-Host "Mixed result. Share Stage 1/2 status lines only."
}

Write-Host ""
Write-Host "Focused /api/v1/renewals/plans diagnosis:"
if ((Is-Api-Ok $backendPlans) -and -not (Is-Api-Ok $frontendPlans)) {
  Write-Host "PLANS_PROBLEM=frontend_proxy_or_base_url"
  Write-Host "Share: Stage 1 renewals/plans line and the frontend URL in the browser address bar."
} elseif (-not (Is-Api-Ok $backendPlans)) {
  Write-Host "PLANS_PROBLEM=backend_or_plan_payload"
  Write-Host ("Backend plans detail: status={0} api_code={1} api_msg={2} detail={3}" -f $backendPlans.Status, $backendPlans.ApiCode, $backendPlans.Message, $backendPlans.Error)
  Write-Host "If api_msg mentions invalid character, gz64, gzip, compressed json, max_allowed_packet, or unmarshal, it is likely historical plan payload compatibility/corruption."
} elseif ((Is-Api-Ok $backendPlans) -and (Is-Api-Ok $frontendPlans)) {
  Write-Host "PLANS_PROBLEM=not_reproduced_by_script"
  Write-Host ("Backend list_count={0} total={1}; Frontend list_count={2} total={3}" -f $backendPlans.DataListCount, $backendPlans.DataTotal, $frontendPlans.DataListCount, $frontendPlans.DataTotal)
  Write-Host "If the page still fails, check browser Network for the actual request host and path."
} else {
  Write-Host "PLANS_PROBLEM=mixed_or_unknown"
}
