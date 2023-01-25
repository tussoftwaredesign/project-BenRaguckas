package main

import (
	"github.com/gin-gonic/gin" // gin for rest api
)

type default_check struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"name"`
}

func test_function(c *gin.Context) {
	return
}
