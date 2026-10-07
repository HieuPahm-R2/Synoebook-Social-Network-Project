package middleware

import (
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// c *fiber.Ctx là context của request hiện tại. Nó chứa body, params, query, response, status code, v.v
func AuthMiddleware(c *fiber.Ctx) error {
	tok := c.Get("Authorization")

	if tok == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	// Check if the token starts with "Bearer"
	if !strings.HasPrefix(tok, "Bearer ") {
		tok = strings.TrimSpace(tok)
	} else {
		splited := strings.Split(tok, "Bearer ")
		if len(splited) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "unauthenticated",
			})
		}
		tok = splited[1]
	}

	SecretKey := os.Getenv("JWT_SECRET")
	// interface{} đơn giản là kiểu “bất kỳ”. Trong trường hợp này, nó trả về secret key.
	token, err := jwt.ParseWithClaims(tok, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		//sẽ chuyển chuỗi "my_secret" thành dạng bytes:
		return []byte(SecretKey), nil
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)

	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}
	// lưu userId vào context của request hiện tại.
	c.Locals("userId", claims.Issuer)
	return c.Next()

}
