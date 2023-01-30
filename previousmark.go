// previousmark API

package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const path = "localhost:8080"

func main() {
	db, err := dbInit(dbName)
	if err != nil {
		fmt.Println("Error connecting to the database")
	}

	router := gin.Default()
	router.LoadHTMLFiles("index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Benchmarks
	router.GET("/api/benchmarks", getBenchmarks(db))
	router.GET("/api/benchmarks?page=&per_page=", getBenchmarks(db))
	router.GET("/api/benchmarks/:bid", getBenchmark(db))
	router.POST("/api/benchmarks", postBenchmark(db))
	router.PATCH("/api/benchmarks/:bid", updBenchmark(db))
	router.DELETE("/api/benchmarks/:bid", delBenchmark(db))

	// Blogposts and comments
	router.GET("/api/posts", getBlogposts(db))
	router.GET("/api/posts?uid=&page=&per_page=", getBlogposts(db))
	router.GET("/api/posts/:bpid", getBlogpost(db))
	router.POST("/api/posts", postBlogpost(db))
	router.PATCH("/api/posts/:bpid", updBlogpost(db))
	router.DELETE("/api/posts/:bpid", delBlogpost(db))
	// TODO comments

	// Submissions and comments
	router.GET("/api/submissions", getSubmissions(db))
	router.GET("/api/submissions?uid=&bid=&page=&per_page=", getSubmissions(db))
	router.GET("/api/submissions/:sid", getSubmission(db))
	router.POST("/api/submissions", postSubmission(db))
	router.PATCH("/api/submissions/:sid", updSubmission(db))
	router.DELETE("/api/submissions/:sid", delSubmission(db))
	router.GET("/api/submissions/:sid/comments", getSubmissionComments(db))
	router.GET("/api/submissions/:sid/comments?page=&per_page=", getSubmissionComments(db))
	router.GET("/api/submissions/:sid/comments/:scid", getSubmissionComment(db))
	router.POST("/api/submissions/:sid/comments", postSubmissionComment(db))
	router.DELETE("/api/submissions/:sid/comments/:scid", delSubmissionComment(db))

	// Auth
	router.POST("/api/auth", loginUser(db))
	router.DELETE("/api/auth/:uid", logoutUser(db))
	router.GET("/api/auth/reset", resetPass(db))
	router.POST("/api/auth/new", updPass(db))
	router.GET("/api/auth/refresh", refresh())

	// Users
	router.GET("/api/users", getUsers(db))
	router.GET("/api/users?page=&per_page=", getUsers(db))
	router.GET("/api/users/:uid", getUser(db))
	router.POST("/api/users", postUser(db))
	router.PATCH("/api/users/:uid", updUser(db))

	// Not allowed
	router.POST("/api/benchmarks/:bid", notAllowed405Msg("POST into resource"))
	router.PUT("/api/benchmarks/:bid", notAllowed405Msg("bulk update resource"))
	router.PUT("/api/benchmarks", notAllowed405Msg("bulk update collection"))
	router.PATCH("/api/benchmarks", notAllowed405Msg("PATCH whole collection"))
	router.DELETE("/api/benchmarks", notAllowed405Msg("DELETE whole collection"))
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
	// TODO

	err = router.Run(path)
	if err == nil {
		fmt.Println("Error initializing the router")
	}
}
