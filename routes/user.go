package routes

import (
	"jwt_example/database"
	"jwt_example/models"
	"log"
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

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func extractClaims(tokenStr string) (jwt.MapClaims, bool) {
    hmacSecretString := "secret"
    hmacSecret := []byte(hmacSecretString)
    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        return hmacSecret, nil
    })

    if err != nil {
		log.Println(err.Error())

        return nil, false
    }

	log.Println(token)

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println(claims)
        return claims, true
    } else {
        log.Printf("Invalid JWT Token")
        return nil, false
    }
}

func createJWTTOken(user models.User) (string, time.Time, error) {
	exp := time.Now().Add(time.Hour * 30)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["exp"] = exp.Unix()
	t, err := token.SignedString([]byte("secret"))

	if err != nil {
	  return "", time.Now(), err
	}

	return t, exp, nil
}

func Login(c *fiber.Ctx) error {
	db := database.Database.DB
	req := new(LoginRequest)
	var user models.User

	if err := c.BodyParser(&req); err != nil {
		c.Status(400).JSON(fiber.Map{"error": err.Error()})
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body")
	}

	if req.Password == "" || req.Email == "" {
		c.Status(400).JSON(fiber.Map{
			"message": "Every field is required",
		})

		return fiber.NewError(fiber.StatusBadRequest, "Invalid sign in credentials")
	}

	db.Where("email = ?", req.Email).First(&user)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.Status(400).JSON(fiber.Map{
			"message": "wrong credentials",
		})

		return fiber.NewError(fiber.StatusBadRequest, "Invalid sign in credentials")
	}
	
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

func Register(c *fiber.Ctx) error {
	db := database.Database.DB
	req := new(RegisterRequest)
	var user models.User
	var validUser models.User

	if err := c.BodyParser(&req); err != nil {
		c.Status(400).JSON(fiber.Map{"error": err.Error()})
		return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body")
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

func GetUser(c *fiber.Ctx) error {
	tokenStr := c.Cookies("jwt")
	claims, err := extractClaims(tokenStr)
	db := database.Database.DB

	if !err {
		c.Status(503).JSON(fiber.Map{
			"error": "Unauthorized",
		})

		return fiber.NewError(fiber.StatusBadRequest, "Unauthorized")
	}

	var user models.User
	db.Where("ID = ?", claims["user_id"]).First(&user)

	c.JSON(user)

	return nil
}