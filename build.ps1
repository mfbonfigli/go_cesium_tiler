# Exit on error and undefined variables
$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Main {
    Write-Host "Gocesiumtiler build script" -ForegroundColor Green

    Write-Host -NoNewline " => Removing old build artifacts... " -ForegroundColor Blue
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue ".\build"
    Write-Host "done" -ForegroundColor Blue

    Write-Host " => Starting dockerized build..." -ForegroundColor Blue
    
    # Generate a unique build ID
    $build_id = Get-Date -Format "yyyyMMddHHmmss-fffffff"
    Write-Host " => Build id: $build_id" -ForegroundColor Blue
    
    Write-Host " => Building..." -ForegroundColor Blue
    docker build -t gocesiumtiler:build --target=final --output .\build --build-arg BUILD_ID=$build_id .
    # ensure that the command exited cleanly
    CheckLastExitCode

    # Get the full path to the build directory
    $build_dir = (Get-Item -Path ".\build").FullName
    Write-Host "=> Build complete, artifacts saved in: $build_dir" -ForegroundColor Green
}

function CheckLastExitCode {
    param ([int[]]$SuccessCodes = @(0), [scriptblock]$CleanupScript=$null)

    if ($SuccessCodes -notcontains $LastExitCode) {
        if ($CleanupScript) {
            "Executing cleanup script: $CleanupScript"
            &$CleanupScript
        }
        $msg = "Build failed, check docker logs"
        throw $msg
    }
}

Main