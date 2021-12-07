package main

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func isFulfillable(c *gin.Context, fulfillableExpectations []string) bool {
	if c.GetHeader("Expect") == "" {
		return true
	}
	for _, fulfillableExpectation := range fulfillableExpectations {
		if strings.Contains(c.GetHeader("Expect"), fulfillableExpectation) {
			return true
		}
	}
	return false
}

func isAcceptable(c *gin.Context, acceptableType string) bool {
	return c.GetHeader("Accept") == "" || strings.Contains(c.GetHeader("Accept"), acceptableType)
}

func isSupported(c *gin.Context, supportedType string) bool {
	return c.GetHeader("Content-Type") == "" || strings.Contains(c.GetHeader("Content-Type"), supportedType)
}
