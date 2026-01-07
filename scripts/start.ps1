# XDR Platform - Start Infrastructure Script
# Script PowerShell pour demarrer l'infrastructure Docker facilement

Write-Host "XDR Platform - Demarrage de l'infrastructure..." -ForegroundColor Cyan
Write-Host ""

# Verifier que Docker est bien lance
$dockerRunning = docker info 2>&1 | Select-String "Server Version"
if (-not $dockerRunning) {
    Write-Host "Erreur: Docker n'est pas lance!" -ForegroundColor Red
    Write-Host "Veuillez demarrer Docker Desktop et reessayer." -ForegroundColor Yellow
    exit 1
}

Write-Host "Docker est en cours d'execution" -ForegroundColor Green

# Verifier que le fichier docker-compose.yml existe
if (-not (Test-Path "..\docker-compose.yml")) {
    Write-Host "Erreur: docker-compose.yml introuvable!" -ForegroundColor Red
    Write-Host "Assurez-vous d'etre dans le dossier scripts du projet." -ForegroundColor Yellow
    exit 1
}

Write-Host "Fichier docker-compose.yml trouve" -ForegroundColor Green
Write-Host ""

# Creer les dossiers necessaires s'ils n'existent pas
Write-Host "Creation des dossiers necessaires..." -ForegroundColor Cyan
$folders = @("..\logs", "..\data")
foreach ($folder in $folders) {
    if (-not (Test-Path $folder)) {
        New-Item -ItemType Directory -Path $folder | Out-Null
        Write-Host "Dossier cree: $folder" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Demarrage des containers Docker..." -ForegroundColor Cyan
Write-Host "Cela peut prendre quelques minutes la premiere fois..." -ForegroundColor Gray
Write-Host ""

# Demarrer les containers
Set-Location ..
docker-compose up -d
$exitCode = $LASTEXITCODE
Set-Location scripts

if ($exitCode -eq 0) {
    Write-Host ""
    Write-Host "Infrastructure demarree avec succes!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Services disponibles:" -ForegroundColor Cyan
    Write-Host "  TimescaleDB    : localhost:5432" -ForegroundColor White
    Write-Host "  Redis          : localhost:6379" -ForegroundColor White
    Write-Host "  Kafka          : localhost:9092" -ForegroundColor White
    Write-Host "  Kafka UI       : http://localhost:8080" -ForegroundColor White
    Write-Host "  pgAdmin        : http://localhost:5050" -ForegroundColor White
    Write-Host ""
    Write-Host "Identifiants pgAdmin:" -ForegroundColor Cyan
    Write-Host "  Email    : admin@xdr.local" -ForegroundColor White
    Write-Host "  Password : admin_password_2024" -ForegroundColor White
    Write-Host ""
    Write-Host "Commandes utiles:" -ForegroundColor Cyan
    Write-Host "  Voir les logs         : docker-compose logs -f" -ForegroundColor Gray
    Write-Host "  Arreter l'infra       : .\scripts\stop.ps1" -ForegroundColor Gray
    Write-Host "  Redemarrer l'infra    : docker-compose restart" -ForegroundColor Gray
    Write-Host "  Voir le statut        : docker-compose ps" -ForegroundColor Gray
    Write-Host ""
    
    # Attendre que tous les services soient healthy
    Write-Host "Verification de la sante des services..." -ForegroundColor Cyan
    Start-Sleep -Seconds 5
    
    $services = @("xdr-timescaledb", "xdr-redis", "xdr-kafka")
    $allHealthy = $true
    
    foreach ($service in $services) {
        $health = docker inspect --format='{{.State.Health.Status}}' $service 2>$null
        if ($health -eq "healthy" -or $health -eq "") {
            Write-Host "  OK : $service" -ForegroundColor Green
        } else {
            Write-Host "  En cours : $service" -ForegroundColor Yellow
            $allHealthy = $false
        }
    }
    
    Write-Host ""
    if ($allHealthy) {
        Write-Host "Tous les services sont operationnels!" -ForegroundColor Green
    } else {
        Write-Host "Certains services demarrent encore. Patientez 1-2 minutes." -ForegroundColor Yellow
        Write-Host "Utilisez 'docker-compose ps' pour verifier le statut." -ForegroundColor Gray
    }
    
    Write-Host ""
    Write-Host "Vous pouvez maintenant commencer a developper!" -ForegroundColor Cyan
    Write-Host ""
    
} else {
    Write-Host ""
    Write-Host "Erreur lors du demarrage de l'infrastructure!" -ForegroundColor Red
    Write-Host "Consultez les logs avec: docker-compose logs" -ForegroundColor Yellow
    exit 1
}
