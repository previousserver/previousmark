// previousmark API

package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

const path = "localhost:8080"

func main() {
	db, err := dbInit(dbName)
	dbS, err2 := dbInit(dbNameS)
	if err != nil || err2 != nil {
		fmt.Println("Error connecting to the database")
	}
	_ = dbS.Ping()

	router := gin.Default()

	// Benchmarks and comments
	router.GET("/api/benchmarks", getBenchmarks(db, dbS))
	router.GET("/api/benchmarks?page=&per_page=", getBenchmarks(db, dbS))
	router.GET("/api/benchmarks/:bid", getBenchmark(db, dbS))
	router.POST("/api/benchmarks", postBenchmark(db, dbS))
	router.PATCH("/api/benchmarks/:bid", updBenchmark(db, dbS))
	router.DELETE("/api/benchmarks/:bid", delBenchmark(db, dbS))
	router.GET("/api/benchmarks/:bid/comments", getBenchmarkComments(db, dbS))
	router.GET("/api/benchmarks/:bid/comments?page=&per_page=", getBenchmarkComments(db, dbS))
	router.GET("/api/benchmarks/:bid/comments/:cid", getBenchmarkComment(db, dbS))
	router.POST("/api/benchmarks/:bid/comments", postBenchmarkComment(db, dbS))
	router.DELETE("/api/benchmarks/:bid/comments/:cid", delBenchmarkComment(db, dbS))

	// Submissions and comments
	router.GET("/api/submissions", getSubmissions(db, dbS))
	router.GET("/api/submissions?id=&bid=&page=&per_page=", getSubmissions(db, dbS))
	router.GET("/api/submissions/:sid", getSubmission(db, dbS))
	router.POST("/api/submissions", postSubmission(db, dbS))
	router.PATCH("/api/submissions/:sid", updSubmission(db, dbS))
	router.DELETE("/api/submissions/:sid", delSubmission(db, dbS))
	router.GET("/api/submissions/:sid/comments", getSubmissionComments(db, dbS))
	router.GET("/api/submissions/:sid/comments?page=&per_page=", getSubmissionComments(db, dbS))
	router.GET("/api/submissions/:sid/comments/:cid", getSubmissionComment(db, dbS))
	router.POST("/api/submissions/:sid/comments", postSubmissionComment(db, dbS))
	router.DELETE("/api/submissions/:sid/comments/:cid", delSubmissionComment(db, dbS))

	// Users
	router.POST("/api/auth", loginUser(db, dbS))
	router.DELETE("/api/auth/:id", logoutUser(db, dbS))
	router.GET("/api/users", getUsers(db, dbS))
	router.GET("/api/users?page=&per_page=", getUsers(db, dbS))
	router.GET("/api/users/:id", getUser(db, dbS))
	router.POST("/api/users", postUser(db, dbS))
	router.PATCH("/api/users/:id", updUser(db, dbS))
	router.DELETE("/api/users/:id", delUser(db, dbS))

	// Not allowed
	router.POST("/api/benchmarks/:bid", notAllowed405Msg("POST into resource"))
	router.PUT("/api/benchmarks/:bid", notAllowed405Msg("bulk update resource"))
	router.PUT("/api/benchmarks", notAllowed405Msg("bulk update collection"))
	router.PATCH("/api/benchmarks", notAllowed405Msg("PATCH whole collection"))
	router.DELETE("/api/benchmarks", notAllowed405Msg("DELETE whole collection"))
	router.POST("/api/benchmarks/:bid/comments/:cid", notAllowed405Msg("POST into resource"))
	router.PUT("/api/benchmarks/:bid/comments/:cid", notAllowed405Msg("bulk update resource"))
	router.PUT("/api/benchmarks/:bid/comments", notAllowed405Msg("bulk update collection"))
	router.PATCH("/api/benchmarks/:bid/comments", notAllowed405Msg("PATCH whole collection"))
	router.DELETE("/api/benchmarks/:bid/comments", notAllowed405Msg("DELETE whole collection"))
	router.POST("/api/submissions/:sid", notAllowed405Msg("POST into resource"))
	router.PUT("/api/submissions/:sid", notAllowed405Msg("bulk update resource"))
	router.PUT("/api/submissions", notAllowed405Msg("bulk update collection"))
	router.PATCH("/api/submissions", notAllowed405Msg("PATCH whole collection"))
	router.DELETE("/api/submissions", notAllowed405Msg("DELETE whole collection"))
	router.POST("/api/submissions/:sid/comments/:cid", notAllowed405Msg("POST into resource"))
	router.PUT("/api/submissions/:sid/comments/:cid", notAllowed405Msg("bulk update resource"))
	router.PUT("/api/submissions/:sid/comments", notAllowed405Msg("bulk update collection"))
	router.PATCH("/api/submissions/:sid/comments", notAllowed405Msg("PATCH whole collection"))
	router.DELETE("/api/submissions/:sid/comments", notAllowed405Msg("DELETE whole collection"))
	router.GET("/api/auth", notAllowed405Msg("GET authentication"))
	router.GET("/api/auth/:id", notAllowed405Msg("GET authentication"))
	router.PUT("/api/auth", notAllowed405Msg("bulk update authentication"))
	router.PUT("/api/auth/:id", notAllowed405Msg("bulk update authentication"))
	router.PATCH("/api/auth", notAllowed405Msg("PATCH authentication"))
	router.PATCH("/api/auth/:id", notAllowed405Msg("PATCH authentication"))
	router.DELETE("/api/auth", notAllowed405Msg("DELETE authentication mechanism"))
	router.POST("/api/users/:id", notAllowed405Msg("POST into resource"))
	router.PUT("/api/users/:id", notAllowed405Msg("bulk update resource"))
	router.PUT("/api/users", notAllowed405Msg("bulk update collection"))
	router.PATCH("/api/users", notAllowed405Msg("PATCH whole collection"))
	router.DELETE("/api/users", notAllowed405Msg("DELETE whole collection"))

	err = router.Run(path)
	if err == nil {
		fmt.Println("Error initializing the router")
	}
}
