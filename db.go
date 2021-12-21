package main

import (
	"database/sql"
	"fmt"
	_ "lib/pq"
	"strconv"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "wayko2x8S"
	dbName     = "postgres"
	dbNameS    = "secrets"
)

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
		err = rows.Scan(&b.BID, &b.Title, &b.Icon, &b.Description, &b.Metric, &b.Rating, &b.RatingCount)
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
		err = row.Scan(&b.BID, &b.Title, &b.Icon, &b.Description, &b.Metric, &b.Rating, &b.RatingCount)
	}
	if err != nil {
		return benchmark{}, msg{}, err
	}
	return b, msg{}, nil
}

func dbQueryPostBenchmark(db *sql.DB, title string, icon string, description string, metric string) (benchmark, msg, error) {
	err := db.Ping()
	if err != nil {
		return benchmark{}, dbErr500ErrMsg, err
	}
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO benchmarks(title, icon, description, metric)
		VALUES($1, $2, $3, $4)
		RETURNING bid`, title, icon, description, metric).Scan(&lastInsertId)
	if err != nil {
		return benchmark{}, conflict409ErrMsg, err
	}
	return benchmark{BID: string(rune(lastInsertId)), Title: title, Icon: icon, Description: description, Metric: metric}, msg{}, nil
}

func dbQueryUpdateBenchmark(db *sql.DB, id string, bid string, title string, icon string, description string, metric string, rating float32, ratingCount int) (benchmark, msg, error) {
	idI, err := strconv.Atoi(bid)
	var idJ int
	idJ, err = strconv.Atoi(id)
	if err != nil {
		return benchmark{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return benchmark{}, dbErr500ErrMsg, err
	}
	var b benchmark
	row, err2 := db.Query(`
		SELECT *
		FROM benchmark_ratings
		WHERE benchmark = $1 AND "user" = $2`, idI, idJ)
	if row != nil && err2 == nil {
		err = db.QueryRow(`
		UPDATE benchmarks
		SET title = $1, icon = $2, description = $3, metric = $4
		WHERE bid = $5
		RETURNING bid, title, icon, description, metric, rating, rating_count`, title, icon, description, metric, idI).Scan(&b.BID, &b.Title, &b.Icon, &b.Description, &b.Metric, &b.Rating, &b.RatingCount)
		if b.BID == "" || err != nil {
			return benchmark{}, notFound404ErrMsg, err
		}
	} else {
		err = db.QueryRow(`
		UPDATE benchmarks
		SET title = $1, icon = $2, description = $3, metric = $4, rating = $5, rating_count = $6
		WHERE bid = $7
		RETURNING bid, title, icon, description, metric, rating, rating_count`, title, icon, description, metric, rating, ratingCount, idI).Scan(&b.BID, &b.Title, &b.Icon, &b.Description, &b.Metric, &b.Rating, &b.RatingCount)
	}
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
	_, err = db.Exec(`
	DELETE
	FROM benchmarks
	WHERE bid = $1`, idI)
	if err != nil {
		return msg{}, err
	}
	return msg{}, nil
}

func dbQueryGetBenchmarkComments(db *sql.DB, bid string) (benchmarkComments, msg, error) {
	var rows *sql.Rows
	idI, err := strconv.Atoi(bid)
	if err != nil {
		return benchmarkComments{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return benchmarkComments{}, msg{}, err
	}
	var bcs benchmarkComments
	b, _, err2 := dbQueryGetBenchmark(db, bid)
	if err2 == nil {
		bcs.Benchmark = b
	}
	rows, err = db.Query(`
		SELECT *
		FROM benchmark_comments
		WHERE benchmark = $1`, idI)
	if rows == nil || err != nil {
		if err2 != nil {
			return benchmarkComments{}, notFound404ErrMsg, err2
		} else {
			return benchmarkComments{}, msg{}, nil
		}
	} else {
		for rows.Next() {
			var bc benchmarkComment
			err = rows.Scan(&bc.CID, &bc.Body, &bc.Benchmark.BID, &bc.User.ID)
			if err != nil {
				return benchmarkComments{}, etcErr500ErrMsg, err
			}
			bcs.BenchmarkComments = append(bcs.BenchmarkComments, bc)
		}
		return bcs, msg{}, nil
	}
}

func dbQueryGetBenchmarkComment(db *sql.DB, bid string, cid string) (benchmarkComment, msg, error) {
	idI, err := strconv.Atoi(bid)
	idJ, err2 := strconv.Atoi(cid)
	if err != nil || err2 != nil {
		return benchmarkComment{}, badReq400ErrMsg, err
	}
	var bc benchmarkComment
	b, _, err4 := dbQueryGetBenchmark(db, bid)
	if err4 == nil {
		bc.Benchmark = b
	}
	row, err3 := db.Query(`
		SELECT *
		FROM benchmark_comments
		WHERE benchmark = $1 AND cid = $2`, idI, idJ)
	if row == nil || err3 != nil {
		return benchmarkComment{}, notFound404ErrMsg, err4
	} else {
		for row.Next() {
			err = row.Scan(&bc.CID, &bc.Body, &bc.Benchmark.BID, &bc.User.ID)
		}
		if err != nil {
			return benchmarkComment{}, etcErr500ErrMsg, err
		}
		return bc, msg{}, nil
	}
}

func dbQueryPostBenchmarkComment(db *sql.DB, bid string, body string, id string) (benchmarkComment, msg, error) {
	err := db.Ping()
	if err != nil {
		return benchmarkComment{}, dbErr500ErrMsg, err
	}
	idJ, err2 := strconv.Atoi(id)
	if err2 != nil {
		return benchmarkComment{}, badReq400ErrMsg, err
	}
	b, _, err4 := dbQueryGetBenchmark(db, bid)
	if err4 != nil || b.BID == "" {
		return benchmarkComment{}, notFound404ErrMsg, err4
	}
	idK, err5 := strconv.Atoi(b.BID)
	if err5 != nil {
		return benchmarkComment{}, etcErr500ErrMsg, err
	}
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO benchmark_comments(body, benchmark, "user")
		VALUES($1, $2, $3)
		RETURNING cid`, body, idK, idJ).Scan(&lastInsertId)
	if err != nil {
		return benchmarkComment{}, notFound404ErrMsg, nil
	}
	return benchmarkComment{CID: string(rune(lastInsertId)), Body: body, Benchmark: benchmark{BID: bid}, User: user{ID: id}}, msg{}, nil
}

