-- ==========================================
-- XDR Platform - TimescaleDB Initialization Schema
-- ==========================================
-- Ce script initialise la base de données avec :
-- - Extension TimescaleDB pour time-series
-- - Tables pour événements de sécurité
-- - Indexes optimisés pour les requêtes SOC
-- - Hypertables pour performance
-- ==========================================

-- Activer l'extension TimescaleDB
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- ==========================================
-- TABLE : raw_events
-- Tous les événements bruts collectés par les agents
-- ==========================================
CREATE TABLE IF NOT EXISTS raw_events (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    agent_id VARCHAR(64) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    event_type VARCHAR(64) NOT NULL, -- system, network, process, file, registry, etc.
    severity VARCHAR(20) NOT NULL,   -- low, medium, high, critical
    raw_data JSONB NOT NULL,         -- Données brutes en JSON
    source_ip INET,
    destination_ip INET,
    process_name VARCHAR(255),
    process_pid INTEGER,
    username VARCHAR(255),
    tags TEXT[],                     -- Tags pour filtrage rapide
    metadata JSONB,                  -- Métadonnées additionnelles
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (timestamp, id)
);

-- Convertir en hypertable (optimisé pour time-series)
SELECT create_hypertable('raw_events', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Index pour recherches rapides
CREATE INDEX IF NOT EXISTS idx_raw_events_agent_id ON raw_events(agent_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_raw_events_event_type ON raw_events(event_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_raw_events_severity ON raw_events(severity, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_raw_events_hostname ON raw_events(hostname, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_raw_events_source_ip ON raw_events(source_ip, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_raw_events_dest_ip ON raw_events(destination_ip, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_raw_events_process ON raw_events(process_name, timestamp DESC);

-- Index GIN pour recherche JSON
CREATE INDEX IF NOT EXISTS idx_raw_events_raw_data_gin ON raw_events USING GIN(raw_data);
CREATE INDEX IF NOT EXISTS idx_raw_events_tags_gin ON raw_events USING GIN(tags);

-- ==========================================
-- TABLE : alerts
-- Alertes générées par le moteur de détection
-- ==========================================
CREATE TABLE IF NOT EXISTS alerts (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    alert_name VARCHAR(255) NOT NULL,
    rule_id VARCHAR(64) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    confidence_score REAL NOT NULL DEFAULT 0.0, -- Score ML 0-1
    status VARCHAR(32) NOT NULL DEFAULT 'open',  -- open, investigating, closed, false_positive
    description TEXT,
    mitre_attack_id VARCHAR(32)[],               -- Ex: ['T1078', 'T1059']
    related_event_ids BIGINT[],                  -- IDs des événements liés
    affected_hosts TEXT[],
    affected_users TEXT[],
    source_ips INET[],
    destination_ips INET[],
    tags TEXT[],
    metadata JSONB,
    assigned_to VARCHAR(255),
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (timestamp, id)
);

-- Convertir en hypertable
SELECT create_hypertable('alerts', 'timestamp',
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- Index
CREATE INDEX IF NOT EXISTS idx_alerts_rule_id ON alerts(rule_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_mitre ON alerts USING GIN(mitre_attack_id);
CREATE INDEX IF NOT EXISTS idx_alerts_tags ON alerts USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_alerts_metadata ON alerts USING GIN(metadata);

-- ==========================================
-- TABLE : threat_intelligence
-- IOCs (Indicators of Compromise) et threat intel
-- ==========================================
CREATE TABLE IF NOT EXISTS threat_intelligence (
    id SERIAL PRIMARY KEY,
    ioc_type VARCHAR(32) NOT NULL,      -- ip, domain, hash, url, email
    ioc_value TEXT NOT NULL,
    threat_type VARCHAR(64),            -- malware, phishing, c2, etc.
    severity VARCHAR(20) NOT NULL,
    source VARCHAR(255) NOT NULL,       -- AbuseIPDB, OTX, internal, etc.
    description TEXT,
    tags TEXT[],
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index
CREATE UNIQUE INDEX IF NOT EXISTS idx_ti_ioc_unique ON threat_intelligence(ioc_type, ioc_value);
CREATE INDEX IF NOT EXISTS idx_ti_ioc_value ON threat_intelligence(ioc_value);
CREATE INDEX IF NOT EXISTS idx_ti_threat_type ON threat_intelligence(threat_type);
CREATE INDEX IF NOT EXISTS idx_ti_is_active ON threat_intelligence(is_active);
CREATE INDEX IF NOT EXISTS idx_ti_tags ON threat_intelligence USING GIN(tags);

-- ==========================================
-- TABLE : agents
-- Informations sur les agents déployés
-- ==========================================
CREATE TABLE IF NOT EXISTS agents (
    id SERIAL PRIMARY KEY,
    agent_id VARCHAR(64) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    ip_address INET,
    os_type VARCHAR(64),
    os_version VARCHAR(128),
    agent_version VARCHAR(32),
    status VARCHAR(32) NOT NULL DEFAULT 'active', -- active, inactive, error
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    configuration JSONB,
    tags TEXT[],
    metadata JSONB,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agents_agent_id ON agents(agent_id);
CREATE INDEX IF NOT EXISTS idx_agents_hostname ON agents(hostname);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_last_heartbeat ON agents(last_heartbeat DESC);

-- ==========================================
-- TABLE : detection_rules
-- Règles de détection (Sigma, YARA-like, custom)
-- ==========================================
CREATE TABLE IF NOT EXISTS detection_rules (
    id SERIAL PRIMARY KEY,
    rule_id VARCHAR(64) UNIQUE NOT NULL,
    rule_name VARCHAR(255) NOT NULL,
    rule_type VARCHAR(32) NOT NULL,     -- sigma, yara, custom, ml
    severity VARCHAR(20) NOT NULL,
    description TEXT,
    rule_content TEXT NOT NULL,         -- YAML Sigma ou autre format
    mitre_attack_id VARCHAR(32)[],
    tags TEXT[],
    is_enabled BOOLEAN DEFAULT TRUE,
    false_positive_rate REAL DEFAULT 0.0,
    detection_count INTEGER DEFAULT 0,
    last_triggered TIMESTAMPTZ,
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rules_rule_id ON detection_rules(rule_id);
CREATE INDEX IF NOT EXISTS idx_rules_is_enabled ON detection_rules(is_enabled);
CREATE INDEX IF NOT EXISTS idx_rules_severity ON detection_rules(severity);
CREATE INDEX IF NOT EXISTS idx_rules_mitre ON detection_rules USING GIN(mitre_attack_id);

-- ==========================================
-- TABLE : incident_timeline
-- Timeline d'activité pour investigation
-- ==========================================
CREATE TABLE IF NOT EXISTS incident_timeline (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    alert_id BIGINT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    actor VARCHAR(255),                 -- User, process, IP, etc.
    action TEXT NOT NULL,
    target TEXT,
    details JSONB,
    PRIMARY KEY (timestamp, id)
);

SELECT create_hypertable('incident_timeline', 'timestamp',
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_timeline_alert_id ON incident_timeline(alert_id, timestamp DESC);

-- ==========================================
-- TABLE : ml_models
-- Métadonnées des modèles ML déployés
-- ==========================================
CREATE TABLE IF NOT EXISTS ml_models (
    id SERIAL PRIMARY KEY,
    model_name VARCHAR(255) NOT NULL,
    model_type VARCHAR(64) NOT NULL,    -- isolation_forest, lstm, autoencoder, etc.
    version VARCHAR(32) NOT NULL,
    accuracy_score REAL,
    precision_score REAL,
    recall_score REAL,
    f1_score REAL,
    training_date TIMESTAMPTZ,
    deployment_date TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT FALSE,
    model_path TEXT,
    hyperparameters JSONB,
    feature_columns TEXT[],
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ml_models_name ON ml_models(model_name);
CREATE INDEX IF NOT EXISTS idx_ml_models_active ON ml_models(is_active);

-- ==========================================
-- VUES MATERIALISÉES pour analytics
-- ==========================================

-- Vue : Statistiques des alertes par heure
CREATE MATERIALIZED VIEW IF NOT EXISTS alerts_hourly_stats AS
SELECT
    time_bucket('1 hour', timestamp) AS hour,
    severity,
    COUNT(*) AS alert_count,
    COUNT(DISTINCT rule_id) AS unique_rules,
    COUNT(DISTINCT affected_hosts) AS affected_hosts_count
FROM alerts
WHERE timestamp > NOW() - INTERVAL '30 days'
GROUP BY hour, severity
ORDER BY hour DESC;

CREATE INDEX IF NOT EXISTS idx_alerts_hourly_stats ON alerts_hourly_stats(hour DESC);

-- Rafraîchissement automatique (toutes les 15 minutes)
-- Note: à activer manuellement si besoin via un cron job

-- ==========================================
-- FONCTIONS UTILITAIRES
-- ==========================================

-- Fonction pour nettoyer les vieux événements (rétention de données)
CREATE OR REPLACE FUNCTION cleanup_old_events(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM raw_events
    WHERE timestamp < NOW() - (retention_days || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour mettre à jour le statut des agents inactifs
CREATE OR REPLACE FUNCTION update_agent_status()
RETURNS VOID AS $$
BEGIN
    UPDATE agents
    SET status = 'inactive'
    WHERE last_heartbeat < NOW() - INTERVAL '5 minutes'
    AND status = 'active';
END;
$$ LANGUAGE plpgsql;

-- ==========================================
-- TRIGGERS
-- ==========================================

-- Trigger pour mettre à jour updated_at automatiquement
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_alerts_updated_at BEFORE UPDATE ON alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_detection_rules_updated_at BEFORE UPDATE ON detection_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_threat_intelligence_updated_at BEFORE UPDATE ON threat_intelligence
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ml_models_updated_at BEFORE UPDATE ON ml_models
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==========================================
-- POLITIQUE DE RETENTION
-- ==========================================

-- Compression automatique des anciens chunks (économie d'espace)
SELECT add_compression_policy('raw_events', INTERVAL '7 days');
SELECT add_compression_policy('alerts', INTERVAL '30 days');
SELECT add_compression_policy('incident_timeline', INTERVAL '30 days');

-- Rétention : supprimer les données de plus de 180 jours
SELECT add_retention_policy('raw_events', INTERVAL '180 days');
SELECT add_retention_policy('incident_timeline', INTERVAL '180 days');

-- ==========================================
-- DONNÉES DE TEST (optionnel)
-- ==========================================

-- Insérer un agent de test
INSERT INTO agents (agent_id, hostname, ip_address, os_type, os_version, agent_version, status)
VALUES ('agent-test-001', 'test-server-01', '192.168.1.100', 'Linux', 'Ubuntu 22.04', '1.0.0', 'active')
ON CONFLICT (agent_id) DO NOTHING;

-- Insérer une règle de détection de test
INSERT INTO detection_rules (rule_id, rule_name, rule_type, severity, description, rule_content, mitre_attack_id, is_enabled)
VALUES (
    'rule-001',
    'Suspicious PowerShell Execution',
    'sigma',
    'high',
    'Détecte l''exécution de PowerShell avec des arguments suspects',
    'detection: condition: selection',
    ARRAY['T1059.001'],
    TRUE
) ON CONFLICT (rule_id) DO NOTHING;

-- ==========================================
-- GRANTS (permissions)
-- ==========================================

-- Accorder tous les privilèges à l'utilisateur xdr_admin
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO xdr_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO xdr_admin;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO xdr_admin;

-- ==========================================
-- FIN DE L'INITIALISATION
-- ==========================================

-- Afficher les tables créées
SELECT tablename FROM pg_tables WHERE schemaname = 'public';

-- Afficher les hypertables TimescaleDB
SELECT hypertable_name, num_chunks FROM timescaledb_information.hypertables;
