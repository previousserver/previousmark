package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const Key = ""
const Siz = 15

func getBenchmarks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "400", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
				a, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
				if !a {
					c.Header("WWW-Authenticate", "Bearer")
					c.JSON(http.StatusUnauthorized, msgM)
				}
			}
			page := c.DefaultQuery("page", "")
			perPage := c.DefaultQuery("per_page", "")
			var resource benchmarks
			//var msgM msg
			//var tokenT2 = tokenT
			var err error
			//if idMine != "" && tokenT != "" {
			//	tokenT2, msgM, _ = queryVerifyNoUpd(tokenT, idMine)
			//}
			if (page != "") != (perPage != "") {
				c.JSON(http.StatusBadRequest, badReq400ErrMsg)
			} else {
				resource, err = dbQueryGetBenchmarks(db)
				if err != nil {
					c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
				} else {
					if page != "" && perPage != "" {
						pageInt, err3 := strconv.Atoi(page)
						perPageInt, err4 := strconv.Atoi(perPage)
						if err3 == nil &&
							err4 == nil &&
							pageInt > 0 &&
							perPageInt > 0 &&
							len(resource.Benchmarks) > 0 &&
							(pageInt-1)*perPageInt >= 0 &&
							(pageInt-1)*perPageInt < len(resource.Benchmarks) {
							end := int(math.Min(float64(pageInt*perPageInt-1), float64(len(resource.Benchmarks)-1)))
							l := len(resource.Benchmarks)
							resource.Benchmarks = resource.Benchmarks[((pageInt - 1) * perPageInt):(end + 1)]
							if (pageInt-2)*perPageInt >= 0 {
								resource.Previous = path + "/api/benchmarks?page=" + strconv.Itoa(pageInt-1) + "&per_page=" + perPage
							}
							if pageInt*perPageInt < l {
								resource.Next = path + "/api/benchmarks?page=" + strconv.Itoa(pageInt+1) + "&per_page=" + perPage
							}
						} else {
							resource.Benchmarks = nil
						}
					}
					//resource.NewToken = token{Token: tokenT2, Status: msgM}
					c.JSON(http.StatusOK, resource)
				}
			}
		}
	}
}

func getBenchmark(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "400", "404", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			bid := c.Param("bid")
			//idMine := c.GetHeader("pm-id")
			//tokenT := c.GetHeader("Authorization")
			//if tokenT != "" {
			//	tokenT = tokenT[len("Bearer"):]
			//	a, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
			//	if !a {
			//		c.JSON(http.StatusUnauthorized, msgM)
			//	}
			//}
			var resource benchmark
			var err error
			var msgM msg
			//var tokenT2 = tokenT
			resource, msgM, err = dbQueryGetBenchmark(db, bid)
			//resource.NewToken = token{Token: tokenT2, Status: msgM}
			if msgM.Body == badReq400ErrMsg.Body {
				c.JSON(http.StatusBadRequest, msgM)
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
			} else if resource.BID == "" {
				c.JSON(http.StatusNotFound, notFound404ErrMsg)
			} else {
				//if idMine != "" && tokenT != "" {
				//	tokenT2, msgM, _ = dbSQueryVerifyToken(dbS, tokenT, false, idMine)
				//	resource.NewToken = token{Token: tokenT2, Status: msgM}
				//}
				c.JSON(http.StatusOK, resource)
			}
		}
	}
}

func postBenchmark(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "415", "401", "403", "500", "422", "409", "201"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if !isSupported(c, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, notSupp415ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			body, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusInternalServerError, reqErr500ErrMsg)
			} else {
				var newResource benchmark
				err = json.Unmarshal(body, &newResource)
				if err != nil {
					c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
				} else {
					newResource.Title = strings.Replace(
						strings.Replace(
							newResource.Title, "<", "", -1),
						">", "", -1)
					newResource.Description = strings.Replace(
						strings.Replace(
							newResource.Description, "<", "", -1),
						">", "", -1)
					newResource.Url = strings.Replace(
						strings.Replace(
							newResource.Url, "<", "", -1),
						">", "", -1)
					_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
					isMod := dbQueryIsMod(db, idMine)
					if msgM.Body == notAuth401ErrMsg.Body {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, msgM)
					} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
						c.JSON(http.StatusInternalServerError, msgM)
					} else if idMine != "" && tokenT != "" {
						if isMod {
							resource, msg2, err2 := dbQueryPostBenchmark(db, newResource.Title, newResource.Description, newResource.Version, newResource.Url)
							if msg2.Body == conflict409ErrMsg.Body {
								c.JSON(http.StatusConflict, conflict409ErrMsg)
							} else if msg2.Body == dbErr500ErrMsg.Body {
								c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
							} else if err2 != nil {
								c.JSON(http.StatusInternalServerError, etcErr500ErrMsg)
							} else {
								c.Header("Location", path+"/api/benchmarks/"+resource.BID)
								c.JSON(http.StatusCreated, resource)
							}
						} else {
							c.JSON(http.StatusForbidden, noPerms403ErrMsg)
						}
					} else {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
					}
				}
			}
		}
	}
}

