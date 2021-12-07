package main

import (
	"database/sql"
	"fmt"
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

func dbQueryGetBenchmark(db *sql.DB, bid string) (benchmark, error) {
	idI, err := strconv.Atoi(bid)
	if err != nil {
		return benchmark{}, err
	}
	row, err2 := db.Query(`
		SELECT *
		FROM benchmarks
		WHERE bid = $1`, idI)
	if row == nil {
		return benchmark{}, nil
	}
	if err2 != nil {
		return benchmark{}, err
	}
	var b benchmark
	for row.Next() {
		err = row.Scan(&b.BID, &b.Title, &b.Icon, &b.Description, &b.Metric, &b.Rating, &b.RatingCount)
	}
	if err != nil {
		return benchmark{}, err
	}
	return b, nil
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

func dbQueryUpdateBenchmark(db *sql.DB, bid string, title string, icon string, description string, metric string, rating float32, ratingCount int) (benchmark, msg, error) {
	idI, err := strconv.Atoi(bid)
	if err != nil {
		return benchmark{}, etcErr500ErrMsg, err
	}
	err = db.Ping()
	if err != nil {
		return benchmark{}, dbErr500ErrMsg, err
	}
	var b benchmark
	err = db.QueryRow(`
		UPDATE benchmarks
		SET title = $1, icon = $2, description = $3, metric = $4, rating = $5, rating_count = $6
		WHERE bid = $5
		RETURNING bid, title, icon, description, metric, rating, rating_count`, title, icon, description, metric, rating, ratingCount, idI).Scan(&b.BID, &b.Title, &b.Icon, &b.Description, &b.Metric, &b.Rating, &b.RatingCount)
	if err != nil || b.BID == "" {
		return benchmark{}, dbErr500ErrMsg, err
	}
	return b, msg{}, nil
}

func dbQueryDeleteBenchmark(db *sql.DB, bid string) error {
	idI, err := strconv.Atoi(bid)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
	DELETE
	FROM benchmarks
	WHERE id = $1`, idI)
	if err != nil {
		return err
	}
	return nil
}

func dbQueryGetBenchmarkComments(db *sql.DB, bid string) (benchmarkComments, error) {
	var rows *sql.Rows
	idI, err := strconv.Atoi(bid)
	if err != nil {
		return benchmarkComments{}, err
	}
	err = db.Ping()
	if err != nil {
		return benchmarkComments{}, err
	}
	rows, err = db.Query(`
		SELECT *
		FROM benchmark_comments
		WHERE benchmark = $1`, idI)
	if rows == nil || err != nil {
		return benchmarkComments{}, err
	}
	var bcs benchmarkComments
	for rows.Next() {
		var bc benchmarkComment
		err = rows.Scan(&bc.CID, &bc.Body, &bc.Benchmark.BID, &bc.User.ID)
		if err != nil {
			return benchmarkComments{}, err
		}
		bcs.BenchmarkComments = append(bcs.BenchmarkComments, bc)
	}
	b, err2 := dbQueryGetBenchmark(db, bid)
	if err2 == nil {
		bcs.Benchmark = b
	}
	return bcs, nil
}

func dbQueryGetBenchmarkComment(db *sql.DB, bid string, cid string) (benchmarkComment, error) {
	idI, err := strconv.Atoi(bid)
	idJ, err2 := strconv.Atoi(cid)
	if err != nil || err2 != nil {
		return benchmarkComment{}, err
	}
	row, err3 := db.Query(`
		SELECT *
		FROM benchmark_comments
		WHERE benchmark = $1 AND cid = $2`, idI, idJ)
	if row == nil {
		return benchmarkComment{}, err
	}
	if err3 != nil {
		return benchmarkComment{}, err
	}
	var bc benchmarkComment
	for row.Next() {
		err = row.Scan(&bc.CID, &bc.Body, &bc.Benchmark.BID, &bc.User.ID)
	}
	if err != nil {
		return benchmarkComment{}, err
	}
	b, err4 := dbQueryGetBenchmark(db, bid)
	if err4 == nil {
		bc.Benchmark = b
	}
	return bc, nil
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
	b, err4 := dbQueryGetBenchmark(db, bid)
	if err4 != nil {
		return benchmarkComment{}, dbErr500ErrMsg, err2
	} else if b.BID == "" {
		return benchmarkComment{}, notFound404ErrMsg, nil
	}
	var lastInsertId int
	err = db.QueryRow(`
		INSERT INTO benchmark_comments(body, benchmark, user)
		VALUES($1, $2, $3)
		RETURNING cid`, body, b, idJ).Scan(&lastInsertId)
	if err != nil {
		return benchmarkComment{}, notFound404ErrMsg, nil
	}
	return benchmarkComment{CID: string(rune(lastInsertId)), Body: body, User: user{ID: id}}, msg{}, nil
}

func dbQueryDeleteBenchmarkComment(db *sql.DB, bid string, cid string) error {
	err := db.Ping()
	if err != nil {
		return err
	}
	idJ, err2 := strconv.Atoi(cid)
	if err2 != nil {
		return err
	}
	b, err4 := dbQueryGetBenchmark(db, bid)
	if err4 != nil || b.BID == "" {
		return err
	}
	_, err = db.Exec(`
	DELETE
	FROM benchmark_comments
	WHERE cid = $1`, idJ)
	if err != nil {
		return err
	}
	return nil
}

func dbQueryGetSubmissions(db *sql.DB, id string, bid string) (submissions, error) {
	var rows *sql.Rows
	var ss submissions
	var u user
	var b benchmark
	var err4 error
	err3 := db.Ping()
	if err3 != nil {
		return submissions{}, err3
	}
	idJ, err := strconv.Atoi(bid)
	idI, err2 := strconv.Atoi(id)
	if err != nil {
		u, err4 = dbQueryGetUser(db, id, "")
		if err4 == nil {
			ss.User = u
		}
	}
	if err2 != nil {
		b, err4 = dbQueryGetBenchmark(db, bid)
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
			return submissions{}, err
		}
		rows, err3 = db.Query(`
		SELECT *
		FROM submissions
		WHERE benchmark = $1`, idJ)
	} else if bid == "" {
		if err2 != nil {
			return submissions{}, err
		}
		rows, err3 = db.Query(`
		SELECT *
		FROM submissions
		WHERE user = $1`, idI)
	} else {
		if err != nil || err2 != nil {
			return submissions{}, err
		}
		rows, err3 = db.Query(`
		SELECT *
		FROM submissions
		WHERE user = $1 AND benchmark = $2`, idI, idJ)
	}
	if rows == nil || err3 != nil {
		return submissions{}, err
	}
	for rows.Next() {
		var s submission
		err = rows.Scan(&s.SID, &s.Result, &s.Screenshot, &s.Rating, &s.RatingCount, &s.IsVerified, &s.Benchmark.BID, &s.User.ID)
		if err != nil {
			return submissions{}, err
		}
		ss.Submissions = append(ss.Submissions, s)
	}
	return ss, nil
}

func dbQueryGetSubmission(db *sql.DB, sid string) (submission, error) {
	row, err := db.Query(`
		SELECT *
		FROM submissions
		WHERE sid = $1`, sid)
	if row == nil {
		return submission{}, nil
	}
	if err != nil {
		return submission{}, err
	}
	var s submission
	for row.Next() {
		var Benchmark string
		var User string
		err = row.Scan(&s.SID, &s.Result, &s.Screenshot, &s.Rating, &s.RatingCount, &s.IsVerified, &Benchmark, &User)
		BenchmarkB, err2 := dbQueryGetBenchmark(db, Benchmark)
		if err2 != nil {
			return submission{}, err2
		}
		s.Benchmark = BenchmarkB
		// user code
	}
	if err != nil {
		return submission{}, err
	}
	return s, nil
}

func dbQueryPostSubmission(db *sql.DB, result float32, screenshot string, benchmark benchmark) (submission, error) {
	fmt.Println(db.Ping())
	return submission{SID: "0", Result: result, Screenshot: screenshot, Benchmark: benchmark}, nil
}

func dbQueryUpdateSubmission(db *sql.DB, sid string, result float32, rating float32) (submission, error) {
	fmt.Println(db.Ping())
	s, _ := dbQueryGetSubmission(db, sid)
	s.Result = result
	s.Rating = (s.Rating*float32(s.RatingCount) + rating) / float32(s.RatingCount+1)
	s.RatingCount = s.RatingCount + 1
	return s, nil
}

func dbQueryDeleteSubmission(db *sql.DB, sid string) error {
	fmt.Println(sid)
	return nil
}

func dbQueryGetSubmissionComments(db *sql.DB, sid string) (submissionComments, error) {
	rows, err := db.Query(`
		SELECT cid, body, user
		FROM submission_comments
		WHERE submission = $1`, sid)
	if rows == nil || err != nil {
		return submissionComments{}, err
	}
	var scs submissionComments
	for rows.Next() {
		var sc submissionComment
		var User string
		err = rows.Scan(&sc.CID, &sc.Body, &User)
		if err != nil {
			return submissionComments{}, err
		}
		// add user code
		scs.SubmissionComments = append(scs.SubmissionComments, sc)
	}
	s, err2 := dbQueryGetSubmission(db, sid)
	if err2 != nil {
		return submissionComments{}, err2
	}
	scs.Submission = s
	return scs, nil
}

func dbQueryGetSubmissionComment(db *sql.DB, sid string, cid string) (submissionComment, error) {
	row, err := db.Query(`
		SELECT cid, body, user
		FROM submission_comments
		WHERE submission = $1 AND cid = $2`, sid, cid)
	if row == nil {
		return submissionComment{}, nil
	}
	if err != nil {
		return submissionComment{}, err
	}
	var sc submissionComment
	var User string
	for row.Next() {
		err = row.Scan(&sc.CID, &sc.Body, &User)
	}
	if err != nil {
		return submissionComment{}, err
	}
	// add user code
	s, err2 := dbQueryGetSubmission(db, sid)
	if err2 != nil {
		return submissionComment{}, err2
	}
	sc.Submission = s
	return sc, nil
}

func dbQueryPostSubmissionComment(db *sql.DB, sid string, body string, user user) (submissionComment, error) {
	fmt.Println(user)
	return submissionComment{CID: "0", Body: sid + " " + body, User: user}, nil
}

func dbQueryDeleteSubmissionComment(db *sql.DB, sid string, cid string) error {
	fmt.Println(cid + " " + sid)
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

func dbQueryGetUser(db *sql.DB, id string, nickname string) (user, error) {
	idI, err := strconv.Atoi(id)
	if err != nil {
		return user{}, err
	}
	row, err2 := db.Query(`
		SELECT id, nickname, avatar, is_banned, is_verified, is_mod
		FROM users
		WHERE id = $1 OR nickname = $2`, idI, nickname)
	if row == nil {
		return user{}, nil
	}
	if err2 != nil {
		return user{}, err
	}
	var u user
	for row.Next() {
		err = row.Scan(&u.ID, &u.Nickname, &u.Avatar, &u.IsBanned, &u.IsVerified, &u.IsMod)
	}
	if err != nil {
		return user{}, err
	}
	return u, nil
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
		return user{}, etcErr500ErrMsg, err
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
