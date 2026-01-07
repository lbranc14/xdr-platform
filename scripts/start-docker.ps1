# Script de démarrage de la plateforme XDR complète avec Docker

Write-Host "🚀 Démarrage de la plateforme XDR..." -ForegroundColor Cyan

# Vérifier que Docker est lancé
$dockerRunning = docker info 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker n'est pas lancé. Veuillez démarrer Docker Desktop." -ForegroundColor Red
    exit 1
}

Write-Host "✅ Docker est actif" -ForegroundColor Green

# Arrêter les anciens containers s'ils existent
Write-Host "🛑 Arrêt des anciens containers..." -ForegroundColor Yellow
docker-compose -f docker-compose.yml down

# Build et démarrage de tous les services
Write-Host "🏗️  Build des images Docker..." -ForegroundColor Cyan
docker-compose -f docker-compose.yml build

Write-Host "🚀 Démarrage de tous les services..." -ForegroundColor Cyan
docker-compose -f docker-compose.yml up -d

# Attendre que les services soient prêts
Write-Host "⏳ Attente du démarrage des services..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Vérifier l'état des containers
Write-Host "`n📊 État des services:" -ForegroundColor Cyan
docker-compose -f docker-compose.yml ps

Write-Host "`n✅ Plateforme XDR démarrée avec succès!" -ForegroundColor Green
Write-Host "`n🌐 Services disponibles:" -ForegroundColor Cyan
Write-Host "   - Dashboard:    http://localhost" -ForegroundColor White
Write-Host "   - API Gateway:  http://localhost:8000" -ForegroundColor White
Write-Host "   - Kafka UI:     http://localhost:8080" -ForegroundColor White
Write-Host "   - pgAdmin:      http://localhost:5050" -ForegroundColor White

Write-Host "`n📝 Commandes utiles:" -ForegroundColor Cyan
Write-Host "   - Voir les logs:        docker-compose -f docker-compose.yml logs -f" -ForegroundColor White
Write-Host "   - Arrêter la stack:     docker-compose -f docker-compose.yml down" -ForegroundColor White
Write-Host "   - Redémarrer un service: docker-compose -f docker-compose.yml restart <service>" -ForegroundColor White

Write-Host "`n🎉 Profitez de votre plateforme XDR!" -ForegroundColor Green
