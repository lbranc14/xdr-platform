# XDR Platform - Stop Infrastructure Script
# Script PowerShell pour arreter proprement l'infrastructure Docker

Write-Host "XDR Platform - Arret de l'infrastructure..." -ForegroundColor Cyan
Write-Host ""

# Verifier que docker-compose.yml existe
if (-not (Test-Path "..\docker-compose.yml")) {
    Write-Host "Erreur: docker-compose.yml introuvable!" -ForegroundColor Red
    Write-Host "Assurez-vous d'etre dans le dossier scripts du projet." -ForegroundColor Yellow
    exit 1
}

# Demander confirmation
Write-Host "Cette action va arreter tous les containers Docker." -ForegroundColor Yellow
$confirmation = Read-Host "Voulez-vous continuer? (o/N)"

if ($confirmation -ne "o" -and $confirmation -ne "O") {
    Write-Host "Operation annulee." -ForegroundColor Red
    exit 0
}

Write-Host ""
Write-Host "Arret des containers en cours..." -ForegroundColor Cyan

# Arreter les containers
Set-Location ..
docker-compose down
$exitCode = $LASTEXITCODE
Set-Location scripts

if ($exitCode -eq 0) {
    Write-Host ""
    Write-Host "Infrastructure arretee avec succes!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Notes:" -ForegroundColor Cyan
    Write-Host "  Les donnees sont conservees dans les volumes Docker" -ForegroundColor Gray
    Write-Host "  Pour redemarrer: .\scripts\start.ps1" -ForegroundColor Gray
    Write-Host "  Pour supprimer aussi les volumes: docker-compose down -v" -ForegroundColor Gray
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "Erreur lors de l'arret de l'infrastructure!" -ForegroundColor Red
    exit 1
}
