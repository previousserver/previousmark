package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var expFail417ErrMsg = msg{Body: "Cannot provide expected status"}
var notAcc406ErrMsg = msg{Body: "Cannot provide resource in requested format"}
var notSupp415ErrMsg = msg{Body: "Cannot process resource in provided format"}
var notAuth401ErrMsg = msg{Body: "You are not logged in, your nickname and/or password is invalid, or your token is invalid. Your session might have expired. Please log in or modify request to continue"}
var noPerms403ErrMsg = msg{Body: "You do not have the permissions to perform this action"}
var reqErr500ErrMsg = msg{Body: "Could not parse request body"}
var dbErr500ErrMsg = msg{Body: "Internal database error. There might have been a conflict"}
var etcErr500ErrMsg = msg{Body: "Internal error"}
var badReq400ErrMsg = msg{Body: "Malformed request. Please modify request to continue"}
var notFound404ErrMsg = msg{Body: "Cannot find resource. Database might be offline"}
var conflict409ErrMsg = msg{Body: "Resource already exists. Alternatively, you might be logged in elsewhere"}
var unproc422ErrMsg = msg{Body: "Cannot process provided corrupt resource"}
var noCont204Msg = msg{Body: "Successfully logged out or deleted resource"}

func notAllowed405Msg(err string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, msg{Body: "Cannot " + err})
	}
}