func dbQueryDeleteBenchmarkComment(db *sql.DB, bid string, cid string) (msg, error) {
	err := db.Ping()
	if err != nil {
		return dbErr500ErrMsg, err
	}
	idJ, err2 := strconv.Atoi(cid)
	if err2 != nil {
		return etcErr500ErrMsg, err
	}
	b, _, err4 := dbQueryGetBenchmark(db, bid)
	if err4 != nil || b.BID == "" {
		return notFound404ErrMsg, err
	}
	_, err = db.Exec(`
	DELETE
	FROM benchmark_comments
	WHERE cid = $1`, idJ)
	if err != nil {
		return dbErr500ErrMsg, err
	}
	return msg{}, nil
}

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
	if err != nil {
		u, _, err4 = dbQueryGetUser(db, id, true)
		if err4 == nil {
			ss.User = u
		}
	}
	if err2 != nil {
		b, _, err4 = dbQueryGetBenchmark(db, bid)
		if err4 == nil {
			ss.Benchmark = b
		}
	}
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
		if showAll {
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE "user" = $1`, idI)
		} else {
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE "user" = $1 AND is_verified = TRUE`, idI)
		}
	} else {
		if err != nil || err2 != nil {
			return submissions{}, badReq400ErrMsg, err
		}
		if showAll {
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE "user" = $1 AND benchmark = $2`, idI, idJ)
		} else {
			rows, err3 = db.Query(`
			SELECT *
			FROM submissions
			WHERE "user" = $1 AND benchmark = $2 AND is_verified = TRUE`, idI, idJ)
		}
	}
	for rows.Next() {
		var s submission
		err = rows.Scan(&s.SID, &s.Result, &s.Screenshot, &s.Rating, &s.RatingCount, &s.IsVerified, &s.Benchmark.BID, &s.User.ID)
		if err != nil {
			return submissions{}, etcErr500ErrMsg, err
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
		err = row.Scan(&s.SID, &s.Result, &s.Screenshot, &s.Rating, &s.RatingCount, &s.IsVerified, &s.Benchmark.BID, &s.User.ID)
	}
	if err != nil {
		return submission{}, etcErr500ErrMsg, err
	}
	b, _, err4 := dbQueryGetBenchmark(db, s.Benchmark.BID)
	if err4 == nil {
		s.Benchmark = b
	}
	u, _, err5 := dbQueryGetUser(db, s.User.ID, true)
	if err5 == nil {
		s.User = u
	}
	return s, msg{}, nil
}

func dbQueryPostSubmission(db *sql.DB, screenshot string, benchmark benchmark, id string) (submission, msg, error) {
	err0 := db.Ping()
	if err0 != nil {
		return submission{}, dbErr500ErrMsg, err0
	}
	idI, err := strconv.Atoi(benchmark.BID)
	idJ, err3 := strconv.Atoi(id)
	if err != nil || err3 != nil {
		return submission{}, badReq400ErrMsg, err
	}
	b, _, err2 := dbQueryGetBenchmark(db, benchmark.BID)
	if err2 != nil || b.BID == "" {
		return submission{}, notFound404ErrMsg, err2
	}
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO submissions(screenshot, benchmark, "user")
		VALUES($1, $2, $3)
		RETURNING sid`, screenshot, idI, idJ).Scan(&lastInsertId)
	if err != nil {
		return submission{}, conflict409ErrMsg, err
	}
	return submission{SID: string(rune(lastInsertId)), Screenshot: screenshot, Benchmark: benchmark, User: user{ID: id}}, msg{}, nil
}

