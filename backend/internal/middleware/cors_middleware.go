package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

// CORSMiddleware creates CORS middleware for the given frontend URL
func CORSMiddleware(frontendURL string) gin.HandlerFunc {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			frontendURL,
			"http://localhost:5173",
			"http://localhost:3000",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           int(12 * time.Hour / time.Second),
	})

	return func(c *gin.Context) {
		corsHandler.HandlerFunc(c.Writer, c.Request)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

