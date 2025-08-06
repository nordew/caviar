package rest

import (
	"fmt"

	"caviar/pkg/apperror"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
		)
	})
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			h.handleError(c, apperror.New(apperror.CodeUnauthorized, "Authorization header is required"))
			c.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			h.handleError(c, apperror.New(apperror.CodeUnauthorized, "Invalid authorization format. Use 'Bearer <token>'"))
			c.Abort()
			return
		}

		token := authHeader[7:]
		
		// Debug logging for development
		if !h.isProd {
			h.logger.Debug("Auth validation",
				zap.String("provided_token", token),
				zap.String("expected_secret", h.authSecret),
				zap.Bool("tokens_match", token == h.authSecret),
				zap.Int("provided_length", len(token)),
				zap.Int("expected_length", len(h.authSecret)))
		}
		
		if token != h.authSecret {
			h.handleError(c, apperror.New(apperror.CodeUnauthorized, "Invalid authorization token"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *Handler) ErrorMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			h.handleError(c, err)
		} else {
			h.handleError(c, apperror.New(apperror.CodeInternal, "Internal server error"))
		}
	})
} 