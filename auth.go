package main

import (
	"database/sql"
	"fmt"
	"github.com/golang-jwt/jwt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	secretSize = 16
	refreshMin = 15
)

func makeSecret() string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, secretSize)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func dbSQueryCheckSecretLogin(dbS *sql.DB, id string) (string, msg, error) {
	var secret string
	idI, err := strconv.Atoi(id)
	if err != nil {
		return "", notAuth401ErrMsg, err
	}
	err = dbS.QueryRow(`
		SELECT secret
		FROM secrets
		WHERE cid = $1`, idI).Scan(&secret)
	if err != nil {
		return "", dbErr500ErrMsg, err
	}
	if secret != "" {
		return "", conflict409ErrMsg, nil
	}
	return dbSQuerySendAndRefreshSecret(dbS, id)
}

func dbSQuerySendAndRefreshSecret(dbS *sql.DB, id string) (string, msg, error) {
	secret, err := dbSQueryRefreshSecret(dbS, id, false,0)
	if err != nil {
		return "", dbErr500ErrMsg, err
	}
	atClaims := jwt.MapClaims{}
	atClaims["id"] = id
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	tokenT, err2 := at.SignedString([]byte(secret))
	if err2 != nil {
		return "", etcErr500ErrMsg, err2
	}
	go func() {
		_, _ = dbSQueryRefreshSecret(dbS, id, false, refreshMin)
	}()
	return tokenT, msg{ "Successfully refreshed secret" }, nil
}

func dbSQueryRefreshSecret(dbS *sql.DB, id string, nullify bool, sleep int) (string, error) {
	time.Sleep(time.Minute * time.Duration(sleep))
	idI, err := strconv.Atoi(id)
	if err != nil {
		return "", err
	}
	var secret string
	if nullify {
		secret = ""
	} else {
		secret = makeSecret()
	}
	_, err = dbS.Exec(`
		UPDATE secrets
		SET secret = $1
		WHERE cid = $2
		RETURNING secret`, secret, idI)
	if err != nil {
		return "", err
	}
	return secret, nil
}

func dbSQueryVerifyToken(dbS *sql.DB, token string, nullify bool, id string) (string, msg, error) {
	var secret string
	idI, err := strconv.Atoi(id)
	if err != nil {
		return token, notAuth401ErrMsg, err
	}
	err2 := dbS.QueryRow(`
		SELECT secret
		FROM secrets
		WHERE cid = $1`, idI).Scan(&secret)
	if err2 != nil {
		return token, dbErr500ErrMsg, err2
	}
	if secret == "" {
		return token, notAuth401ErrMsg, nil
	}
	var tokenT = ""
	t := strings.Split(token, " ")
	if len(t) == 2 {
		tokenT = t[1]
	}
	tokenT2, err3 := jwt.Parse(tokenT, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("")
		}
		return []byte(secret), nil
	})
	if err3 != nil {
		return token, notAuth401ErrMsg, err3
	}
	if tokenT2.Valid {
		if nullify {
			_, err4 := dbSQueryRefreshSecret(dbS, id, true, 0)
			if err4 != nil {
				return token, dbErr500ErrMsg, err4
			}
			return "", noCont204Msg, nil
		}
		return dbSQuerySendAndRefreshSecret(dbS, id)
	}
	return token, notAuth401ErrMsg, nil
}

func dbQueryIsMod (db *sql.DB, id string) bool {
	i, err := strconv.Atoi(id)
	if err != nil {
		return false
	}
	var idI string
	err = db.QueryRow(`
	SELECT id
	FROM users
	WHERE id = $1 AND is_mod = TRUE`, i).Scan(&idI)
	if err != nil || idI != id {
		return false
	}
	return true
}

func dbQueryLoginUser(nickname string, password string, db *sql.DB, dbS *sql.DB) (string, msg, error) {
	var nickN string
	var disposableP string
	err := db.QueryRow(`
	SELECT nickname, password
	FROM users
	WHERE nickname = $1 AND password = sha256($2)`, nickname, password).Scan(&nickN, &disposableP)
	if err != nil {
		return "", dbErr500ErrMsg, err
	}
	if nickN != nickname {
		return "", notAuth401ErrMsg, nil
	}
	userU, err2 := dbQueryGetUser(db, "", nickN)
	if err2 != nil {
		return "", dbErr500ErrMsg, err2
	}
	id := userU.ID
	return dbSQueryCheckSecretLogin(dbS, id)
}