func updBenchmark(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "415", "401", "400", "500", "422", "404", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if !isSupported(c, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, notSupp415ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			bid := c.Param("bid")
			idMine := c.GetHeader("x-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			var err error
			body, err2 := c.GetRawData()
			if err2 != nil {
				c.JSON(http.StatusInternalServerError, reqErr500ErrMsg)
			} else {
				var newResource benchmark
				err = json.Unmarshal(body, &newResource)
				if err != nil {
					c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
				} else {
					newResource.Title = strings.Replace(
						strings.Replace(
							newResource.Title, "<", "", -1),
						">", "", -1)
					newResource.Description = strings.Replace(
						strings.Replace(
							newResource.Description, "<", "", -1),
						">", "", -1)
					newResource.Url = strings.Replace(
						strings.Replace(
							newResource.Url, "<", "", -1),
						">", "", -1)
					_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
					isMod := dbQueryIsMod(db, idMine)
					if msgM.Body == notAuth401ErrMsg.Body {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, msgM)
					} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
						c.JSON(http.StatusInternalServerError, msgM)
					} else if idMine != "" && tokenT != "" {
						resource, msgM2, err3 := dbQueryGetBenchmark(db, bid)
						if msgM2 == badReq400ErrMsg {
							c.JSON(http.StatusBadRequest, msgM2)
						} else if err3 != nil {
							c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
						} else if resource.BID == "" {
							c.JSON(http.StatusNotFound, notFound404ErrMsg)
						} else {
							if isMod {
								resource, msgM, err3 = dbQueryUpdateBenchmark(db, bid, newResource.Title, newResource.Description, newResource.Version, newResource.Url)
							} else {
								c.JSON(http.StatusForbidden, noPerms403ErrMsg)
							}
							//resource.NewToken = token{Token: tokenT2, Status: msgM}
							if msgM.Body == notFound404ErrMsg.Body {
								c.JSON(http.StatusNotFound, msgM)
							} else if msgM.Body == badReq400ErrMsg.Body {
								c.JSON(http.StatusBadRequest, msgM)
							} else if err3 != nil {
								c.JSON(http.StatusInternalServerError, resource)
							} else {
								c.JSON(http.StatusOK, resource)
							}
						}
					} else {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
					}
				}
			}
		}
	}
}

func delBenchmark(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "401", "403", "400", "500", "404", "204"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			bid := c.Param("bid")
			idMine := c.GetHeader("x-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
			isMod := dbQueryIsMod(db, idMine)
			if msgM.Body == notAuth401ErrMsg.Body {
				c.Header("WWW-Authenticate", "Bearer")
				c.JSON(http.StatusUnauthorized, msgM)
			} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
				c.JSON(http.StatusInternalServerError, msgM)
			} else if idMine != "" && tokenT != "" {
				if isMod {
					msgM2, err := dbQueryDeleteBenchmark(db, bid)
					if msgM2 == badReq400ErrMsg {
						c.JSON(http.StatusBadRequest, msgM2)
					} else if err != nil {
						c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
					} else {
						c.JSON(http.StatusNoContent, nil)
					}
				} else {
					c.JSON(http.StatusForbidden, noPerms403ErrMsg)
				}
			} else {
				c.Header("WWW-Authenticate", "Bearer")
				c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
			}
		}
	}
}

