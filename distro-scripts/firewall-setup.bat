@echo off
:: Voer dit script eenmalig uit als administrator
if not "%CMDEXTVERSION%"=="" goto run
cmd /c "%~f0" %*
exit /b %errorlevel%

:run
chcp 65001 >nul

echo ========================================
echo  Bolttunnel firewall setup
echo ========================================
echo.
echo Dit script voegt firewall regels toe voor Bolttunnel.
echo Vereist administrator rechten.
echo.

:: Controleer admin rechten
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo [FOUT] Start dit script als Administrator.
    echo Klik rechts op het script en kies "Als administrator uitvoeren".
    echo.
    pause
    exit /b 1
)

:: Verwijder oude regels
netsh advfirewall firewall delete rule name="Bolttunnel Client"     >nul 2>&1
netsh advfirewall firewall delete rule name="Bolttunnel Client TCP" >nul 2>&1
netsh advfirewall firewall delete rule name="Bolttunnel Server"     >nul 2>&1
netsh advfirewall firewall delete rule name="Bolttunnel Server TCP" >nul 2>&1

echo ── Client regels ────────────────────────────────────────
echo.

if exist "%~dp0bolttunnel-client.exe" (
    netsh advfirewall firewall add rule ^
        name="Bolttunnel Client" ^
        program="%~dp0bolttunnel-client.exe" ^
        protocol=TCP dir=in action=allow >nul
    echo [OK] Binary regel: bolttunnel-client.exe
) else (
    echo [SKIP] bolttunnel-client.exe niet gevonden in deze map
)

:: Poort 7781 - server pingt client hierop terug
netsh advfirewall firewall add rule ^
    name="Bolttunnel Client TCP" ^
    protocol=TCP localport=7781 dir=in action=allow >nul
echo [OK] Poort 7781/TCP inkomend (server ping naar client)

echo.
echo ── Server regels ────────────────────────────────────────
echo.

if exist "%~dp0bolttunnel-server.exe" (
    netsh advfirewall firewall add rule ^
        name="Bolttunnel Server" ^
        program="%~dp0bolttunnel-server.exe" ^
        protocol=TCP dir=in action=allow >nul
    echo [OK] Binary regel: bolttunnel-server.exe
) else (
    echo [SKIP] bolttunnel-server.exe niet gevonden in deze map
)

:: Poort 7780 - client verbindt hiermee
netsh advfirewall firewall add rule ^
    name="Bolttunnel Server TCP" ^
    protocol=TCP localport=7780 dir=in action=allow >nul
echo [OK] Poort 7780/TCP inkomend (client verbinding naar server)

echo.
echo ========================================
echo  Klaar
echo ========================================
echo.
echo Regels aangemaakt:
echo   Client binary + poort 7781
echo   Server binary + poort 7780
echo.
echo Voer dit script opnieuw uit na een herinstallatie.
echo.
pause
