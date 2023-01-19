package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"reflect"
	"strconv"
)

const (
	dbHost     = ""
	dbPort     = 5432
	dbUser     = ""
	dbPassword = ""
	dbName     = ""
)

type ignore struct{}

var Ignore ignore

func (ignore) Scan(_ interface{}) error {
	return nil
}

func dbInit(dbToOpen string) (*sql.DB, error) {
	source := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbToOpen)
	db, err := sql.Open("postgres", source)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, err
}

func dbQueryGetBenchmarks(db *sql.DB) (benchmarks, error) {
	var rows *sql.Rows
	var err error
	rows, err = db.Query(`
		SELECT *
		FROM benchmarks`)
	if rows == nil || err != nil {
		return benchmarks{}, err
	}
	var bs benchmarks
	for rows.Next() {
		var b benchmark
		err = rows.Scan(&b.BID, &b.Title, &b.Description, &b.Version, &b.Url)
		if err != nil {
			return benchmarks{}, err
		}
		bs.Benchmarks = append(bs.Benchmarks, b)
	}
	return bs, nil
}

func dbQueryGetBenchmark(db *sql.DB, bid string) (benchmark, msg, error) {
	idI, err := strconv.Atoi(bid)
	if err != nil {
		return benchmark{}, badReq400ErrMsg, err
	}
	row, err2 := db.Query(`
		SELECT *
		FROM benchmarks
		WHERE bid = $1`, idI)
	if row == nil {
		return benchmark{}, msg{}, nil
	}
	if err2 != nil {
		return benchmark{}, msg{}, err
	}
	var b benchmark
	for row.Next() {
		err = row.Scan(&b.BID, &b.Title, &b.Description, &b.Version, &b.Url)
	}
	if err != nil {
		return benchmark{}, msg{}, err
	}
	return b, msg{}, nil
}

