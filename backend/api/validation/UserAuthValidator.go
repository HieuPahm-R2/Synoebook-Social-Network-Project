package validation

import (
	"Server/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var ValidatorUser = validator.New()

func ValidateUser(c *fiber.Ctx) error {
	var errors []*models.IError
	var body models.UserModel

	// lấy JSON body từ request rồi convert vào biến body.
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	//kiểm tra body theo các rule trong tag validate
	err := ValidatorUser.Struct(body)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var el models.IError
			el.Field = err.Field()
			el.Tag = err.Tag()
			errors = append(errors, &el)
		}
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}
	// ok -> request đi tiếp sang handler tiếp theo
	return c.Next()
}