func getBlogposts(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func getBlogpost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func postBlogpost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func updBlogpost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func delBlogpost(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

// TODO blog comments

func getSubmissions(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "400", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			id := c.DefaultQuery("uid", "")
			bid := c.DefaultQuery("bid", "")
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
				a, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
				if !a {
					c.JSON(http.StatusUnauthorized, msgM)
				}
			}
			page := c.DefaultQuery("page", "")
			perPage := c.DefaultQuery("per_page", "")
			var isMod = false
			var resource submissions
			var msgM msg
			var err error
			if idMine != "" && tokenT != "" {
				_, msgM, _ = queryVerifyNoUpd(tokenT, idMine)
				isMod = dbQueryIsMod(db, idMine)
			}
			if (page != "") != (perPage != "") {
				c.JSON(http.StatusBadRequest, badReq400ErrMsg)
			} else {
				resource, msgM, err = dbQueryGetSubmissions(db, id, bid, isMod)
				if msgM.Body == badReq400ErrMsg.Body {
					c.JSON(http.StatusBadRequest, msgM)
				} else if err != nil {
					c.JSON(http.StatusInternalServerError, msgM)
				} else {
					if page != "" && perPage != "" {
						pageInt, err3 := strconv.Atoi(page)
						perPageInt, err4 := strconv.Atoi(perPage)
						if err3 == nil &&
							err4 == nil &&
							pageInt != 0 &&
							perPageInt != 0 &&
							len(resource.Submissions) > 0 &&
							(pageInt-1)*perPageInt >= 0 &&
							(pageInt-1)*perPageInt < len(resource.Submissions) {
							end := int(math.Min(float64(pageInt*perPageInt-1), float64(len(resource.Submissions)-1)))
							l := len(resource.Submissions)
							resource.Submissions = resource.Submissions[((pageInt - 1) * perPageInt):(end + 1)]
							if (pageInt-2)*perPageInt >= 0 {
								resource.Previous = path + "/api/submissions/?id=" + id + "&bid=" + bid + "&page=" + strconv.Itoa(pageInt-1) + "&per_page=" + perPage
							}
							if pageInt*perPageInt < l {
								resource.Next = path + "/api/submissions/?id=" + id + "&bid=" + bid + "&page=" + strconv.Itoa(pageInt+1) + "&per_page=" + perPage
							}
						}
					}
					//resource.NewToken = token{Token: tokenT2, Status: msgM}
					c.JSON(http.StatusOK, resource)
				}
			}
		}
	}
}

func getSubmission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "400", "404", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			sid := c.Param("sid")
			//idMine := c.GetHeader("x-id")
			//tokenT := c.GetHeader("Authorization")
			//if tokenT != "" {
			//	tokenT = tokenT[len("Bearer"):]
			//}
			var resource submission
			var err error
			var msgM msg
			//var tokenT2 = tokenT
			resource, msgM, err = dbQueryGetSubmission(db, sid)
			//resource.NewToken = token{Token: tokenT2, Status: msgM}
			if msgM.Body == notFound404ErrMsg.Body {
				c.JSON(http.StatusNotFound, msgM)
			} else if msgM.Body == badReq400ErrMsg.Body {
				c.JSON(http.StatusBadRequest, msgM)
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, msgM)
			} else if resource.SID == "" {
				c.JSON(http.StatusNotFound, notFound404ErrMsg)
			} else {
				c.JSON(http.StatusOK, resource)
			}
		}
	}
}