func dbQueryPostBenchmark(db *sql.DB, title string, description string, version string, url string) (benchmark, msg, error) {
	err := db.Ping()
	if err != nil {
		return benchmark{}, dbErr500ErrMsg, err
	}
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO benchmarks(title, description, ver, url)
		VALUES($1, $2, $3, $4)
		RETURNING bid`, title, description, version, url).Scan(&lastInsertId)
	if err != nil {
		return benchmark{}, conflict409ErrMsg, err
	}
	return benchmark{BID: string(rune(lastInsertId)), Title: title, Description: description, Version: version, Url: url}, msg{}, nil
}

func dbQueryUpdateBenchmark(db *sql.DB, bid string, title string, description string, version string, url string) (benchmark, msg, error) {
	idI, err := strconv.Atoi(bid)
	if err != nil || title == "" {
		return benchmark{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return benchmark{}, dbErr500ErrMsg, err
	}
	var b benchmark
	err = db.QueryRow(`
		UPDATE benchmarks
		SET title = $1, description = $2, ver = $3, url = $4
		WHERE bid = $5
		RETURNING bid, title, description, ver, url`, idI, title, description, version, url).Scan(&b.BID, &b.Title, &b.Description, &b.Version, &b.Url)
	if b.BID == "" || err != nil {
		return benchmark{}, notFound404ErrMsg, err
	}
	return b, msg{}, nil
}

func dbQueryDeleteBenchmark(db *sql.DB, bid string) (msg, error) {
	idI, err := strconv.Atoi(bid)
	if err != nil {
		return badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return dbErr500ErrMsg, err
	}
	_, err = db.Exec(`
	DELETE
	FROM benchmarks
	WHERE bid = $1`, idI)
	if err != nil {
		return msg{}, err
	}
	return msg{}, nil
}

func dbQueryGetBlogposts(db *sql.DB, id string, tags []string, showAll bool) (blogposts, msg, error) {
	var rows *sql.Rows
	var ps blogposts
	var u user
	var err4 error
	err3 := db.Ping()
	if err3 != nil {
		return blogposts{}, dbErr500ErrMsg, err3
	}
	idI, err2 := strconv.Atoi(id)
	if err2 == nil {
		u, _, err4 = dbQueryGetUser(db, id, true)
		if err4 == nil {
			ps.User = u
		}
	}
	if showAll {
		if id == "" {
			if tags == nil {
				rows, err3 = db.Query(`
				SELECT *
				FROM blogposts`)
				for rows.Next() {
					var p blogpost
					err := rows.Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
					if err == nil {
						p.User, _, _ = dbQueryGetUser(db, p.User.UID, true)
					}
					ps.Blogposts = append(ps.Blogposts, p)
				}
			} else {
				for _, tag := range tags {
					rowsTemp, err5 := db.Query(`
					SELECT submission
					FROM posttags
					WHERE tag = $1`, tag)
					if err5 == nil {
						for rowsTemp.Next() {
							var i string
							var ii int
							_ = rowsTemp.Scan(&i)
							ii, err5 = strconv.Atoi(i)
							if err5 == nil {
								var p blogpost
								err := db.QueryRow(`
								SELECT *
								FROM blogposts
								WHERE bpid = $1`, ii).Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
								if err == nil {
									p.User, _, err = dbQueryGetUser(db, p.User.UID, true)
									if err == nil {
										for _, pFinal := range ps.Blogposts {
											if reflect.DeepEqual(p, pFinal) {
												continue
											}
											ps.Blogposts = append(ps.Blogposts, p)
										}
									}
								}
							}
						}
					}
				}
			}
		} else {
			ps.User, _, _ = dbQueryGetUser(db, id, true)
			if tags == nil {
				rows, err3 = db.Query(`
				SELECT *
				FROM blogposts
				WHERE "user" = $1`, idI)
				for rows.Next() {
					var p blogpost
					err := rows.Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
					if err == nil {
						ps.Blogposts = append(ps.Blogposts, p)
					}
				}
			} else {
				for _, tag := range tags {
					rowsTemp, err5 := db.Query(`
					SELECT submission
					FROM posttags
					WHERE tag = $1`, tag)
					if err5 == nil {
						for rowsTemp.Next() {
							var i string
							var ii int
							_ = rowsTemp.Scan(&i)
							ii, err5 = strconv.Atoi(i)
							if err5 != nil {
								var p blogpost
								err := db.QueryRow(`
								SELECT *
								FROM blogposts
								WHERE bpid = $1 AND "user" = $2`, ii, idI).Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
								if err == nil {
									for _, pFinal := range ps.Blogposts {
										if reflect.DeepEqual(p, pFinal) {
											continue
										}
										ps.Blogposts = append(ps.Blogposts, p)
									}
								}
							}
						}
					}
				}
			}
		}
	} else {
		if id == "" {
			if tags == nil {
				rows, err3 = db.Query(`
				SELECT *
				FROM blogposts
				WHERE verified = TRUE`)
				for rows.Next() {
					var p blogpost
					err := rows.Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
					if err == nil {
						p.User, _, _ = dbQueryGetUser(db, p.User.UID, true)
					}
					ps.Blogposts = append(ps.Blogposts, p)
				}
			} else {
				for _, tag := range tags {
					rowsTemp, err5 := db.Query(`
					SELECT submission
					FROM posttags
					WHERE tag = $1`, tag)
					if err5 == nil {
						for rowsTemp.Next() {
							var i string
							var ii int
							_ = rowsTemp.Scan(&i)
							ii, err5 = strconv.Atoi(i)
							if err5 == nil {
								var p blogpost
								err := db.QueryRow(`
								SELECT *
								FROM blogposts
								WHERE bpid = $1 AND verified = TRUE`, ii).Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
								if err == nil {
									p.User, _, err = dbQueryGetUser(db, p.User.UID, true)
									if err == nil {
										for _, pFinal := range ps.Blogposts {
											if reflect.DeepEqual(p, pFinal) {
												continue
											}
											ps.Blogposts = append(ps.Blogposts, p)
										}
									}
								}
							}
						}
					}
				}
			}
		} else {
			ps.User, _, _ = dbQueryGetUser(db, id, true)
			if tags == nil {
				rows, err3 = db.Query(`
				SELECT *
				FROM blogposts
				WHERE "user" = $1 AND verified = TRUE`, idI)
				for rows.Next() {
					var p blogpost
					err := rows.Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
					if err == nil {
						ps.Blogposts = append(ps.Blogposts, p)
					}
				}
			} else {
				for _, tag := range tags {
					rowsTemp, err5 := db.Query(`
					SELECT submission
					FROM posttags
					WHERE tag = $1`, tag)
					if err5 == nil {
						for rowsTemp.Next() {
							var i string
							var ii int
							_ = rowsTemp.Scan(&i)
							ii, err5 = strconv.Atoi(i)
							if err5 == nil {
								var p blogpost
								err := db.QueryRow(`
								SELECT *
								FROM blogposts
								WHERE bpid = $1 AND "user" = $2 AND verified = TRUE`, ii, idI).Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
								if err == nil {
									for _, pFinal := range ps.Blogposts {
										if reflect.DeepEqual(p, pFinal) {
											continue
										}
										ps.Blogposts = append(ps.Blogposts, p)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	for _, p := range ps.Blogposts {
		i, err := strconv.Atoi(p.User.UID)
		if err == nil {
			rows2, err5 := db.Query(`
			SELECT tag
			FROM posttags
			WHERE submission = $1`, i)
			if err5 == nil {
				for rows2.Next() {
					var t string
					err5 = rows2.Scan(&t)
					if err5 == nil {
						p.Tags = append(p.Tags, t)
					}
				}
			}
		}
	}
	return ps, msg{}, nil
}

func dbQueryGetBlogpost(db *sql.DB, bpid string) (blogpost, msg, error) {
	i, err := strconv.Atoi(bpid)
	if err != nil {
		return blogpost{}, badReq400ErrMsg, err
	} else {
		var p blogpost
		err = db.QueryRow(`
		SELECT *
		FROM blogposts
		WHERE bpid = $1`, i).Scan(&p.BPID, &p.Title, &p.Body, &p.Created, &p.IsVerified, &p.User.UID)
		if err == nil {
			p.User, _, _ = dbQueryGetUser(db, p.User.UID, true)
			rows2, err5 := db.Query(`
			SELECT tag
			FROM posttags
			WHERE submission = $1`, i)
			if err5 == nil {
				for rows2.Next() {
					var t string
					err5 = rows2.Scan(&t)
					if err5 == nil {
						p.Tags = append(p.Tags, t)
					}
				}
			}
			return p, msg{}, nil
		}
		return blogpost{}, notFound404ErrMsg, err
	}
}

func dbQueryPostBlogpost(db *sql.DB, title string, body string, tags []string, id string) (blogpost, msg, error) {
	// Post
	err0 := db.Ping()
	if err0 != nil {
		return blogpost{}, dbErr500ErrMsg, err0
	}
	idJ, err3 := strconv.Atoi(id)
	if err3 != nil {
		return blogpost{}, badReq400ErrMsg, err3
	}
	var lastInsertId int
	err := db.QueryRow(`
		INSERT INTO blogposts(title, body, user)
		VALUES($1, $2, $3)
		RETURNING sid`, title, body, idJ).Scan(&lastInsertId)
	if err != nil {
		return blogpost{}, conflict409ErrMsg, err
	}

	// Tags
	for _, tag := range tags {
		var applied string
		err = db.QueryRow(`
		INSERT INTO tags(tag)
		VALUES($1)
		ON CONFLICT DO NOTHING
		RETURNING tag`, tag).Scan(&applied)
		if err == nil && tag == applied {
			_ = db.QueryRow(`
			INSERT INTO POSTTAGS(submission, tag)
			VALUES($1, $2)
			ON CONFLICT DO NOTHING
			RETURNING tag`, lastInsertId, tag)
		}
	}
	return dbQueryGetBlogpost(db, string(rune(lastInsertId)))
}

func dbQueryUpdateBlogpost(db *sql.DB, bpid string, title string, body string, isVerified bool) (blogpost, msg, error) {
	// TODO tag update
	idI, err := strconv.Atoi(bpid)
	if err != nil {
		return blogpost{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return blogpost{}, dbErr500ErrMsg, err
	}
	var p blogpost
	err = db.QueryRow(`
	UPDATE blogposts
	SET title = $1, body = $2, verified = $3
	WHERE bpid = $4
	RETURNING bpid`, title, body, isVerified, idI).Scan(&p.BPID)
	if err != nil || p.BPID == "" {
		return blogpost{}, notFound404ErrMsg, err
	}
	return dbQueryGetBlogpost(db, p.BPID)
}

func dbQueryDeleteBlogpost(db *sql.DB, bpid string) msg {
	idI, err := strconv.Atoi(bpid)
	if err != nil {
		return badReq400ErrMsg
	}
	_, err = db.Exec(`
	DELETE
	FROM blogposts
	WHERE bpid = $1`, idI)
	if err != nil {
		return notFound404ErrMsg
	}
	return msg{}
}

// TODO blogpost comments

// TODO processors, memory

func dbQueryGetSubmissions(db *sql.DB, id string, bid string, showAll bool) (submissions, msg, error) {
	var rows *sql.Rows
	var ss submissions
	var u user
	var b benchmark
	var err4 error
	err3 := db.Ping()
	if err3 != nil {
		return submissions{}, dbErr500ErrMsg, err3
	}
	idJ, err := strconv.Atoi(bid)
	idI, err2 := strconv.Atoi(id)
	if err2 == nil {
		u, _, err4 = dbQueryGetUser(db, id, true)
		if err4 == nil {
			ss.User = u
		}
	}
	if err == nil {
		b, _, err4 = dbQueryGetBenchmark(db, bid)
		if err4 == nil {
			ss.Benchmark = b
		}
	}
	if showAll {
		if id == "" && bid == "" {
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions`)
		} else if id == "" {
			if err != nil {
				return submissions{}, badReq400ErrMsg, err
			}
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE benchmark = $1`, idJ)
		} else if bid == "" {
			if err2 != nil {
				return submissions{}, badReq400ErrMsg, err
			}
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE "user" = $1`, idI)
		} else {
			if err != nil || err2 != nil {
				return submissions{}, badReq400ErrMsg, err
			}
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE "user" = $1 AND benchmark = $2`, idI, idJ)
		}
	} else {
		if id == "" && bid == "" {
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE verified = TRUE`)
		} else if id == "" {
			if err != nil {
				return submissions{}, badReq400ErrMsg, err
			}
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE benchmark = $1 AND verified = TRUE`, idJ)
		} else if bid == "" {
			if err2 != nil {
				return submissions{}, badReq400ErrMsg, err
			}
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE "user" = $1 AND verified = TRUE`, idI)
		} else {
			if err != nil || err2 != nil {
				return submissions{}, badReq400ErrMsg, err
			}
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE "user" = $1 AND benchmark = $2 AND verified = TRUE`, idI, idJ)
		}
	}
	for rows.Next() {
		var s submission
		err = rows.Scan(&s.SID, &s.Result, &s.Processor.PID, &s.Memory.MID, &s.MemCount, &s.Created, &s.IsVerified, &s.Url, &s.Benchmark.BID, &s.User.UID, &s.Post.BPID, &s.Screenshot)
		if err != nil {
			return submissions{}, etcErr500ErrMsg, err
		}
		var p int
		p, err = strconv.Atoi(s.Processor.PID)
		if err == nil {
			err = db.QueryRow(`
			SELECT *
			FROM processors
			WHERE pid = $1`, p).Scan(&s.Processor.PID, &s.Processor.Model, &s.Processor.Lineup)
			if err != nil {
				err = db.QueryRow(`
				SELECT manufacturer
				FROM lineups
				WHERE title = $1`, s.Processor.Lineup).Scan(&s.Processor.Manufacturer)
			}
		}
		var m int
		m, err = strconv.Atoi(s.Memory.MID)
		if err == nil {
			err = db.QueryRow(`
			SELECT *
			FROM memories
			WHERE mid = $1`, m).Scan(&s.Memory.MID, &s.Memory.Model, &s.Memory.Lineup, &s.Memory.Capacity, &s.Memory.Generation)
			if err != nil {
				err = db.QueryRow(`
				SELECT manufacturer
				FROM lineups
				WHERE title = $1`, s.Memory.Lineup).Scan(&s.Memory.Manufacturer)
			}
		}
		b, _, err4 := dbQueryGetBenchmark(db, s.Benchmark.BID)
		if err4 == nil {
			s.Benchmark = b
		}
		u, _, err5 := dbQueryGetUser(db, s.User.UID, true)
		if err5 == nil {
			s.User = u
		}
		bp, _, err6 := dbQueryGetBlogpost(db, s.Post.BPID)
		if err6 == nil {
			s.Post = bp
		}
		ss.Submissions = append(ss.Submissions, s)
	}
	return ss, msg{}, nil
}

func dbQueryGetSubmission(db *sql.DB, sid string) (submission, msg, error) {
	idI, err := strconv.Atoi(sid)
	if err != nil {
		return submission{}, badReq400ErrMsg, err
	}
	row, err2 := db.Query(`
		SELECT *
		FROM submissions
		WHERE sid = $1`, idI)
	if row == nil || err2 != nil {
		return submission{}, notFound404ErrMsg, nil
	}
	var s submission
	for row.Next() {
		err = row.Scan(&s.SID, &s.Result, &s.Processor.PID, &s.Memory.MID, &s.MemCount, &s.Created, &s.IsVerified, &s.Url, &s.Benchmark.BID, &s.User.UID, &s.Post.BPID, &s.Screenshot)
	}
	if err != nil {
		return submission{}, etcErr500ErrMsg, err
	}
	b, _, err4 := dbQueryGetBenchmark(db, s.Benchmark.BID)
	if err4 == nil {
		s.Benchmark = b
	}
	u, _, err5 := dbQueryGetUser(db, s.User.UID, true)
	if err5 == nil {
		s.User = u
	}
	bp, _, err6 := dbQueryGetBlogpost(db, s.Post.BPID)
	if err6 == nil {
		s.Post = bp
	}
	return s, msg{}, nil
}

func dbQueryPostSubmission(db *sql.DB, result float32, processor string, memory string, memcount string, url string, benchmark benchmark, id string, screenshot string) (submission, msg, error) {
	err0 := db.Ping()
	if err0 != nil {
		return submission{}, dbErr500ErrMsg, err0
	}
	idI, err := strconv.Atoi(benchmark.BID)
	idJ, err3 := strconv.Atoi(id)
	idK, err4 := strconv.Atoi(processor)
	idL, err5 := strconv.Atoi(memory)
	idM, err6 := strconv.Atoi(memcount)
	if err != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		return submission{}, badReq400ErrMsg, err
	}
	b, _, err2 := dbQueryGetBenchmark(db, benchmark.BID)
	if err2 != nil || b.BID == "" {
		return submission{}, notFound404ErrMsg, err2
	}
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO submissions(result, processor, memory, memcount, url, benchmark, "user", screenshot)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING sid`, result, idK, idL, idM, url, idI, idJ, screenshot).Scan(&lastInsertId)
	if err != nil {
		return submission{}, conflict409ErrMsg, err
	}
	return dbQueryGetSubmission(db, string(rune(lastInsertId)))
}

