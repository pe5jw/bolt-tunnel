@echo off
chcp 65001 >nul

echo ========================================
echo  Bolttunnel setup-check
echo ========================================

echo.
echo [1/3] Go controleren...

if exist "%~dp0go\bin\go.exe" (
    for /f "tokens=3" %%v in ('"%~dp0go\bin\go.exe" version') do echo      Portable Go gevonden: %%v
    goto go_ok
)

where go >nul 2>&1
if %errorlevel% neq 0 goto download_go

for /f "tokens=3" %%v in ('go version') do set GOVER=%%v
echo      Systeem Go gevonden: %GOVER%

:: Versie check via PowerShell (geen accolades nodig in batch)
powershell -NoProfile -Command "$v='%GOVER%'.Replace('go','').Split('.'); $ok=[int]$v[0]-gt1 -or [int]$v[1]-gt23 -or ([int]$v[1]-eq23 -and [int]($v[2] -replace '[^0-9].*','') -ge 1); exit $(if($ok){0}else{1})"
if %errorlevel% equ 0 (
    echo      Versie voldoet.
    goto go_ok
)

echo      Versie te oud - portable Go 1.23.4 downloaden...
goto download_go

:download_go
echo      Portable Go 1.23.4 downloaden naar .\go\
echo      (ca. 70 MB, even geduld...)
echo.

set "GO_ZIP=%TEMP%\go-portable.zip"

powershell -NoProfile -Command "[Net.ServicePointManager]::SecurityProtocol='Tls12'; Write-Host '     Downloaden...'; Invoke-WebRequest -Uri 'https://dl.google.com/go/go1.23.4.windows-amd64.zip' -OutFile '%GO_ZIP%' -UseBasicParsing"
if %errorlevel% neq 0 (
    curl -L -o "%GO_ZIP%" "https://dl.google.com/go/go1.23.4.windows-amd64.zip"
    if %errorlevel% neq 0 (
        echo [FOUT] Download mislukt.
        exit /b 1
    )
)

echo      Uitpakken naar %~dp0go\...
powershell -NoProfile -Command "Expand-Archive -Path '%GO_ZIP%' -DestinationPath '%~dp0' -Force"
if %errorlevel% neq 0 (
    echo [FOUT] Uitpakken mislukt.
    exit /b 1
)
del /q "%GO_ZIP%" >nul 2>&1

if not exist "%~dp0go\bin\go.exe" (
    echo [FOUT] go.exe niet gevonden na uitpakken.
    exit /b 1
)
echo      Portable Go 1.23.4 geinstalleerd.

:go_ok
echo.
echo [2/3] Go module controleren...
if not exist "%~dp0go.mod" (
    echo [FOUT] go.mod niet gevonden in %~dp0
    exit /b 1
)
echo      go.mod gevonden.

echo.
echo [3/3] Dependencies controleren (tailscale.com/tsnet)...
cd /d "%~dp0"

if not exist "%~dp0go.sum" (
    echo      go.sum ontbreekt - go mod tidy uitvoeren...
    go mod tidy
    if %errorlevel% neq 0 (
        echo [FOUT] go mod tidy mislukt.
        exit /b 1
    )
    echo      Dependencies opgehaald.
) else (
    go list -m tailscale.com >nul 2>&1
    if %errorlevel% neq 0 (
        echo      Module cache incompleet - go mod tidy uitvoeren...
        go mod tidy
        if %errorlevel% neq 0 (
            echo [FOUT] go mod tidy mislukt.
            exit /b 1
        )
    ) else (
        echo      Alle dependencies aanwezig.
    )
)

echo.
echo ========================================
echo  Setup klaar - starten...
echo ========================================
echo.
exit /b 0
