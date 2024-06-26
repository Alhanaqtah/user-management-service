package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"user-management-service/internal/config"
	"user-management-service/internal/models"

	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token is expired")
)

func NewAccessToken(user *models.User, cfg config.Token) (string, error) {
	const op = "NewAccessToken"

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.UUID
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(cfg.JWT.TTL).Unix()

	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}

func NewRefreshToken(user *models.User, cfg config.Token) (string, error) {
	const op = "NewRefreshToken"

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.UUID
	claims["exp"] = time.Now().Add(cfg.Refresh.TTL).Unix()

	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}

func GetClaim(claims map[string]interface{}, claim string) (string, error) {
	const op = "GetClaim"

	c, ok := claims[claim]
	if !ok {
		return "", fmt.Errorf("%s: %w", op, errors.New("claim not found"))
	}

	var res string
	switch c.(type) {
	case float64:
		cf, ok := c.(float64)
		if !ok {
			return "", fmt.Errorf("%s: %w", op, errors.New("failed to extract string from claim"))
		}

		res = strconv.FormatFloat(cf, 'g', -1, 64)
	case string:
		res, ok = c.(string)
		if !ok {
			return "", fmt.Errorf("%s: %w", op, errors.New("failed to extract string from claim"))
		}
	}

	return res, nil
}

func ExtractClaimsFromHeader(r *http.Request, secret string) (jwt.MapClaims, error) {
	const op = "ExtractClaimsFromHeader"

	tokenString := jwtauth.TokenFromHeader(r)
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.As(err, &jwtauth.ErrExpired) {
			return nil, fmt.Errorf("%s: %w", op, ErrTokenExpired)
		}
		return nil, fmt.Errorf("%s: %w", op, errors.New("failed to parse token"))
	}

	claims := token.Claims.(jwt.MapClaims)

	return claims, nil
}

/* func CheckClaim(ctx context.Context, claim, expectedClaim string) (bool, error) {
	const op = "CheckClaim"

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	c, ok := claims[claim]
	if !ok {
		return false, fmt.Errorf("%s: claim not found", op)
	}

	switch c.(type) {
	case float64:
		claim, ok := c.(float64)
		if !ok {
			return false, fmt.Errorf("%s: %w", op, errors.New("type not found"))
		}

		expClaim, err := strconv.ParseFloat(expectedClaim, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		if claim != expClaim {
			return false, nil
		}
	case string:
		claim, ok := c.(string)
		if !ok {
			return false, fmt.Errorf("%s: %w", op, errors.New("type not found"))
		}

		if claim != expectedClaim {
			return false, nil
		}
	}

	return true, nil
}
*/
