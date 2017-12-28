package server

import (
	"log"
	"net/http"
	"os"

	"github.com/de0gee/de0gee-data/src/database"
	"github.com/gin-gonic/gin"
)

func Run(port string) {
	r := gin.Default()
	r.GET("/ping", ping)
	r.HEAD("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.POST("/data", handlerData)  // typical data handler
	r.POST("/learn", handlerFIND) // backwards-compatible with FIND
	r.POST("/track", handlerFIND) // backwards-compatible with FIND
	r.GET("/location", handlerLocation)
	r.Run(":" + port) // listen and serve on 0.0.0.0:8080
}

func handlerLocation(c *gin.Context) {
	AddCORS(c)
	type Payload struct {
		Family string `json:"family" binding:"required"`
		Device string `json:"device" binding:"required"`
	}
	success := false
	var message string
	var p Payload
	if errBind := c.ShouldBindJSON(&p); errBind == nil {
		d, err := database.Open(p.Family)
		defer d.Close()
		if err != nil {
			message = err.Error()
		} else {
			s, err := d.GetLatest(p.Device)
			if err != nil {
				message = err.Error()
			} else {
				type ClassifyPayload struct {
					Sensor       database.SensorData `json:"sensor-data"`
					DataLocation string              `json:"data-location"`
				}
				var p2 ClassifyPayload
				p2.Sensor = s
				dir, err := os.Getwd()
				if err != nil {
					log.Fatal(err)
				}
				p2.DataLocation = dir
				c.JSON(http.StatusOK, gin.H{"message": "got latest", "success": true, "payload": p2})
				return
			}
		}
	} else {
		message = errBind.Error()
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "success": success})
}

func handlerData(c *gin.Context) {
	AddCORS(c)
	var err error
	var message string
	var d database.SensorData
	err = c.BindJSON(&d)
	if err == nil {
		err2 := d.Save()
		if err2 == nil {
			message = "inserted data"
		} else {
			err = err2
		}
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true})
	}
}

func handlerFIND(c *gin.Context) {
	AddCORS(c)
	var j database.FINDFingerprint
	var err error
	var message string
	err = c.BindJSON(&j)
	if err == nil {
		if c.Request.URL.Path == "/track" {
			j.Location = ""
		}
		d := j.Convert()
		err2 := d.Save()
		if err2 == nil {
			message = "inserted data"
		} else {
			err = err2
		}
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true})
	}
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
