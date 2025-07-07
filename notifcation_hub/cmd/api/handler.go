package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TestingRequset struct{
	Name string `json:"name"`
}

func Testing(c *gin.Context){
	var req TestingRequset
	if err:= c.ShouldBindJSON(&req) ; err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":"Invalid request",
			"err":err.Error(),
		})
		return
	}
	log.Printf("%v",req)

	c.JSON(http.StatusAccepted,gin.H{
		"message":"Accepted",
	})
}