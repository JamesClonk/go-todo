package main

import "fmt"
import "math"
import "time"
import "strconv"
import "net/http"
import "crypto/sha512"
import "crypto/rand"
import "encoding/base64"

func Authenticate(r *http.Request) (*int, error) {
	query := r.URL.Query()

	requestAccountId, err := strconv.Atoi(query.Get("rId"))
	if err != nil {
		return nil, err
	}
	requestTimestamp, err := strconv.Atoi(query.Get("rTimestamp"))
	if err != nil {
		return nil, err
	}
	requestSalt := query.Get("rSalt")
	requestToken := query.Get("rToken")

	serverTimestamp := time.Now().Unix()
	timeDiff := math.Abs(float64(serverTimestamp - int64(requestTimestamp)))
	if timeDiff > 5 {
		return nil, nil
	}

	account, err := GetAccountById(requestAccountId)
	if err != nil {
		return nil, err
	}

	serverToken := GenerateToken(strconv.Itoa(requestTimestamp), requestSalt, account.Password)
	if requestToken != serverToken {
		return nil, nil
	}

	return &account.Id, nil
}

func GenerateToken(timestamp string, salt string, passwordHash string) string {
	return fmt.Sprintf("%x", sha512.Sum512([]byte(timestamp+salt+passwordHash)))
}

func HashPassword(salt string, password string) string {
	return fmt.Sprintf("%x", sha512.Sum512([]byte(salt+password)))
}

const RANDOM_LENGTH = 32

func GenerateRandomString() (*string, error) {

	bytes := make([]byte, RANDOM_LENGTH)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}

	enc := base64.URLEncoding
	dest := make([]byte, enc.EncodedLen(len(bytes)))
	enc.Encode(dest, bytes)

	s := string(dest)
	return &s, nil
}
