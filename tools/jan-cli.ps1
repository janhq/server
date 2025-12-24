# jan-cli wrapper script for Windows PowerShell
# Automatically builds and runs jan-cli from cmd/jan-cli/

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$CliDir = Join-Path $ScriptDir "jan-cli"
$CliBinary = Join-Path $CliDir "jan-cli.exe"
$MainGo = Join-Path $CliDir "main.go"

# Check if binary needs to be built
$needsBuild = $false
if (-not (Test-Path $CliBinary)) {
    $needsBuild = $true
} else {
    # Check if any .go file is newer than the binary
    $binaryTime = (Get-Item $CliBinary).LastWriteTime
    $goFiles = Get-ChildItem -Path $CliDir -Filter "*.go"
    foreach ($goFile in $goFiles) {
        if ($goFile.LastWriteTime -gt $binaryTime) {
            $needsBuild = $true
            break
        }
    }
}

# Build if needed
if ($needsBuild) {
    Write-Host "Building jan-cli..." -ForegroundColor Yellow
    Push-Location $CliDir
    try {
        go build -o jan-cli.exe .
        if ($LASTEXITCODE -ne 0) {
            throw "Build failed with exit code $LASTEXITCODE"
        }
    } finally {
        Pop-Location
    }
}

# Run jan-cli with all arguments
& $CliBinary $args
exit $LASTEXITCODE
