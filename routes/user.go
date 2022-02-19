package routes

import (
	"jwt_example/database"
	"jwt_example/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

func createJWTTOken(user models.User) (string, time.Time, error) {
	exp := time.Now().Add(time.Hour * 30)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["exp"] = exp
	t, err := token.SignedString([]byte("secret"))

	if err != nil {
	  return "", time.Now(), err
	}

	return t, exp, nil
}

func Login(c *fiber.Ctx) error {

	return nil
}

func Register(c *fiber.Ctx) error {
	db := database.Database.DB
	req := new(RegisterRequest)
	var user models.User
	var validUser models.User

	if err := c.BodyParser(&req); err != nil {
		c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Password == "" || req.Email == "" || req.Name == "" {
		c.Status(400).JSON(fiber.Map{
			"message": "Every field is required",
		})

		return fiber.NewError(fiber.StatusBadRequest, "Invalid sign up credentials")
	}

	db.Where("email = ?", req.Email).First(&validUser)

	if (validUser.ID != 0) {
		c.Status(400).JSON(fiber.Map{
			"message": "Email already taken",
		})

		return fiber.NewError(fiber.StatusBadRequest, "Email already taken")
	}

	db.Where("name = ?", req.Name).First(&validUser)

	if (validUser.ID != 0) {
		c.Status(400).JSON(fiber.Map{
			"message": "Name already taken",
		})

		return fiber.NewError(fiber.StatusBadRequest, "Name already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)

	if err != nil {
		c.Status(500).JSON(fiber.Map{
			"message": "Couldn't generate hash",
		})

		return fiber.NewError(fiber.StatusBadGateway, "Couldn't generate hash")
	}


	user.Email = req.Email
	user.Name = req.Name
	user.Password = string(hash)
	db.Create(&user)

	token, exp, err := createJWTTOken(user)
	if err != nil {
		c.Status(500).JSON(fiber.Map{
			"message": "Couldn't create token",
		})

		return fiber.NewError(fiber.StatusBadGateway, "Couldn't create token")
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    string(token),
		Expires:  exp,
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	c.Status(200).JSON(user)
	return nil
}

func GetUsers(c *fiber.Ctx) error {
	db := database.Database.DB
	var users []models.User

	db.Find(&users)
	c.Status(200).JSON(users)

	return nil
}

func Public(c *fiber.Ctx) error {
	return nil
}

func Private(c *fiber.Ctx) error {
	return nil
}