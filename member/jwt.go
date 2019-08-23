package member

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
)

type Jwt struct {
	secret string
}

func NewJwt(secret string) *Jwt {
	return &Jwt{
		secret: secret,
	}
}

func (j *Jwt) Parse(token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})
	switch e := err.(type) {
	case *jwt.ValidationError:
		if e.Errors == jwt.ValidationErrorExpired {
			return nil, errors.ErrReLogin
		}
	case nil:
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			err = errors.ErrReLogin
		}
		if !t.Valid {
			err = errors.ErrLogin
		}
		if err != nil {
			return nil, err
		}
		return claims, nil
	}
	return nil, err
}