func postSubmission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "415", "401", "400", "500", "422", "409", "404", "201"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if !isSupported(c, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, notSupp415ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			body, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusInternalServerError, reqErr500ErrMsg)
			} else {
				var newResource submission
				err = json.Unmarshal(body, &newResource)
				if err != nil {
					c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
				} else {
					newResource.Processor.PID = strings.Replace(
						strings.Replace(
							newResource.Processor.PID, "<", "", -1),
						">", "", -1)
					newResource.Memory.MID = strings.Replace(
						strings.Replace(
							newResource.Memory.MID, "<", "", -1),
						">", "", -1)
					newResource.Url = strings.Replace(
						strings.Replace(
							newResource.Url, "<", "", -1),
						">", "", -1)
					_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
					if msgM.Body == notAuth401ErrMsg.Body {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, msgM)
					} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
						c.JSON(http.StatusInternalServerError, msgM)
					} else if idMine != "" && tokenT != "" {
						a, b := extract(newResource.Screenshot)
						if a == -1 {
							c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
						}
						resource, msg2, err2 := dbQueryPostSubmission(db, a, newResource.Processor.PID, newResource.Memory.MID, newResource.MemCount, newResource.Url, newResource.Benchmark, newResource.User.UID, b)
						if msg2.Body == dbErr500ErrMsg.Body {
							c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
						} else if msg2.Body == badReq400ErrMsg.Body {
							c.JSON(http.StatusBadRequest, dbErr500ErrMsg)
						} else if msg2.Body == notFound404ErrMsg.Body {
							c.JSON(http.StatusNotFound, dbErr500ErrMsg)
						} else if err2 != nil {
							c.JSON(http.StatusInternalServerError, etcErr500ErrMsg)
						} else {
							u, _, err3 := dbQueryGetUser(db, resource.User.UID, true)
							if err3 == nil {
								resource.User = u
							}
							b, _, err4 := dbQueryGetBenchmark(db, newResource.Benchmark.BID)
							if err4 == nil {
								resource.Benchmark = b
							}
							//resource.NewToken = token{tokenT2, msgM}
							c.Header("Location", path+"/api/submissions/"+resource.SID)
							c.JSON(http.StatusCreated, resource)
						}
					} else {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
					}
				}
			}
		}
	}
}

func updSubmission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "415", "401", "400", "500", "422", "404", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if !isSupported(c, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, notSupp415ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			sid := c.Param("sid")
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			isMod := dbQueryIsMod(db, idMine)
			var err error
			body, err2 := c.GetRawData()
			if err2 != nil {
				c.JSON(http.StatusInternalServerError, token{tokenT, reqErr500ErrMsg})
			} else {
				var newResource submission
				err = json.Unmarshal(body, &newResource)
				if err != nil {
					c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
				} else {
					_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
					if msgM.Body == notAuth401ErrMsg.Body {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, msgM)
					} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
						c.JSON(http.StatusInternalServerError, msgM)
					} else if idMine != "" && tokenT != "" {
						resource, msgM2, err3 := dbQueryGetSubmission(db, sid)
						if msgM2.Body == badReq400ErrMsg.Body {
							c.JSON(http.StatusBadRequest, msgM2)
						} else if msgM2.Body == notFound404ErrMsg.Body {
							c.JSON(http.StatusNotFound, msgM2)
						} else if err3 != nil {
							c.JSON(http.StatusInternalServerError, msgM2)
						} else if resource.SID == "" {
							c.JSON(http.StatusNotFound, notFound404ErrMsg)
						} else {
							if isMod {
								resource, msgM, err3 = dbQueryUpdateSubmission(db, resource.SID, newResource.IsVerified)
							} else {
								c.JSON(http.StatusForbidden, noPerms403ErrMsg)
							}
							//resource.NewToken = token{Token: tokenT2, Status: msgM}
							if msgM.Body == notFound404ErrMsg.Body {
								c.JSON(http.StatusNotFound, msgM)
							} else if msgM.Body == badReq400ErrMsg.Body {
								c.JSON(http.StatusBadRequest, msgM)
							} else if err3 != nil {
								c.JSON(http.StatusInternalServerError, resource)
							} else {
								c.JSON(http.StatusOK, resource)
							}
						}
					} else {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
					}
				}
			}
		}
	}
}

func delSubmission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "401", "403", "500", "400", "404", "204"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			sid := c.Param("sid")
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
			isMod := dbQueryIsMod(db, idMine)
			if msgM.Body == notAuth401ErrMsg.Body {
				c.Header("WWW-Authenticate", "Bearer")
				c.JSON(http.StatusUnauthorized, msgM)
			} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
				c.JSON(http.StatusInternalServerError, msgM)
			} else if idMine != "" && tokenT != "" {
				if isMod {
					msgM2 := dbQueryDeleteSubmission(db, sid)
					if msgM2.Body == notFound404ErrMsg.Body {
						c.JSON(http.StatusNotFound, notFound404ErrMsg)
					} else if msgM2.Body == badReq400ErrMsg.Body {
						c.JSON(http.StatusBadRequest, badReq400ErrMsg)
					} else if msgM2.Body != "" {
						c.JSON(http.StatusInternalServerError, msgM2)
					} else {
						c.JSON(http.StatusNoContent, nil)
					}
				} else {
					c.JSON(http.StatusForbidden, noPerms403ErrMsg)
				}
			} else {
				c.Header("WWW-Authenticate", "Bearer")
				c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
			}
		}
	}
}

