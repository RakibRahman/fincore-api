package api

import (
	"net/http"

	"github.com/RakibRahman/fincore-api/db/sqlc"
	"github.com/gin-gonic/gin"
)

type listUsersRequest struct {
	Page  int32 `form:"page" binding:"min=0"`
	Limit int32 `form:"limit" binding:"min=1"`
}

type getUserByEmail struct {
	Email string
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

func (server *Server) getUsers(ctx *gin.Context) {
	email := ctx.Query("email")

	if email != "" {
		server.getUserByEmail(ctx, email)
		return
	}

	server.listUsers(ctx)
}

func (server *Server) getUserByEmail(ctx *gin.Context, email string) {
	user, err := server.store.GetUserByEmail(ctx, email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (server *Server) listUsers(ctx *gin.Context) {
	var req listUsersRequest
	req.Page = 0
	req.Limit = 20

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

	ctx.JSON(http.StatusOK, gin.H{"data": users})
}
