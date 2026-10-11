@echo off
if not "%CMDEXTVERSION%"=="" goto run
cmd /c "%~f0" %*
exit /b %errorlevel%

:run
chcp 65001 >nul

if exist "%~dp0go\bin\go.exe" (
    set "GOROOT=%~dp0go"
    set "GOPATH=%~dp0.gopath"
    set "GOCACHE=%~dp0.gocache"
    set "PATH=%~dp0go\bin;%PATH%"
)

echo ========================================
echo  Bolttunnel — Windows build
echo ========================================

call "%~dp0setup.bat"
if %errorlevel% neq 0 ( echo Setup mislukt. & pause & exit /b 1 )

if exist "%~dp0go\bin\go.exe" (
    set "GOROOT=%~dp0go"
    set "GOPATH=%~dp0.gopath"
    set "GOCACHE=%~dp0.gocache"
    set "PATH=%~dp0go\bin;%PATH%"
)

if not exist "%~dp0bin" mkdir "%~dp0bin"
if not exist "%~dp0distro" mkdir "%~dp0distro"

cd /d "%~dp0"
echo.
echo Modules bijwerken...
go mod tidy
go mod download
if %errorlevel% neq 0 ( echo [FOUT] go mod download mislukt. & pause & exit /b 1 )

echo.
echo ========================================
echo  Auth key configuratie
echo ========================================
echo.
echo Voer de Tailscale auth key in om in te bakken in de client binary.
echo Laat leeg om de key weg te laten (gebruiker logt zelf in via browser).
echo.
set /p BUILD_AUTHKEY="Auth key (tskey-auth-...): "
echo.

set GOOS=windows
set GOARCH=amd64

echo [1/2] Client bouwen (Windows amd64)...
cd /d "%~dp0cmd\client"
if "%BUILD_AUTHKEY%"=="" (
    go build -o "%~dp0bin\bolttunnel-client.exe" .
) else (
    go build -ldflags "-X main.builtinAuthKey=%BUILD_AUTHKEY%" -o "%~dp0bin\bolttunnel-client.exe" .
    set "BUILD_AUTHKEY="
)
if %errorlevel% neq 0 ( echo [FOUT] Client build mislukt. & pause & exit /b 1 )
echo      OK: bin\bolttunnel-client.exe

echo.
echo [2/2] Server bouwen (Windows amd64)...
cd /d "%~dp0cmd\server"
go build -o "%~dp0bin\bolttunnel-server.exe" .
if %errorlevel% neq 0 ( echo [FOUT] Server build mislukt. & pause & exit /b 1 )
echo      OK: bin\bolttunnel-server.exe

echo.
echo Distro map vullen...
copy /y "%~dp0bin\bolttunnel-client.exe" "%~dp0distro\" >nul
copy /y "%~dp0bin\bolttunnel-server.exe" "%~dp0distro\" >nul
copy /y "%~dp0distro-scripts\start-client.bat" "%~dp0distro\" >nul
copy /y "%~dp0distro-scripts\start-server.bat" "%~dp0distro\" >nul
copy /y "%~dp0distro-scripts\firewall-setup.bat" "%~dp0distro\" >nul

echo.
echo ========================================
echo  Build klaar — Windows
echo ========================================
echo.
dir /b "%~dp0distro\*.exe"
echo.
pause
