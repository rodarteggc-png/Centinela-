@echo off
title Detener Centinela SQL Server PI
cd /d "%~dp0"
echo ============================================================
echo   DETENIENDO FOGlight LOCAL PARA SQL SERVER
echo ============================================================
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0portable\stop-centinela.ps1"
pause
