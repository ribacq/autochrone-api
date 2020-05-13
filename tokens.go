package main

import (
	"github.com/dgrijalva/jwt-go"

	"errors"
	"os"
	"time"
)

// CanUseScope checks if a string is a valid scope for this user
func (user *User) CanUseScope(scope string) bool {
	// TODO access user authorized scopes in the DB
	switch scope {
	case "null", "basic":
		return true
	}
	return false
}

// UserAuthClaims includes custom claims for authenticating users
type UserAuthClaims struct {
	Username string `json:"username"`
	Scope    string `json:"scope"`
	jwt.StandardClaims
}

// GenerateToken generate, signs and returns a token as a string
func (user *User) GenerateToken(scope string) (string, error) {
	// check scope
	if !user.CanUseScope(scope) {
		return "", errors.New("invalid scope")
	}

	// generate token with method HS384
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, UserAuthClaims{
		Username: user.Username,
		Scope:    "basic",
		StandardClaims: jwt.StandardClaims{
			Audience:  "autochrone-front",
			ExpiresAt: now.Add(time.Minute * time.Duration(15)).Unix(),
			// Id
			IssuedAt:  now.Unix(),
			Issuer:    "autochrone-api",
			NotBefore: now.Unix(),
			// Subject
		},
	})

	// get signing key
	tokenSigningKeyFile, err := os.Open(tokenSigningKeyFilePath)
	if err != nil {
		return "", errors.New("could not access signing key")
	}
	var tokenSigningKey []byte
	_, err = tokenSigningKeyFile.Read(tokenSigningKey)
	if err != nil {
		return "", errors.New("could not access signing key")
	}

	// sign and return
	return token.SignedString(tokenSigningKey)
}

// ParseToken parses a token from a string
// returns token claims or an error
func ParseToken(tokenString string) (UserAuthClaims, error) {
	// get signing key
	tokenSigningKeyFile, err := os.Open(tokenSigningKeyFilePath)
	if err != nil {
		return UserAuthClaims{}, errors.New("could not access signing key")
	}
	var tokenSigningKey []byte
	_, err = tokenSigningKeyFile.Read(tokenSigningKey)
	if err != nil {
		return UserAuthClaims{}, errors.New("could not access signing key")
	}

	// parse token
	token, err := jwt.ParseWithClaims(tokenString, &UserAuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// check token was signed with HMAC
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || method != jwt.SigningMethodHS384 {
			return nil, errors.New("invalid signing method")
		}

		return tokenSigningKey, nil
	})
	if err != nil {
		return UserAuthClaims{}, errors.New("could not parse token")
	}

	// token has been parsed
	if token.Valid {
		// valid token: get claims
		if claims, ok := token.Claims.(*UserAuthClaims); ok {
			return *claims, nil
		} else {
			return UserAuthClaims{}, errors.New("could not get claims")
		}
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		// token validation
		if ve.Errors&(jwt.ValidationErrorMalformed|jwt.ValidationErrorUnverifiable) != 0 {
			return UserAuthClaims{}, errors.New("malformed or unverifiable token")
		} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
			return UserAuthClaims{}, errors.New("invalid token signature")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorIssuedAt|jwt.ValidationErrorNotValidYet) != 0 {
			return UserAuthClaims{}, errors.New("bad token timing")
		} else if ve.Errors&jwt.ValidationErrorClaimsInvalid != 0 {
			return UserAuthClaims{}, errors.New("invalid token claims")
		}
	}
	return UserAuthClaims{}, errors.New("unknown error")
}

// TokenValidInScope parses a token string and checks if it gives a given scope to a user
func (user *User) TokenValidInScope(tokenString, scope string) (ok bool, err error) {
	if !user.CanUseScope(scope) {
		return false, errors.New("unauthorized scope for user")
	}

	userAuthClaims, err := ParseToken(tokenString)
	if err != nil {
		return false, err
	}

	return userAuthClaims.Scope == scope, nil
}
