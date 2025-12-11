# start.ps1
Write-Host "Starting application..." -ForegroundColor Green
docker compose up --build -d

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Application is running" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "❌ Failed to start application" -ForegroundColor Red
    exit 1
}