package domain

import "github.com/gin-gonic/gin"

const (
	HeaderContentType          = "Content-Type"
	ContentTypeApplicationJson = "application/json"
)

type Error struct {
	Error string `json:"message"`
}

type RestHandler interface {
	Register(router *gin.Engine)
}
