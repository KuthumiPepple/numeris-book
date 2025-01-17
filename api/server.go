package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kuthumipepple/numeris-book/db"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	return &Server{
		store:  store,
		router: gin.Default(),
	}
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