func dbQueryUpdateSubmission(db *sql.DB, id string, sid string, result float32, rating float32, ratingCount int, isVerified bool) (submission, msg, error) {
	idI, err := strconv.Atoi(sid)
	var idJ int
	idJ, err = strconv.Atoi(id)
	if err != nil {
		return submission{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return submission{}, dbErr500ErrMsg, err
	}
	var s submission
	row, err2 := db.Query(`
		SELECT *
		FROM submission_ratings
		WHERE submission = $1 AND "user" = $2`, idI, idJ)
	if row != nil && err2 == nil {
		err = db.QueryRow(`
		UPDATE submissions
		SET result = $1, is_verified = $2
		WHERE sid = $3
		RETURNING sid, result, rating, rating_count`, result, isVerified, idI).Scan(&s.SID, &s.Result, &s.Rating, &s.RatingCount)
		if err != nil {
			return submission{}, notFound404ErrMsg, err
		} else if s.SID == "" {
			return submission{}, notFound404ErrMsg, err
		}
	} else {
		err = db.QueryRow(`
		UPDATE submissions
		SET result = $1, rating = $2, rating_count = $3, is_verified = $4
		WHERE sid = $5
		RETURNING sid, result, rating, rating_count`, result, rating, ratingCount, isVerified, idI).Scan(&s.SID, &s.Result, &s.Rating, &s.RatingCount)
	}
	if err != nil || s.SID == "" {
		return submission{}, notFound404ErrMsg, err
	}
	return s, msg{}, nil
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
		FROM submission_comments
		WHERE submission = $1`, idI)
	var scs submissionComments
	for rows.Next() {
		var sc submissionComment
		err = rows.Scan(&sc.CID, &sc.Body, &sc.Submission.SID, &sc.User.ID)
		if err != nil {
			return submissionComments{}, etcErr500ErrMsg, err
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
	if err != nil || err2 != nil {
		return submissionComment{}, badReq400ErrMsg, err
	}
	row, err3 := db.Query(`
		SELECT *
		FROM submission_comments
		WHERE submission = $1 AND cid = $2`, idI, idJ)
	if row == nil || err3 != nil {
		return submissionComment{}, notFound404ErrMsg, err
	}
	var sc submissionComment
	for row.Next() {
		err = row.Scan(&sc.CID, &sc.Body, &sc.Submission.SID, &sc.User.ID)
	}
	if err != nil {
		return submissionComment{}, etcErr500ErrMsg, err
	}
	s, _, err4 := dbQueryGetSubmission(db, sid)
	if err4 == nil {
		sc.Submission = s
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
		INSERT INTO submission_comments(body, submission, "user")
		VALUES($1, $2, $3)
		RETURNING cid`, body, idK, idJ).Scan(&lastInsertId)
	if err != nil {
		return submissionComment{}, notFound404ErrMsg, nil
	}
	return submissionComment{CID: string(rune(lastInsertId)), Body: body, Submission: submission{SID: sid}, User: user{ID: id}}, msg{}, nil
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
	s, _, err4 := dbQueryGetSubmission(db, sid)
	if err4 != nil || s.SID == "" {
		return err
	}
	_, err = db.Exec(`
	DELETE
	FROM submission_comments
	WHERE cid = $1`, idJ)
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
			SELECT id, nickname, avatar, is_banned, is_verified, is_mod
			FROM users`)
	} else {
		rows, err = db.Query(`
			SELECT id, nickname, avatar, is_banned, is_verified, is_mod
			FROM users
			WHERE is_verified = TRUE`)
	}
	if rows == nil || err != nil {
		return users{}, err
	}
	var us users
	for rows.Next() {
		var u user
		err = rows.Scan(&u.ID, &u.Nickname, &u.Avatar, &u.IsBanned, &u.IsVerified, &u.IsMod)
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
		SELECT id, nickname, avatar, is_banned, is_verified, is_mod
		FROM users
		WHERE id = $1`, idI)
	} else {
		row, err = db.Query(`
		SELECT id, nickname, avatar, is_banned, is_verified, is_mod
		FROM users
		WHERE nickname = $1`, str)
	}
	if row == nil {
		return user{}, notFound404ErrMsg, nil
	}
	if err != nil {
		return user{}, dbErr500ErrMsg, err
	}
	var u user
	for row.Next() {
		err = row.Scan(&u.ID, &u.Nickname, &u.Avatar, &u.IsBanned, &u.IsVerified, &u.IsMod)
	}
	if err != nil {
		return user{}, etcErr500ErrMsg, err
	}
	return u, msg{}, nil
}

