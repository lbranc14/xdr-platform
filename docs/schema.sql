-- XDR Platform - TimescaleDB Schema
-- Database: xdr_events
-- PostgreSQL Version: 15

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Main events table
CREATE TABLE raw_events (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    hostname TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    source_ip TEXT,
    destination_ip TEXT,
    process_name TEXT,
    process_pid INTEGER,
    username TEXT,
    tags TEXT[],
    metadata JSONB,
    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable (partitioned by time)
SELECT create_hypertable('raw_events', 'timestamp');

-- Indexes for performance
CREATE INDEX idx_raw_events_timestamp ON raw_events (timestamp DESC);
CREATE INDEX idx_raw_events_event_type ON raw_events (event_type);
CREATE INDEX idx_raw_events_severity ON raw_events (severity);
CREATE INDEX idx_raw_events_hostname ON raw_events (hostname);
CREATE INDEX idx_raw_events_agent_id ON raw_events (agent_id);
CREATE INDEX idx_raw_events_source_ip ON raw_events (source_ip);
CREATE INDEX idx_raw_events_tags ON raw_events USING GIN (tags);
CREATE INDEX idx_raw_events_metadata ON raw_events USING GIN (metadata);

-- Compression policy (compress chunks older than 7 days)
SELECT add_compression_policy('raw_events', INTERVAL '7 days');

-- Retention policy (drop chunks older than 90 days)
SELECT add_retention_policy('raw_events', INTERVAL '90 days');

-- Sample data generation (for testing)
DO $$
DECLARE
    i INT;
    event_types TEXT[] := ARRAY['system', 'network', 'process', 'file'];
    severities TEXT[] := ARRAY['low', 'medium', 'high', 'critical'];
    hostnames TEXT[] := ARRAY['web-server-01', 'db-server-01', 'app-server-01', 'worker-01', 'worker-02'];
    processes TEXT[] := ARRAY['nginx', 'postgres', 'node', 'python', 'redis-server', 'dockerd'];
    usernames TEXT[] := ARRAY['root', 'www-data', 'postgres', 'admin', 'jenkins'];
    all_tags TEXT[] := ARRAY['production', 'security', 'performance', 'monitoring', 'backup'];
    
    rand_event_type TEXT;
    rand_severity TEXT;
    rand_hostname TEXT;
    rand_process TEXT;
    rand_username TEXT;
    rand_tags TEXT[];
    rand_timestamp TIMESTAMPTZ;
    rand_source_ip TEXT;
    rand_dest_ip TEXT;
    rand_pid INT;
    raw_data_json JSONB;
BEGIN
    FOR i IN 1..500 LOOP
        rand_event_type := event_types[1 + floor(random() * 4)::int];
        rand_severity := severities[1 + floor(random() * 4)::int];
        rand_hostname := hostnames[1 + floor(random() * 5)::int];
        rand_process := processes[1 + floor(random() * 6)::int];
        rand_username := usernames[1 + floor(random() * 5)::int];
        
        rand_tags := ARRAY[
            all_tags[1 + floor(random() * 5)::int],
            all_tags[1 + floor(random() * 5)::int]
        ];
        
        rand_timestamp := NOW() - (random() * INTERVAL '24 hours');
        rand_source_ip := '10.0.' || floor(random() * 255)::int || '.' || floor(random() * 255)::int;
        rand_dest_ip := '172.16.' || floor(random() * 255)::int || '.' || floor(random() * 255)::int;
        rand_pid := 1 + floor(random() * 65535)::int;
        
        raw_data_json := jsonb_build_object(
            'event_id', i,
            'message', 'Event ' || i || ': ' || rand_event_type || ' on ' || rand_hostname,
            'details', jsonb_build_object(
                'action', rand_event_type,
                'target', rand_hostname,
                'user', rand_username
            )
        );
        
        INSERT INTO raw_events (
            timestamp, event_type, severity, hostname, agent_id,
            source_ip, destination_ip, process_name, process_pid,
            username, tags, raw_data
        ) VALUES (
            rand_timestamp,
            rand_event_type,
            rand_severity,
            rand_hostname,
            'agent-' || rand_hostname,
            rand_source_ip,
            rand_dest_ip,
            rand_process,
            rand_pid,
            rand_username,
            rand_tags,
            raw_data_json
        );
        
        IF i % 100 = 0 THEN
            RAISE NOTICE 'Inserted % events', i;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Done! Total: %', (SELECT COUNT(*) FROM raw_events);
END $$;

-- Useful queries
-- Get recent events
SELECT * FROM raw_events ORDER BY timestamp DESC LIMIT 10;

-- Count by severity
SELECT severity, COUNT(*) FROM raw_events GROUP BY severity;

-- Timeline aggregation (hourly)
SELECT 
    time_bucket('1 hour', timestamp) AS hour,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE severity = 'critical') as critical,
    COUNT(*) FILTER (WHERE severity = 'high') as high,
    COUNT(*) FILTER (WHERE severity = 'medium') as medium,
    COUNT(*) FILTER (WHERE severity = 'low') as low
FROM raw_events
WHERE timestamp > NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour DESC;
