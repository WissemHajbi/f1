[CmdletBinding()]
param(
    [ValidateRange(1950, 2100)]
    [int]$Season = 2025,

    [Parameter(Mandatory = $true)]
    [ValidateNotNullOrEmpty()]
    [int[]]$SessionKeys,

    [hashtable]$RaceSessionLinks = @{},

    [switch]$IncludeTelemetry,

    [switch]$IncludeBestLapTraces,

    [int[]]$DriverNumbers = @(),

    [string]$TelemetryFrom,

    [string]$TelemetryTo,

    [ValidateRange(0, 60000)]
    [int]$ThrottleMilliseconds = 2100
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$binary = Join-Path ([System.IO.Path]::GetTempPath()) "oidysts-sync-$PID.exe"
$invariant = [System.Globalization.CultureInfo]::InvariantCulture
$styles = [System.Globalization.DateTimeStyles]::AssumeUniversal

if ($SessionKeys.Count -eq 0 -or ($SessionKeys | Where-Object { $_ -le 0 })) {
    throw "SessionKeys must contain positive OpenF1 session keys."
}
if ($RaceSessionLinks.Count -eq 0 -and $Season -eq 2025 -and $SessionKeys -contains 9839) {
    $RaceSessionLinks = @{ 24 = 9839 }
}
foreach ($entry in $RaceSessionLinks.GetEnumerator()) {
    if ([int]$entry.Key -le 0 -or [int]$entry.Value -le 0) {
        throw "RaceSessionLinks must map positive rounds to positive session keys."
    }
    if ($SessionKeys -notcontains [int]$entry.Value) {
        throw "RaceSessionLinks session $($entry.Value) must be included in SessionKeys."
    }
}
if ($IncludeBestLapTraces -and $SessionKeys.Count -ne 1) {
    throw "IncludeBestLapTraces currently requires exactly one SessionKeys value."
}
if ($IncludeTelemetry) {
    if ($SessionKeys.Count -ne 1) {
        throw "IncludeTelemetry currently requires exactly one SessionKeys value."
    }
    if ($DriverNumbers | Where-Object { $_ -le 0 }) {
        throw "DriverNumbers must be positive."
    }
    $hasFrom = -not [string]::IsNullOrWhiteSpace($TelemetryFrom)
    $hasTo = -not [string]::IsNullOrWhiteSpace($TelemetryTo)
    if ($hasFrom -ne $hasTo) {
        throw "Provide both TelemetryFrom and TelemetryTo, or omit both for automatic session bounds."
    }
    if ($hasFrom) {
        $telemetryStart = [DateTimeOffset]::Parse($TelemetryFrom, $invariant, $styles).ToUniversalTime()
        $telemetryEnd = [DateTimeOffset]::Parse($TelemetryTo, $invariant, $styles).ToUniversalTime()
        if ($telemetryEnd -le $telemetryStart) {
            throw "TelemetryTo must be after TelemetryFrom."
        }
    }
}

function Invoke-Sync {
    param(
        [Parameter(Mandatory = $true)][string[]]$SyncArgs,
        [switch]$AllowNoSamples
    )

    Write-Host "`n> sync $($SyncArgs -join ' ')" -ForegroundColor Cyan
    $output = & $binary @SyncArgs 2>&1
    $exitCode = $LASTEXITCODE
    $output | ForEach-Object { Write-Host $_ }
    if ($exitCode -ne 0) {
        $text = $output | Out-String
        if ($AllowNoSamples -and $text -match "(car data|location) returned no samples") {
            Write-Host "No samples in this chunk; continuing." -ForegroundColor Yellow
            if ($ThrottleMilliseconds -gt 0) { Start-Sleep -Milliseconds $ThrottleMilliseconds }
            return
        }
        throw "Sync failed: $($SyncArgs -join ' ')"
    }
    if ($ThrottleMilliseconds -gt 0) {
        Start-Sleep -Milliseconds $ThrottleMilliseconds
    }
}

function Format-Rfc3339 {
    param([DateTimeOffset]$Value)
    return $Value.ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ss.fffffffZ", $invariant)
}

Push-Location $repoRoot
try {
    Write-Host "Building sync command..." -ForegroundColor Green
    & go build -o $binary ./cmd/sync
    if ($LASTEXITCODE -ne 0) {
        throw "Could not build cmd/sync."
    }

    # Provider registries first: detailed session tables reference OpenF1 sessions.
    Invoke-Sync -SyncArgs @("-resource", "meetings", "-year", "$Season")

    # Complete season-wide Jolpica datasets.
    foreach ($resource in @("drivers", "calendar", "standings", "results", "classifications")) {
        Invoke-Sync -SyncArgs @("-resource", $resource, "-year", "$Season")
    }

    foreach ($entry in $RaceSessionLinks.GetEnumerator()) {
        Invoke-Sync -SyncArgs @("-resource", "race-link", "-year", "$Season", "-round", "$($entry.Key)", "-session", "$($entry.Value)")
        if ($Season -eq 2025 -and [int]$entry.Key -eq 24 -and [int]$entry.Value -eq 9839) {
            $geometryFile = Join-Path $repoRoot "data/circuits/yas_marina_openf1.json"
            Invoke-Sync -SyncArgs @("-resource", "circuit-geometry", "-year", "$Season", "-round", "$($entry.Key)",
                "-session", "$($entry.Value)", "-file", $geometryFile)
        }
    }

    foreach ($sessionKey in $SessionKeys) {
        Write-Host "`nSynchronizing session $sessionKey..." -ForegroundColor Green

        # Laps must precede stints because the unified timeline derives stint timestamps from lap starts.
        foreach ($resource in @("laps", "stints", "pit", "weather", "race-control", "overtakes", "positions", "intervals", "team-radio")) {
            Invoke-Sync -SyncArgs @("-resource", $resource, "-session", "$sessionKey")
        }
    }

    if ($IncludeBestLapTraces) {
        $sessionKey = $SessionKeys[0]
        $userAgent = if ($env:UPSTREAM_USER_AGENT) { $env:UPSTREAM_USER_AGENT } else { "oidysts/0.1" }
        $traceDrivers = $DriverNumbers
        if ($traceDrivers.Count -eq 0) {
            Write-Host "Discovering drivers for best-lap traces..." -ForegroundColor Green
            $driverRows = Invoke-RestMethod -Uri "https://api.openf1.org/v1/drivers?session_key=$sessionKey" -UserAgent $userAgent
            $traceDrivers = @($driverRows | ForEach-Object { [int]$_.driver_number } | Sort-Object -Unique)
            if ($traceDrivers.Count -eq 0) { throw "OpenF1 returned no session drivers." }
            Start-Sleep -Milliseconds $ThrottleMilliseconds
        }
        foreach ($driverNumber in $traceDrivers) {
            Invoke-Sync -SyncArgs @("-resource", "best-lap-location", "-session", "$sessionKey", "-driver", "$driverNumber")
        }
    }

    if ($IncludeTelemetry) {
        $sessionKey = $SessionKeys[0]
        $userAgent = if ($env:UPSTREAM_USER_AGENT) { $env:UPSTREAM_USER_AGENT } else { "oidysts/0.1" }
        if ($DriverNumbers.Count -eq 0) {
            Write-Host "Discovering session drivers..." -ForegroundColor Green
            $driverRows = Invoke-RestMethod -Uri "https://api.openf1.org/v1/drivers?session_key=$sessionKey" -UserAgent $userAgent
            $DriverNumbers = @($driverRows | ForEach-Object { [int]$_.driver_number } | Sort-Object -Unique)
            if ($DriverNumbers.Count -eq 0) { throw "OpenF1 returned no session drivers." }
            Start-Sleep -Milliseconds $ThrottleMilliseconds
        }
        if ([string]::IsNullOrWhiteSpace($TelemetryFrom)) {
            Write-Host "Discovering session telemetry bounds..." -ForegroundColor Green
            $sessionRows = @(Invoke-RestMethod -Uri "https://api.openf1.org/v1/sessions?session_key=$sessionKey" -UserAgent $userAgent)
            if ($sessionRows.Count -ne 1) { throw "OpenF1 did not return exactly one session." }
            $telemetryStart = [DateTimeOffset]::Parse($sessionRows[0].date_start, $invariant, $styles).ToUniversalTime()
            $telemetryEnd = [DateTimeOffset]::Parse($sessionRows[0].date_end, $invariant, $styles).ToUniversalTime()
            Start-Sleep -Milliseconds $ThrottleMilliseconds
        }
        Write-Host "Telemetry drivers: $($DriverNumbers -join ', ')" -ForegroundColor Green
        Write-Host "Telemetry range: $(Format-Rfc3339 $telemetryStart) to $(Format-Rfc3339 $telemetryEnd)" -ForegroundColor Green
        foreach ($driverNumber in ($DriverNumbers | Sort-Object -Unique)) {
            $cursor = $telemetryStart
            while ($cursor -lt $telemetryEnd) {
                $chunkEnd = $cursor.AddMinutes(15)
                if ($chunkEnd -gt $telemetryEnd) {
                    $chunkEnd = $telemetryEnd
                }
                $from = Format-Rfc3339 $cursor
                $to = Format-Rfc3339 $chunkEnd
                foreach ($resource in @("car-data", "location")) {
                    Invoke-Sync -AllowNoSamples -SyncArgs @("-resource", $resource, "-session", "$sessionKey",
                        "-driver", "$driverNumber", "-from", $from, "-to", $to)
                }
                $cursor = $chunkEnd
            }
        }
    }

    Write-Host "`nDatabase synchronization completed successfully." -ForegroundColor Green
    if (-not $IncludeTelemetry) {
        Write-Host "Full-session car-data/location were skipped. Use -IncludeTelemetry for complete traces or -IncludeBestLapTraces for bounded best-lap geometry." -ForegroundColor Yellow
    }
}
finally {
    Pop-Location
    Remove-Item $binary -Force -ErrorAction SilentlyContinue
}
