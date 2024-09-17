// Copyright 2022 lukemakhanu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lukemakhanu/magic_carpet/cmd/apis/game_server/interfaces/infrustructure/auth"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := auth.TokenValid(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": http.StatusUnauthorized,
				"error":  err.Error(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// Avoid a large file from loading into memory
// If the file size is greater than 8MB dont allow it to even load into memory and waste our time.
func MaxSizeAllowed(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		buff, errRead := c.GetRawData()
		if errRead != nil {
			//c.JSON(http.StatusRequestEntityTooLarge,"too large")
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"status":     http.StatusRequestEntityTooLarge,
				"upload_err": "too large: upload an image less than 8MB",
			})
			c.Abort()
			return
		}
		buf := bytes.NewBuffer(buff)
		c.Request.Body = ioutil.NopCloser(buf)
	}
}

func IPWhiteListMiddleware(whitelist map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIP := c.ClientIP()
		log.Printf("IP received is : %s", userIP)

		if !whitelist[userIP] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "You are not authorized to access this resource!",
			})
		} else {
			c.Next()
		}
	}
}
