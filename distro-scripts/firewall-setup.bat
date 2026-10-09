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
echo Dit script voegt firewall regels toe voor de Bolttunnel binaries.
echo Vereist administrator rechten.
echo.

:: Verwijder oude regels als die er zijn
netsh advfirewall firewall delete rule name="Bolttunnel Client" >nul 2>&1
netsh advfirewall firewall delete rule name="Bolttunnel Server" >nul 2>&1

if exist "%~dp0bolttunnel-client.exe" (
    netsh advfirewall firewall add rule name="Bolttunnel Client" program="%~dp0bolttunnel-client.exe" protocol=TCP dir=in action=allow
    echo [OK] Firewall regel aangemaakt voor bolttunnel-client.exe
) else (
    echo [SKIP] bolttunnel-client.exe niet gevonden
)

if exist "%~dp0bolttunnel-server.exe" (
    netsh advfirewall firewall add rule name="Bolttunnel Server" program="%~dp0bolttunnel-server.exe" protocol=TCP dir=in action=allow
    echo [OK] Firewall regel aangemaakt voor bolttunnel-server.exe
) else (
    echo [SKIP] bolttunnel-server.exe niet gevonden
)

echo.
echo Klaar - geen firewall meldingen meer bij opstarten.
echo.
pause