func dbQueryPostUser(db *sql.DB, dbS *sql.DB, nickname string, email string, password string) (user, msg, error) {
	err := db.Ping()
	if err != nil {
		return user{}, dbErr500ErrMsg, err
	}
	hash := []byte(password)
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO users(email, nickname, password)
		VALUES($1, $2, sha256($3))
		RETURNING id`, email, nickname, hash).Scan(&lastInsertId)
	if err != nil {
		return user{}, conflict409ErrMsg, err
	}
	_, err = dbS.Exec(`
		INSERT INTO secrets(cid)
		VALUES($1)`, lastInsertId)
	if err != nil {
		return user{ID: string(rune(lastInsertId)), Nickname: nickname, Email: email}, dbErr500ErrMsg, err
	}
	return user{ID: string(rune(lastInsertId)), Nickname: nickname, Email: email}, msg{}, nil
}

func dbQueryUpdateUser(db *sql.DB, id string, email string, avatar string, isVerified bool, isBanned bool) (user, msg, error) {
	idI, err := strconv.Atoi(id)
	if err != nil {
		return user{}, badReq400ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return user{}, dbErr500ErrMsg, err
	}
	var u user
	err = db.QueryRow(`
		UPDATE users
		SET email = $1, avatar = $2, is_verified = $3, is_banned = $4
		WHERE id = $5
		RETURNING id, nickname, avatar, is_banned, is_verified, is_mod`, email, avatar, isVerified, isBanned, idI).Scan(&u.ID, &u.Nickname, &u.Avatar, &u.IsBanned, &u.IsVerified, &u.IsMod)
	if err != nil || u.ID == "" {
		return user{}, dbErr500ErrMsg, err
	}
	return u, msg{}, nil
}

func dbQueryDeleteUser(db *sql.DB, dbS *sql.DB, id string) error {
	idI, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
	DELETE
	FROM users
	WHERE id = $1`, idI)
	if err != nil {
		return err
	}
	_, err = dbS.Exec(`
	DELETE
	FROM secrets
	WHERE cid = $1`, idI)
	if err != nil {
		return err
	}
	return nil
}