func dbQueryUpdateSubmission(db *sql.DB, sid string, isVerified bool) (submission, msg, error) {
	idI, err := strconv.Atoi(sid)
	if err != nil {
		return submission{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return submission{}, dbErr500ErrMsg, err
	}
	var s submission
	err = db.QueryRow(`
	UPDATE submissions
	SET verified = $1
	WHERE sid = $2
	RETURNING sid`, isVerified, idI).Scan(&s.SID)
	if err != nil || s.SID == "" {
		return submission{}, notFound404ErrMsg, err
	}
	return dbQueryGetSubmission(db, s.SID)
}

func dbQueryDeleteSubmission(db *sql.DB, sid string) msg {
	idI, err := strconv.Atoi(sid)
	if err != nil {
		return badReq400ErrMsg
	}
	_, err = db.Exec(`
	DELETE
	FROM submissions
	WHERE sid = $1`, idI)
	if err != nil {
		return notFound404ErrMsg
	}
	return msg{}
}

func dbQueryGetSubmissionComments(db *sql.DB, sid string) (submissionComments, msg, error) {
	var rows *sql.Rows
	idI, err := strconv.Atoi(sid)
	if err != nil {
		return submissionComments{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return submissionComments{}, dbErr500ErrMsg, err
	}
	rows, err = db.Query(`
		SELECT *
		FROM subcomms
		WHERE submission = $1`, idI)
	var scs submissionComments
	for rows.Next() {
		var sc submissionComment
		err = rows.Scan(&sc.SCID, &sc.Body, Ignore, &sc.Created, Ignore, &sc.User.UID)
		if err != nil {
			return submissionComments{}, etcErr500ErrMsg, err
		}
		u, _, err3 := dbQueryGetUser(db, sc.User.UID, true)
		if err3 != nil {
			sc.User = u
		}
		scs.SubmissionComments = append(scs.SubmissionComments, sc)
	}
	s, _, err2 := dbQueryGetSubmission(db, sid)
	if err2 == nil {
		scs.Submission = s
	} else {
		return submissionComments{}, notFound404ErrMsg, err
	}
	return scs, msg{}, nil
}

func dbQueryGetSubmissionComment(db *sql.DB, sid string, cid string) (submissionComment, msg, error) {
	idI, err := strconv.Atoi(sid)
	idJ, err2 := strconv.Atoi(cid)
	if err2 != nil {
		return submissionComment{}, badReq400ErrMsg, err
	}
	var row *sql.Rows
	var err3 error
	if err != nil {
		row, err3 = db.Query(`
		SELECT *
		FROM subcomms
		WHERE scid = $1`, idJ)
	} else {
		row, err3 = db.Query(`
		SELECT *
		FROM subcomms
		WHERE submission = $1 AND scid = $2`, idI, idJ)
	}
	if row == nil || err3 != nil {
		return submissionComment{}, notFound404ErrMsg, err
	}
	var sc submissionComment
	for row.Next() {
		err = row.Scan(&sc.SCID, &sc.Body, &sc.Submission.SID, &sc.Created, Ignore, &sc.User.UID)
	}
	if err != nil {
		return submissionComment{}, etcErr500ErrMsg, err
	}
	s, _, err4 := dbQueryGetSubmission(db, sc.Submission.SID)
	if err4 == nil {
		sc.Submission = s
	}
	u, _, err5 := dbQueryGetUser(db, sc.User.UID, true)
	if err5 != nil {
		sc.User = u
	}
	return sc, msg{}, nil
}

func dbQueryPostSubmissionComment(db *sql.DB, sid string, body string, id string) (submissionComment, msg, error) {
	err := db.Ping()
	if err != nil {
		return submissionComment{}, dbErr500ErrMsg, err
	}
	idJ, err2 := strconv.Atoi(id)
	if err2 != nil {
		return submissionComment{}, badReq400ErrMsg, err
	}
	s, _, err4 := dbQueryGetSubmission(db, sid)
	if err4 != nil {
		return submissionComment{}, notFound404ErrMsg, err2
	} else if s.SID == "" {
		return submissionComment{}, notFound404ErrMsg, nil
	}
	idK, err5 := strconv.Atoi(s.SID)
	if err5 != nil {
		return submissionComment{}, etcErr500ErrMsg, err5
	}
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO subcomms(body, submission, "user")
		VALUES($1, $2, $3)
		RETURNING scid`, body, idK, idJ).Scan(&lastInsertId)
	if err != nil {
		return submissionComment{}, notFound404ErrMsg, nil
	}
	return dbQueryGetSubmissionComment(db, sid, string(rune(lastInsertId)))
}

func dbQueryDeleteSubmissionComment(db *sql.DB, sid string, cid string) error {
	err := db.Ping()
	if err != nil {
		return err
	}
	idJ, err2 := strconv.Atoi(cid)
	if err2 != nil {
		return err
	}
	_, err3 := strconv.Atoi(sid)
	if err3 != nil {
		s, _, err4 := dbQueryGetSubmission(db, sid)
		if err4 != nil || s.SID == "" {
			return err
		}
	}
	_, err = db.Exec(`
	DELETE
	FROM subcomms
	WHERE scid = $1`, idJ)
	if err != nil {
		return err
	}
	return nil
}

func dbQueryGetUsers(db *sql.DB, showAll bool) (users, error) {
	var rows *sql.Rows
	var err error
	if showAll {
		rows, err = db.Query(`
			SELECT uid, nick, avatar, aboutme, aboutblog, verified, privileged, created
			FROM users`)
	} else {
		rows, err = db.Query(`
			SELECT uid, nick, avatar, aboutme, aboutblog, verified, privileged, created
			FROM users
			WHERE verified = TRUE`)
	}
	if rows == nil || err != nil {
		return users{}, err
	}
	var us users
	for rows.Next() {
		var u user
		err = rows.Scan(&u.UID, &u.Nickname, &u.Avatar, &u.AboutMe, &u.AboutBlog, &u.IsVerified, &u.IsMod, &u.Created)
		if err != nil {
			return users{}, err
		}
		us.Users = append(us.Users, u)
	}
	return us, nil
}

func dbQueryGetUser(db *sql.DB, str string, isId bool) (user, msg, error) {
	idI, err := strconv.Atoi(str)
	if err != nil && isId {
		return user{}, badReq400ErrMsg, err
	}
	var row *sql.Rows
	if isId {
		row, err = db.Query(`
		SELECT uid, nick, avatar, aboutme, aboutblog, verified, privileged, created
		FROM users
		WHERE uid = $1`, idI)
	} else {
		row, err = db.Query(`
		SELECT uid, nick, avatar, aboutme, aboutblog, verified, privileged, created
		FROM users
		WHERE nick = $1`, str)
	}
	if row == nil {
		return user{}, notFound404ErrMsg, nil
	}
	if err != nil {
		return user{}, dbErr500ErrMsg, err
	}
	var u user
	for row.Next() {
		err = row.Scan(&u.UID, &u.Nickname, &u.Avatar, &u.AboutMe, &u.AboutBlog, &u.IsVerified, &u.IsMod, &u.Created)
	}
	if err != nil {
		return user{}, etcErr500ErrMsg, err
	}
	return u, msg{}, nil
}

func dbQueryPostUser(db *sql.DB, nickname string, email string, password string) (user, msg, error) {
	err := db.Ping()
	if err != nil {
		return user{}, dbErr500ErrMsg, err
	}
	hash := []byte(saltify(password))
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO users(email, nick, password)
		VALUES($1, $2, sha256($3))
		RETURNING id`, email, nickname, hash).Scan(&lastInsertId)
	if err != nil {
		return user{}, conflict409ErrMsg, err
	}
	return dbQueryGetUser(db, string(rune(lastInsertId)), true)
}

func dbQueryUpdateUser(db *sql.DB, id string, nick string, email string, avatar string, aboutMe string, aboutBlog string, isVerified bool) (user, msg, error) {
	idI, err := strconv.Atoi(id)
	if err != nil {
		return user{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return user{}, dbErr500ErrMsg, err
	}
	var uid int
	err = db.QueryRow(`
		UPDATE users
		SET nick = $1, email = $2, avatar = $3, aboutme = $4, aboutblog = $5, verified = $6
		WHERE id = $7
		RETURNING id`, nick, email, avatar, aboutMe, aboutBlog, isVerified, idI).Scan(&uid)
	if err != nil {
		return user{}, dbErr500ErrMsg, err
	}
	return dbQueryGetUser(db, string(rune(uid)), true)
}
