package la_auth

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"time"
)

var (
	secret           []byte
	ErrorMalformed   = errors.New("token被篡改")
	ErrorExpired     = errors.New("token已过期")
	ErrorNotValidYet = errors.New("token验证失败")
	ErrorUnknown     = errors.New("未知错误")
)

type Claims struct {
	UID     string `json:"uid"`
	Version int    `json:"version"`
	jwt.RegisteredClaims
}

func InitSecret(s string) {
	secret = []byte(s)
}

func NewToken(uid string, version int, day int64) (string, error) {
	claims := buildClaims(uid, version, day)
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

func buildClaims(uid string, version int, ttl int64) Claims {
	now := time.Now()
	before := now.Add(-time.Minute * 5)
	return Claims{
		UID:     uid,
		Version: version,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl*24) * time.Hour)), //Expiration time
			IssuedAt:  jwt.NewNumericDate(now),                                        //Issuing time
			NotBefore: jwt.NewNumericDate(before),                                     //Begin Effective time
		}}
}

func getSecret() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}
}

func GetClaimFromToken(tokensString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokensString, &Claims{}, getSecret())
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrorMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrorExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrorNotValidYet
			} else {
				return nil, ErrorUnknown
			}
		} else {
			return nil, ErrorNotValidYet
		}
	} else {
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			return claims, nil
		}
		return nil, ErrorNotValidYet
	}
}

func EncodePassword(pw string) string {
	return hex.EncodeToString(md5.New().Sum([]byte(pw)))
}
