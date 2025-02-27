package logic

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/models"
	"github.com/gravitl/netmaker/servercfg"
)

var jwtSecretKey []byte

// SetJWTSecret - sets the jwt secret on server startup
func SetJWTSecret() {
	currentSecret, jwtErr := FetchJWTSecret()
	if jwtErr != nil {
		newValue := RandomString(64)
		jwtSecretKey = []byte(newValue) // 512 bit random password
		if err := StoreJWTSecret(string(jwtSecretKey)); err != nil {
			logger.FatalLog("something went wrong when configuring JWT authentication")
		}
	} else {
		jwtSecretKey = []byte(currentSecret)
	}
}

// CreateJWT func will used to create the JWT while signing in and signing out
func CreateJWT(uuid string, macAddress string, network string) (response string, err error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &models.Claims{
		ID:         uuid,
		Network:    network,
		MacAddress: macAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Netmaker",
			Subject:   fmt.Sprintf("node|%s", uuid),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err == nil {
		return tokenString, nil
	}
	return "", err
}

// CreateUserJWT - creates a user jwt token
func CreateUserJWT(username string, issuperadmin, isadmin bool) (response string, err error) {
	expirationTime := time.Now().Add(60 * 12 * time.Minute)
	claims := &models.UserClaims{
		UserName:     username,
		IsSuperAdmin: issuperadmin,
		IsAdmin:      isadmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Netmaker",
			Subject:   fmt.Sprintf("user|%s", username),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err == nil {
		return tokenString, nil
	}
	return "", err
}

// VerifyJWT verifies Auth Header
func VerifyJWT(bearerToken string) (username string, issuperadmin, isadmin bool, err error) {
	token := ""
	tokenSplit := strings.Split(bearerToken, " ")
	if len(tokenSplit) > 1 {
		token = tokenSplit[1]
	} else {
		return "", false, false, errors.New("invalid auth header")
	}
	return VerifyUserToken(token)
}

// VerifyUserToken func will used to Verify the JWT Token while using APIS
func VerifyUserToken(tokenString string) (username string, issuperadmin, isadmin bool, err error) {
	claims := &models.UserClaims{}

	if tokenString == servercfg.GetMasterKey() && servercfg.GetMasterKey() != "" {
		return MasterUser, true, true, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if token != nil && token.Valid {
		var user *models.User
		// check that user exists
		user, err = GetUser(claims.UserName)
		if err != nil {
			return "", false, false, err
		}

		if user.UserName != "" {
			return claims.UserName, claims.IsSuperAdmin, claims.IsAdmin, nil
		}
		err = errors.New("user does not exist")
	}
	return "", false, false, err
}

// VerifyHostToken - [hosts] Only
func VerifyHostToken(tokenString string) (hostID string, mac string, network string, err error) {
	claims := &models.Claims{}

	// this may be a stupid way of serving up a master key
	// TODO: look into a different method. Encryption?
	if tokenString == servercfg.GetMasterKey() && servercfg.GetMasterKey() != "" {
		return "mastermac", "", "", nil
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if token != nil {
		return claims.ID, claims.MacAddress, claims.Network, nil
	}
	return "", "", "", err
}