func getSubmissionComments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "404", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			sid := c.Param("sid")
			//idMine := c.GetHeader("pm-id")
			//tokenT := c.GetHeader("Authorization")
			//if tokenT != "" {
			//	tokenT = tokenT[len("Bearer"):]
			//}
			page := c.DefaultQuery("page", "")
			perPage := c.DefaultQuery("per_page", "")
			var resource submissionComments
			var msgM msg
			//var tokenT2 = tokenT
			var err error
			//if idMine != "" && tokenT != "" {
			//	tokenT2, msgM, _ = dbSQueryVerifyToken(dbS, tokenT, false, idMine)
			//}
			if (page != "") != (perPage != "") {
				c.JSON(http.StatusBadRequest, badReq400ErrMsg)
			} else {
				resource, msgM, err = dbQueryGetSubmissionComments(db, sid)
				if msgM.Body == badReq400ErrMsg.Body {
					c.JSON(http.StatusBadRequest, msgM)
				} else if msgM.Body == notFound404ErrMsg.Body {
					c.JSON(http.StatusNotFound, msgM)
				} else if err != nil {
					c.JSON(http.StatusInternalServerError, msgM)
				} else if resource.Submission.SID == "" {
					c.JSON(http.StatusNotFound, notFound404ErrMsg)
				} else {
					if page != "" && perPage != "" {
						pageInt, err3 := strconv.Atoi(page)
						perPageInt, err4 := strconv.Atoi(perPage)
						if err3 == nil &&
							err4 == nil &&
							pageInt != 0 &&
							perPageInt != 0 &&
							len(resource.SubmissionComments) > 0 &&
							(pageInt-1)*perPageInt >= 0 &&
							(pageInt-1)*perPageInt < len(resource.SubmissionComments) {
							end := int(math.Min(float64(pageInt*perPageInt-1), float64(len(resource.SubmissionComments)-1)))
							l := len(resource.SubmissionComments)
							resource.SubmissionComments = resource.SubmissionComments[((pageInt - 1) * perPageInt):(end + 1)]
							if (pageInt-2)*perPageInt >= 0 {
								resource.Previous = path + "/api/submissions/" + sid + "/comments?page=" + strconv.Itoa(pageInt-1) + "&per_page=" + perPage
							}
							if pageInt*perPageInt < l {
								resource.Next = path + "/api/submissions/" + sid + "/comments?page=" + strconv.Itoa(pageInt+1) + "&per_page=" + perPage
							}
						}
					}
					//resource.NewToken = token{Token: tokenT2, Status: msgM}
					c.JSON(http.StatusOK, resource)
				}
			}
		}
	}
}

func getSubmissionComment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "400", "404", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			sid := c.Param("sid")
			cid := c.Param("scid")
			//idMine := c.GetHeader("x-id")
			//tokenT := c.GetHeader("Authorization")
			//if tokenT != "" {
			//	tokenT = tokenT[len("Bearer"):]
			//}
			var resource submissionComment
			var err error
			var msgM msg
			//var tokenT2 = tokenT
			resource, msgM, err = dbQueryGetSubmissionComment(db, sid, cid)
			//resource.NewToken = token{Token: tokenT2, Status: msgM}
			if msgM.Body == notFound404ErrMsg.Body {
				c.JSON(http.StatusNotFound, msgM)
			} else if msgM.Body == badReq400ErrMsg.Body {
				c.JSON(http.StatusBadRequest, msgM)
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, msgM)
			} else if resource.Submission.SID == "" {
				c.JSON(http.StatusNotFound, notFound404ErrMsg)
			} else {
				c.JSON(http.StatusOK, resource)
			}
		}
	}
}

