package api

import (
	"github.com/RakibRahman/fincore-api/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  *sqlc.Store // db layer
	router *gin.Engine // GIN HTTP Router
}

func NewServer(store *sqlc.Store) *Server {
	server := &Server{
		store: store,
	}
	router := gin.Default()
	server.setupRoutes(router)
	server.router = router
	return server
}

func (server *Server) setupRoutes(router *gin.Engine) {
	// User Routes
	router.POST("/users", server.createUser)
	router.GET("/users", server.listUsers)
	router.GET("/users/:id", server.getUser)
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
