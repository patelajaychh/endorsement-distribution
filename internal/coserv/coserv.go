package coserv

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

)

// CoSERV represents a CoSERV query structure
type CoSERV struct {
	Profile string      `json:"0" cbor:"0"`
	Query   QueryObject `json:"1" cbor:"1"`
}

// QueryObject represents the query part of CoSERV
type QueryObject struct {
	ArtifactType        int                    `json:"0" cbor:"0"`
	EnvironmentSelector EnvironmentSelector    `json:"1" cbor:"1"`
}

// EnvironmentSelector represents the environment selector
type EnvironmentSelector struct {
	Class    []ClassSelector `json:"0,omitempty" cbor:"0,omitempty"`
	Instance []InstanceSelector `json:"1,omitempty" cbor:"1,omitempty"`
}

// ClassSelector represents a class selector
type ClassSelector struct {
	ClassID []byte `json:"0" cbor:"0"`
}

// InstanceSelector represents an instance selector
type InstanceSelector struct {
	InstanceID []byte `json:"1" cbor:"1"`
}

// CoSERVResult represents the result structure
type CoSERVResult struct {
	Profile string      `json:"0" cbor:"0"`
	Result  ResultObject `json:"1" cbor:"1"`
}

// ResultObject represents the result part
type ResultObject struct {
	ArtifactType int         `json:"0" cbor:"0"`
	Artifacts    []Artifact  `json:"1" cbor:"1"`
}

// Artifact represents an individual artifact
type Artifact struct {
	Data []byte `json:"0" cbor:"0"`
}

// FromBase64Url decodes a base64url-encoded CoSERV query
func (c *CoSERV) FromBase64Url(encoded string) error {
	// Replace URL-safe characters
	encoded = strings.ReplaceAll(encoded, "-", "+")
	encoded = strings.ReplaceAll(encoded, "_", "/")
	
	// Add padding if needed
	for len(encoded)%4 != 0 {
		encoded += "="
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// For now, we'll use JSON unmarshaling as a simplified approach
	// In a real implementation, you'd use CBOR decoding
	return c.FromCBOR(data)
}

// FromCBOR decodes CBOR data into CoSERV structure
func (c *CoSERV) FromCBOR(data []byte) error {
	// Simplified CBOR decoding - in reality you'd use a proper CBOR library
	// For now, we'll assume the data is JSON-like and can be parsed
	// This is a placeholder implementation
	
	// Try to parse as JSON first (for testing)
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err == nil {
		// Convert to our structure
		if profile, ok := temp["0"].(string); ok {
			c.Profile = profile
		}
		if queryData, ok := temp["1"].(map[string]interface{}); ok {
			if artifactType, ok := queryData["0"].(float64); ok {
				c.Query.ArtifactType = int(artifactType)
			}
		}
		return nil
	}

	return fmt.Errorf("failed to parse CoSERV data")
}

// GetProfile returns the profile from the CoSERV query
func (c *CoSERV) GetProfile() (string, error) {
	if c.Profile == "" {
		return "", fmt.Errorf("profile not found in CoSERV query")
	}
	return c.Profile, nil
}

// GenerateKey generates a database key from the CoSERV query
func (c *CoSERV) GenerateKey(tenantID string) (string, error) {
	profile, err := c.GetProfile()
	if err != nil {
		return "", err
	}

	// Create a simple key format: coserv://tenant/profile/artifact-type/selector-hash
	// In a real implementation, you'd hash the environment selector
	key := fmt.Sprintf("coserv://%s/%s/%d/%x", 
		tenantID, 
		profile, 
		c.Query.ArtifactType,
		c.Query.EnvironmentSelector.hash())

	return key, nil
}

// hash generates a simple hash of the environment selector
func (es *EnvironmentSelector) hash() []byte {
	// Simplified hash - in reality you'd use a proper hashing algorithm
	if len(es.Class) > 0 {
		return es.Class[0].ClassID
	}
	if len(es.Instance) > 0 {
		return es.Instance[0].InstanceID
	}
	return []byte("default")
}

// CreateResult creates a CoSERV result from artifacts
func CreateResult(profile string, artifactType int, artifacts [][]byte) *CoSERVResult {
	result := &CoSERVResult{
		Profile: profile,
		Result: ResultObject{
			ArtifactType: artifactType,
			Artifacts:    make([]Artifact, len(artifacts)),
		},
	}

	for i, artifact := range artifacts {
		result.Result.Artifacts[i] = Artifact{Data: artifact}
	}

	return result
}

// ToCBOR converts the result to CBOR format
func (r *CoSERVResult) ToCBOR() ([]byte, error) {
	// Simplified CBOR encoding - in reality you'd use a proper CBOR library
	// For now, we'll use JSON as a placeholder
	return json.Marshal(r)
} 