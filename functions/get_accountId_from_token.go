package functions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/dgrijalva/jwt-go"
)

func GetAccountIdFromToken(tokenStr string, key []byte) (accoutId string, err error) {
	if tokenStr == "" {
		err = errors.New("token is nil")
		return
	}
	var token *jwt.Token
	//解析token（jwt）
	if token, err = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	}); err != nil {
		logs.Error("jwt parse token failed,", err.Error())
		return
	}

	//token中提取用户名
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		accoutId = fmt.Sprintf("%s", claims["sub"])
	} else {
		err = fmt.Errorf("parse user name from token failed")
		return
	}
	if accoutId == "" {
		err = fmt.Errorf("user account id is nil")
		return
	}
	accoutId = strings.Split(accoutId, "@")[0]
	return accoutId, nil
}
