package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/schooldevops/gin_tutorial/models"
	"github.com/spf13/viper"
)

var ACCESS_SECRET string
var REFRESH_SECRET string

type Token struct {
	AccessToken    string
	RefreshToken   string
	AccessExpires  int64
	RefreshExpires int64
}

func AuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")

	ACCESS_SECRET = viper.GetString("token.access_secret")
	REFRESH_SECRET = viper.GetString("token.refresh_secret")

	auth.POST("/login", login)
	auth.POST("/logout", logout)
	auth.POST("/refresh", refresh)
}

func login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Invalid json format")
		return
	}

	if !isValidUser(user) {
		c.JSON(http.StatusUnauthorized, "Not valid user")
		return
	}

	token, err := CreateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err)
	}

	tokens := createJsonToken(token)

	c.JSON(http.StatusOK, tokens)
}

func logout(c *gin.Context) {

	c.JSON(http.StatusOK, map[string]string{})

}

func refresh(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	refreshToken := mapToken["refresh_token"]
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(REFRESH_SECRET), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, "Refresh token expired")
		return
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		c.JSON(http.StatusUnauthorized, err)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		token, err := CreateToken(claims["user_id"].(string))
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, err)
		}

		tokens := createJsonToken(token)
		c.JSON(http.StatusOK, tokens)

	} else {
		c.JSON(http.StatusUnauthorized, err)
	}

}

func createJsonToken(token Token) map[string]string {
	tokens := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}

	return tokens
}

func CreateToken(userid string) (token Token, err error) {

	accessExpTime := time.Now().Add(10 * time.Minute).Unix()
	refreshExpTime := time.Now().Add(24 * 7 * time.Hour).Unix()

	accessClaims := jwt.MapClaims{}
	accessClaims["authorized"] = true
	accessClaims["user_id"] = userid
	accessClaims["exp"] = accessExpTime
	acctoken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	token.AccessToken, err = acctoken.SignedString([]byte(ACCESS_SECRET))

	if err != nil {
		return token, err
	}

	refreshClaims := jwt.MapClaims{}
	refreshClaims["user_id"] = userid
	refreshClaims["exp"] = refreshExpTime
	reftoken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	token.RefreshToken, err = reftoken.SignedString([]byte(REFRESH_SECRET))

	if err != nil {
		return token, err
	}

	token.AccessExpires = accessExpTime
	token.RefreshExpires = refreshExpTime

	return token, nil
}

func isValidUser(user models.User) bool {
	if user.Username != "admin" || user.Password != "passwd" {
		return false
	}
	return true
}

func ExtractTokenFromHeader(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	tokenArr := strings.Split(bearer, " ")
	if len(tokenArr) == 2 {
		return tokenArr[1]
	}
	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractTokenFromHeader(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(ACCESS_SECRET), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractToken(r *http.Request) (map[string]interface{}, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
