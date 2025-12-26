@echo off
REM Jan Platform Web - Environment Setup Script
REM This script helps you quickly set up your .env.local file

echo.
echo ========================================
echo Jan Platform Web - Environment Setup
echo ========================================
echo.

REM Check if .env.example exists
if not exist ".env.example" (
    echo ERROR: .env.example file not found!
    echo Please run this script from the platform-web directory.
    pause
    exit /b 1
)

REM Check if .env.local already exists
if exist ".env.local" (
    echo WARNING: .env.local already exists!
    set /p overwrite="Do you want to overwrite it? (y/N): "
    if /i not "%overwrite%"=="y" (
        echo Setup cancelled. Existing .env.local kept.
        pause
        exit /b 0
    )
    echo Backing up existing .env.local to .env.local.backup...
    copy /y ".env.local" ".env.local.backup" >nul
)

REM Copy .env.example to .env.local
echo Creating .env.local from .env.example...
copy /y ".env.example" ".env.local" >nul

REM Prompt for API URL
echo.
echo ========================================
echo API Configuration
echo ========================================
echo Please enter your API base URL
echo Press Enter to use default: http://localhost:8000
echo.
echo Common options:
echo   1. http://localhost:8000              (Local development)
echo   2. https://api-gateway-dev.jan.ai     (Staging/Dev server)
echo   3. Custom URL
echo.
set "api_url=http://localhost:8000"
set /p api_url="Enter API URL [http://localhost:8000]: "

REM If empty, use default
if "%api_url%"=="" set "api_url=http://localhost:8000"

REM Remove trailing slash if present
if "%api_url:~-1%"=="/" set "api_url=%api_url:~0,-1%"

REM Validate URL format
echo %api_url% | findstr /r /c:"^http://" /c:"^https://" >nul
if errorlevel 1 (
    echo ERROR: Invalid URL format. URL must start with http:// or https://
    pause
    exit /b 1
)

REM Update .env.local with the API URL using PowerShell
powershell -Command "(Get-Content .env.local) -replace 'NEXT_PUBLIC_JAN_BASE_URL=.*', 'NEXT_PUBLIC_JAN_BASE_URL=%api_url%' | Set-Content .env.local"

echo.
echo ========================================
echo Environment setup complete!
echo ========================================
echo.
echo Configuration:
echo   API URL: %api_url%
echo   File: .env.local
echo.
echo Next steps:
echo   1. Start your backend API server (if running locally)
echo   2. Run: npm run dev
echo   3. Open: http://localhost:3000
echo.
echo To verify your backend is running:
echo   curl %api_url%/healthz
echo.
echo For more information, see ENVIRONMENT_SETUP.md
echo.
pause