func postSubmissionComment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "415", "401", "400", "500", "400", "422", "404", "201"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if !isSupported(c, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, notSupp415ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			sid := c.Param("sid")
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			body, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusInternalServerError, reqErr500ErrMsg)
			} else {
				var newResource submissionComment
				err = json.Unmarshal(body, &newResource)
				if err != nil {
					c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
				} else {
					newResource.Body = strings.Replace(
						strings.Replace(
							newResource.Body, "<", "", -1),
						">", "", -1)
					_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
					if msgM.Body == notAuth401ErrMsg.Body {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, msgM)
					} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
						c.JSON(http.StatusInternalServerError, msgM)
					} else if idMine != "" && tokenT != "" {
						resource, msg2, err2 := dbQueryPostSubmissionComment(db, sid, newResource.Body, idMine)
						if msg2.Body == dbErr500ErrMsg.Body {
							c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
						} else if msg2.Body == badReq400ErrMsg.Body {
							c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
						} else if msg2.Body == notFound404ErrMsg.Body {
							c.JSON(http.StatusNotFound, dbErr500ErrMsg)
						} else if err2 != nil {
							c.JSON(http.StatusInternalServerError, etcErr500ErrMsg)
						} else {
							u, _, err3 := dbQueryGetUser(db, resource.User.UID, true)
							if err3 == nil {
								resource.User = u
							}
							//resource.NewToken = token{tokenT2, msgM}
							c.Header("Location", path+"/api/submissions/"+sid+"/comments/"+resource.SCID)
							c.JSON(http.StatusCreated, resource)
						}
					} else {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
					}
				}
			}
		}
	}
}

func delSubmissionComment(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "415", "401", "403", "500", "400", "404", "204"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if !isSupported(c, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, notSupp415ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			sid := c.Param("sid")
			cid := c.Param("scid")
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
			isMod := dbQueryIsMod(db, idMine)
			if msgM.Body == notAuth401ErrMsg.Body {
				c.Header("WWW-Authenticate", "Bearer")
				c.JSON(http.StatusUnauthorized, msgM)
			} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
				c.JSON(http.StatusInternalServerError, msgM)
			} else if idMine != "" && tokenT != "" {
				sc, msgM2, err := dbQueryGetSubmissionComment(db, sid, cid)
				if msgM2.Body == badReq400ErrMsg.Body {
					c.JSON(http.StatusBadRequest, msgM2)
				} else if msgM2.Body == notFound404ErrMsg.Body {
					c.JSON(http.StatusNotFound, msgM2)
				} else if err != nil {
					c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
				} else if sc.SCID == "" {
					c.JSON(http.StatusNotFound, notFound404ErrMsg)
				} else if isMod || sc.User.UID == idMine {
					err = dbQueryDeleteSubmissionComment(db, sid, cid)
					if err != nil {
						c.JSON(http.StatusNotFound, notFound404ErrMsg)
					} else {
						c.JSON(http.StatusNoContent, nil)
					}
				} else {
					c.JSON(http.StatusForbidden, noPerms403ErrMsg)
				}
			} else {
				c.Header("WWW-Authenticate", "Bearer")
				c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
			}
		}
	}
}

func loginUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "422", "401", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			creds := c.GetHeader("Authorization")
			if creds != "" {
				creds = creds[len("Basic "):]
			}
			credsDec, err2 := base64.StdEncoding.DecodeString(creds)
			if err2 != nil {
				c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
			} else {
				credsDecS := strings.SplitN(string(credsDec), ":", 2)
				tokenT, msgM, _ := dbQueryLoginUser(credsDecS[0], credsDecS[1], db)
				if msgM.Body == notAuth401ErrMsg.Body {
					c.Header("WWW-Authenticate", "Basic")
					c.JSON(http.StatusUnauthorized, msgM)
				} else if msgM.Body == conflict409ErrMsg.Body {
					c.JSON(http.StatusConflict, msgM)
				} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
					c.JSON(http.StatusInternalServerError, msgM)
				} else {
					c.JSON(http.StatusOK, token{Token: tokenT, Status: msg{Body: "Successfully logged in"}})
				}
			}
		}
	}
}

func logoutUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "422", "401", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			idMine := c.Param("uid")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			_, msgM, _ := queryVerifyToken(tokenT, true, idMine)
			if msgM.Body == notAuth401ErrMsg.Body {
				c.Header("WWW-Authenticate", "Bearer")
				c.JSON(http.StatusUnauthorized, msgM)
			} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
				c.JSON(http.StatusInternalServerError, msgM)
			} else {
				c.JSON(http.StatusNoContent, nil)
			}
		}
	}
}

