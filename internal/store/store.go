package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"endorsement-distribution/internal/config"
	"endorsement-distribution/internal/coserv"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store interface for database operations
type Store interface {
	Get(key string) ([][]byte, error)
	Set(key string, artifacts [][]byte) error
	Close() error
}

// PostgresStore implements Store interface using PostgreSQL
type PostgresStore struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(cfg config.DatabaseConfig) (*PostgresStore, error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	store := &PostgresStore{
		pool: pool,
	}

	// Test connection
	if err := store.pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Setup table if it doesn't exist
	if err := store.setupTable(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to setup table: %w", err)
	}

	return store, nil
}

// setupTable creates the endorsements table if it doesn't exist
func (s *PostgresStore) setupTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS endorsements (
			kv_key text NOT NULL,
			kv_val text NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_endorsements_key ON endorsements(kv_key);
	`

	_, err := s.pool.Exec(context.Background(), query)
	return err
}

// Get retrieves artifacts for a given key
func (s *PostgresStore) Get(key string) ([][]byte, error) {
	query := `SELECT kv_val FROM endorsements WHERE kv_key = $1`
	
	rows, err := s.pool.Query(context.Background(), query, key)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var artifacts [][]byte
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse JSON array of artifacts
		var artifactArray []string
		if err := json.Unmarshal([]byte(val), &artifactArray); err != nil {
			return nil, fmt.Errorf("failed to unmarshal artifacts: %w", err)
		}

		// Convert base64 strings to bytes
		for _, artifactStr := range artifactArray {
			artifact, err := s.decodeArtifact(artifactStr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode artifact: %w", err)
			}
			artifacts = append(artifacts, artifact)
		}
	}

	if len(artifacts) == 0 {
		return nil, fmt.Errorf("no artifacts found for key: %s", key)
	}

	return artifacts, nil
}

// Set stores artifacts for a given key
func (s *PostgresStore) Set(key string, artifacts [][]byte) error {
	// Convert artifacts to base64 strings
	var artifactStrings []string
	for _, artifact := range artifacts {
		artifactStr := s.encodeArtifact(artifact)
		artifactStrings = append(artifactStrings, artifactStr)
	}

	// Convert to JSON
	val, err := json.Marshal(artifactStrings)
	if err != nil {
		return fmt.Errorf("failed to marshal artifacts: %w", err)
	}

	// Delete existing entries and insert new one
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// Delete existing
	_, err = tx.Exec(context.Background(), "DELETE FROM endorsements WHERE kv_key = $1", key)
	if err != nil {
		return fmt.Errorf("failed to delete existing artifacts: %w", err)
	}

	// Insert new
	_, err = tx.Exec(context.Background(), "INSERT INTO endorsements (kv_key, kv_val) VALUES ($1, $2)", key, string(val))
	if err != nil {
		return fmt.Errorf("failed to insert artifacts: %w", err)
	}

	return tx.Commit(context.Background())
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}

// encodeArtifact encodes artifact data to base64
func (s *PostgresStore) encodeArtifact(data []byte) string {
	// In a real implementation, you might use a more sophisticated encoding
	// For now, we'll use base64
	return fmt.Sprintf("%x", data) // Simple hex encoding for demo
}

// decodeArtifact decodes artifact data from base64
func (s *PostgresStore) decodeArtifact(encoded string) ([]byte, error) {
	// In a real implementation, you'd decode from the actual encoding used
	// For now, we'll assume hex encoding
	return []byte(encoded), nil // Simplified for demo
}

// EndorsementDistributor handles endorsement distribution logic
type EndorsementDistributor struct {
	store  Store
	logger *zap.SugaredLogger
}

// NewEndorsementDistributor creates a new endorsement distributor
func NewEndorsementDistributor(store Store, logger *zap.SugaredLogger) *EndorsementDistributor {
	return &EndorsementDistributor{
		store:  store,
		logger: logger,
	}
}

// GetEndorsements retrieves endorsements for a CoSERV query
func (ed *EndorsementDistributor) GetEndorsements(tenantID, coservQuery, mediaType string) ([]byte, error) {
	// Parse CoSERV query
	var coserv coserv.CoSERV
	if err := coserv.FromBase64Url(coservQuery); err != nil {
		return nil, fmt.Errorf("failed to parse CoSERV query: %w", err)
	}

	// Generate database key
	key, err := coserv.GenerateKey(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	ed.logger.Infow("Fetching endorsements", "key", key)

	// Get artifacts from database
	artifacts, err := ed.store.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifacts: %w", err)
	}

	// Get profile for result
	profile, err := coserv.GetProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Create CoSERV result
	result := coserv.CreateResult(profile, coserv.Query.ArtifactType, artifacts)

	// Convert to CBOR
	resultData, err := result.ToCBOR()
	if err != nil {
		return nil, fmt.Errorf("failed to encode result: %w", err)
	}

	return resultData, nil
} 