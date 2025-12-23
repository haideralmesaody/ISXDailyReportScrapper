@echo off
:: ISX Daily Reports Scrapper - Production Credentials Setup
:: Wrapper for PowerShell setup script

powershell.exe -ExecutionPolicy Bypass -File "tools\setup-production-credentials.ps1" %*