func resetPass(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func updPass(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func refresh() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func getUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "400", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
				a, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
				if !a {
					c.JSON(http.StatusUnauthorized, msgM)
				}
			}
			page := c.DefaultQuery("page", "")
			perPage := c.DefaultQuery("per_page", "")
			var resource users
			//var msgM msg
			//var tokenT2 = tokenT
			var err error
			var isMod = false
			//resource.NewToken = token{Token: tokenT2, Status: msgM}
			if idMine != "" && tokenT != "" {
				//tokenT2, msgM, _ = dbSQueryVerifyToken(dbS, tokenT, false, idMine)
				isMod = dbQueryIsMod(db, idMine)
			}
			if (page != "") != (perPage != "") {
				c.JSON(http.StatusBadRequest, badReq400ErrMsg)
			} else {
				resource, err = dbQueryGetUsers(db, isMod)
				if err != nil {
					c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
				} else {
					if page != "" && perPage != "" {
						pageInt, err3 := strconv.Atoi(page)
						perPageInt, err4 := strconv.Atoi(perPage)
						if err3 == nil &&
							err4 == nil &&
							pageInt != 0 &&
							perPageInt != 0 &&
							len(resource.Users) > 0 &&
							(pageInt-1)*perPageInt >= 0 &&
							(pageInt-1)*perPageInt < len(resource.Users) {
							end := int(math.Min(float64(pageInt*perPageInt-1), float64(len(resource.Users)-1)))
							l := len(resource.Users)
							resource.Users = resource.Users[((pageInt - 1) * perPageInt):(end + 1)]
							if (pageInt-2)*perPageInt >= 0 {
								resource.Previous = path + "/api/users?page=" + strconv.Itoa(pageInt-1) + "&per_page=" + perPage
							}
							if pageInt*perPageInt < l {
								resource.Next = path + "/api/users?page=" + strconv.Itoa(pageInt+1) + "&per_page=" + perPage
							}
						}
					}
					//resource.NewToken = token{Token: tokenT2, Status: msgM}
					c.JSON(http.StatusOK, resource)
				}
			}
		}
	}
}

func getUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "500", "404", "400", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			id := c.Param("uid")
			//idMine := c.GetHeader("x-id")
			//tokenT := c.GetHeader("Authorization")
			//if tokenT != "" {
			//	tokenT = tokenT[len("Bearer"):]
			//}
			var resource user
			var err error
			var msgM msg
			//var tokenT2 = tokenT
			resource, msgM, err = dbQueryGetUser(db, id, true)
			//resource.NewToken = token{Token: tokenT2, Status: msgM}
			if msgM.Body == notFound404ErrMsg.Body {
				c.JSON(http.StatusNotFound, msgM)
			} else if msgM.Body == badReq400ErrMsg.Body {
				c.JSON(http.StatusBadRequest, msgM)
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, msgM)
			} else if resource.UID == "" {
				c.JSON(http.StatusNotFound, notFound404ErrMsg)
			} else {
				c.JSON(http.StatusOK, resource)
			}
		}
	}
}

func postUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "415", "500", "422", "409", "201"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if !isSupported(c, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, notSupp415ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			body, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusInternalServerError, reqErr500ErrMsg)
			} else {
				var newResource user
				err = json.Unmarshal(body, &newResource)
				if err != nil {
					c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
				} else {
					newResource.Nickname = regexp.MustCompile(`[^a-zA-Z0-9_-]+`).ReplaceAllString(newResource.Nickname, "")
					resource, msg2, err2 := dbQueryPostUser(db, newResource.Nickname, newResource.Email, newResource.Nuke)
					if msg2.Body == conflict409ErrMsg.Body {
						c.JSON(http.StatusConflict, conflict409ErrMsg)
					} else if msg2.Body == dbErr500ErrMsg.Body {
						c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
					} else if err2 != nil {
						c.JSON(http.StatusInternalServerError, etcErr500ErrMsg)
					} else {
						c.Header("Location", path+"/api/users/"+newResource.UID)
						c.JSON(http.StatusCreated, resource)
					}
				}
			}
		}
	}
}

func updUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFulfillable(c, []string{"417", "406", "415", "401", "403", "500", "400", "422", "404", "200"}) {
			c.JSON(http.StatusExpectationFailed, expFail417ErrMsg)
		} else if !isAcceptable(c, "application/json") {
			c.JSON(http.StatusNotAcceptable, notAcc406ErrMsg)
		} else if !isSupported(c, "application/json") {
			c.JSON(http.StatusUnsupportedMediaType, notSupp415ErrMsg)
		} else if db.Ping() != nil {
			c.JSON(http.StatusInternalServerError, dbErr500ErrMsg)
		} else {
			id := c.Param("uid")
			idMine := c.GetHeader("pm-id")
			tokenT := c.GetHeader("Authorization")
			if tokenT != "" {
				tokenT = tokenT[len("Bearer"):]
			}
			var err error
			body, err2 := c.GetRawData()
			if err2 != nil {
				c.JSON(http.StatusInternalServerError, reqErr500ErrMsg)
			} else {
				var newResource user
				err = json.Unmarshal(body, &newResource)
				if err != nil {
					c.JSON(http.StatusUnprocessableEntity, unproc422ErrMsg)
				} else {
					newResource.Nickname = regexp.MustCompile(`[^a-zA-Z0-9_-]+`).ReplaceAllString(newResource.Nickname, "")
					newResource.AboutMe = strings.Replace(
						strings.Replace(
							newResource.AboutMe, "<", "", -1),
						">", "", -1)
					newResource.AboutBlog = strings.Replace(
						strings.Replace(
							newResource.AboutBlog, "<", "", -1),
						">", "", -1)
					_, msgM, _ := queryVerifyNoUpd(tokenT, idMine)
					isMod := dbQueryIsMod(db, idMine)
					if msgM.Body == notAuth401ErrMsg.Body {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, msgM)
					} else if msgM.Body == etcErr500ErrMsg.Body || msgM.Body == dbErr500ErrMsg.Body {
						c.JSON(http.StatusInternalServerError, msgM)
					} else if idMine != "" && tokenT != "" {
						resource, msgM2, err3 := dbQueryGetUser(db, id, true)
						if msgM2.Body == notFound404ErrMsg.Body {
							c.JSON(http.StatusNotFound, msgM2)
						} else if msgM2.Body == badReq400ErrMsg.Body {
							c.JSON(http.StatusBadRequest, msgM2)
						} else if err3 != nil {
							c.JSON(http.StatusInternalServerError, msgM2)
						} else if resource.UID == "" {
							c.JSON(http.StatusNotFound, notFound404ErrMsg)
						} else {
							if isMod {
								if id == idMine {
									resource, msgM, err3 = dbQueryUpdateUser(db, id, newResource.Nickname, newResource.Email, newResource.Avatar, newResource.AboutMe, newResource.AboutBlog, resource.IsVerified)
								} else {
									resource, msgM, err3 = dbQueryUpdateUser(db, id, resource.Nickname, resource.Email, resource.Avatar, resource.AboutMe, resource.AboutBlog, newResource.IsVerified)
								}
							} else {
								if id == idMine {
									resource, msgM, err3 = dbQueryUpdateUser(db, id, newResource.Nickname, newResource.Email, newResource.Avatar, newResource.AboutMe, newResource.AboutBlog, resource.IsVerified)
								} else {
									c.JSON(http.StatusForbidden, noPerms403ErrMsg)
									return
								}
							}
							//resource.NewToken = token{Token: tokenT2, Status: msgM}
							if err3 != nil {
								c.JSON(http.StatusInternalServerError, resource)
							} else {
								c.JSON(http.StatusOK, resource)
							}
						}
					} else {
						c.Header("WWW-Authenticate", "Bearer")
						c.JSON(http.StatusUnauthorized, notAuth401ErrMsg)
					}
				}
			}
		}
	}
}

func extract(submission string) (float32, string) {
	r, err := base64.StdEncoding.DecodeString(submission)
	if err == nil {
		r = xor(r, []byte(Key))
		//i := int(r[0])
		float, _ := strconv.ParseFloat(string(r[1:Siz]), 32)
		shot := base64.StdEncoding.EncodeToString(r[Siz+1:])
		return float32(float), shot
	}
	return -1, ""
}

func xor(x []byte, y []byte) []byte {
	yy := y
	for i := math.Ceil(float64(len(x) / len(y))); i > 0; i-- {
		yy = append(yy, y...)
	}
	for j := range x {
		x[j] = x[j] ^ yy[j]
	}
	return x
}
