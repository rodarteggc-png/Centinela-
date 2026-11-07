@echo off
title Centinela SQL Server PI
cd /d "%~dp0"
echo ============================================================
echo   INICIANDO FOGlight LOCAL PARA SQL SERVER
echo ============================================================
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0portable\start-centinela.ps1"
pause
