package api

import (
	"net/http"

	"github.com/RakibRahman/fincore-api/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type listUsersRequest struct {
	Page  int32 `form:"page" binding:"min=0"`
	Limit int32 `form:"limit" binding:"min=1"`
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req sqlc.CreateUserParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (server *Server) getUser(ctx *gin.Context) {
	userIdStr := ctx.Param("id")

	// Convert string to pgtype.UUID
	var userId pgtype.UUID
	err := userId.Scan(userIdStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (server *Server) listUsers(ctx *gin.Context) {
	var req listUsersRequest

	// Set defaults first
	req.Page = 0
	req.Limit = 20

	// Bind will override defaults if params provided
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	offset := req.Limit * req.Page
	users, err := server.store.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  req.Limit,
		Offset: offset,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}
