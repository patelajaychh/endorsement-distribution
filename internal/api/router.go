package api

import (
	"path"

	"github.com/gin-gonic/gin"
)

const (
	edApiPath = "/endorsement-distribution/v1"
)

func NewRouter(handler *Handler) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Well-known endpoint
	router.GET("/.well-known/veraison/endorsement-distribution", handler.GetEdApiWellKnownInfo)

	// Main CoSERV endpoint
	coservEndpoint := path.Join(edApiPath, "coserv/:query")
	router.GET(coservEndpoint, handler.CoservRequest)

	return router
} 