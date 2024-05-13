package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Homepage(c *gin.Context) {
	c.HTML(http.StatusOK, "/tmpl/index.tmpl", gin.H{})

}
