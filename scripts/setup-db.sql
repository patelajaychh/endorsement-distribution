-- Copyright 2025 Contributors to the Veraison project.
-- SPDX-License-Identifier: Apache-2.0

-- Create database
CREATE DATABASE IF NOT EXISTS endorsements;

-- Connect to the database
\c endorsements;

-- Create endorsements table
CREATE TABLE IF NOT EXISTS endorsements (
    kv_key text NOT NULL,
    kv_val text NOT NULL
);

-- Create index for better performance
CREATE INDEX IF NOT EXISTS idx_endorsements_key ON endorsements(kv_key);

-- Insert sample data for testing
INSERT INTO endorsements (kv_key, kv_val) VALUES 
(
    'coserv://0/tag:arm.com,2023:cca_platform#1.0.0/2/7f454c4602010100000000000000000003003e00010000005058000000000000',
    '["sample_endorsement_data_1", "sample_endorsement_data_2"]'
),
(
    'coserv://0/tag:arm.com,2023:cca_platform#1.0.0/1/0107060504030201000f0e0d0c0b0a090817161514131211101f1e1d1c1b1a1918',
    '["sample_trust_anchor_data"]'
);

-- Grant permissions (adjust as needed)
GRANT ALL PRIVILEGES ON TABLE endorsements TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres; 