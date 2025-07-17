# Скрипт для отслеживания логов сервиса в реальном времени
# Запуск: .\scripts\watch_logs.ps1

param(
    [string]$Service = "subscription-service",
    [switch]$Follow = $true,
    [int]$Lines = 100
)

Write-Host "Отслеживание логов для сервиса $Service..." -ForegroundColor Green

if ($Follow) {
    docker-compose logs -f --tail=$Lines $Service
} else {
    docker-compose logs --tail=$Lines $Service
} 