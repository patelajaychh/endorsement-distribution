package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"endorsement-distribution/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	tenantID       = "0"
	EdApiMediaType = "application/coserv+cbor"
)

type Handler struct {
	Logger                *zap.SugaredLogger
	EndorsementDistributor *store.EndorsementDistributor
}

func NewHandler(endorsementDistributor *store.EndorsementDistributor, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		EndorsementDistributor: endorsementDistributor,
		Logger:                logger,
	}
}

// GetEdApiWellKnownInfo handles the well-known endpoint
func (o *Handler) GetEdApiWellKnownInfo(c *gin.Context) {
	// Simple well-known response
	response := map[string]interface{}{
		"version": "1.0.0",
		"status":  "SERVICE_STATUS_READY",
		"endpoints": map[string]string{
			"coservRequest": "/endorsement-distribution/v1/coserv/:query",
		},
		"supportedMediaTypes": []string{EdApiMediaType},
	}

	c.JSON(http.StatusOK, response)
}

// CoservRequest handles the main endorsement distribution endpoint
func (o *Handler) CoservRequest(c *gin.Context) {
	// Check Accept header
	offered := c.NegotiateFormat(EdApiMediaType)
	if offered != EdApiMediaType {
		o.reportProblem(c, http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", EdApiMediaType))
		return
	}

	// Get query parameter
	coservQuery := c.Param("query")
	if coservQuery == "" {
		o.reportProblem(c, http.StatusBadRequest, "missing query parameter")
		return
	}

	// Extract profile from Accept header if present
	acceptHeader := c.GetHeader("Accept")
	mediaType := EdApiMediaType
	if strings.Contains(acceptHeader, "profile=") {
		// Extract profile from Accept header
		// Format: application/coserv+cbor; profile="tag:arm.com,2023:cca_platform#1.0.0"
		parts := strings.Split(acceptHeader, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "profile=") {
				profile := strings.Trim(part[8:], `"`)
				mediaType = fmt.Sprintf(`%s; profile=%q`, EdApiMediaType, profile)
				break
			}
		}
	}

	o.Logger.Infow("Processing CoSERV request", "query", coservQuery, "mediaType", mediaType)

	// Get endorsements
	res, err := o.EndorsementDistributor.GetEndorsements(tenantID, coservQuery, mediaType)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errors.New("no artifacts found")) {
			status = http.StatusNotFound
		}

		o.reportProblem(c, status, err.Error())
		return
	}

	// Return the result
	c.Data(http.StatusOK, EdApiMediaType, res)
}

// reportProblem reports an error using RFC7807 problem format
func (o *Handler) reportProblem(c *gin.Context, status int, details ...string) {
	problem := map[string]interface{}{
		"status": status,
		"title":  http.StatusText(status),
	}

	if len(details) > 0 {
		problem["detail"] = strings.Join(details, ", ")
	}

	o.Logger.Errorw("API error", "status", status, "details", details)

	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, problem)
} 