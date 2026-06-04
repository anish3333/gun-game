package auth

import (
	"errors"
	"fmt"
	"os" 
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)


type Claims struct {
	PlayerID    string `json:"player_id"`
	DisplayName string `json:"display_name"`
	jwt.RegisteredClaims
}

// 2. Create a helper to safely grab the secret
func getSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Fallback for local development if something goes wrong
		return []byte("fallback-dev-secret")
	}
	return []byte(secret)
}

func GenerateGuestToken() (string, string, string, error) {
	playerID := uuid.New().String()
	displayName := "Guest-" + playerID[:6]

	claims := Claims{
		PlayerID:    playerID,
		DisplayName: displayName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// 3. Use the helper function here
	signedToken, err := token.SignedString(getSecret())
	return signedToken, playerID, displayName, err
}

func ValidateGuestToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		// 4. Use the helper function here too
		return getSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}	