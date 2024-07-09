package main

import (
	"fmt"

	"log"

	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	host     = "localhost" // or the Docker service name if running in another container
	port     = 5432        // default PostgreSQL port
	user     = "postgres"  // as defined in docker-compose.yml
	password = "postgres"  // as defined in docker-compose.yml
	dbname   = "partner"   // as defined in docker-compose.yml
)

func main() {
	// Configure your PostgreSQL database details here
	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// New logger for detailed SQL logging
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Enable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect to database")
	}

	// Migrate the schema
	db.AutoMigrate(&Payment{})

	// Setup Fiber
	app := fiber.New()
	// Apply CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Adjust this to be more restrictive if needed
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	// JWT Secret Key
	secretKey := "secret"

	// Login route
	app.Post("/login", login(secretKey))

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(secretKey),
	}))

	//routes
	app.Get("/payments", func(c *fiber.Ctx) error {
		return getPayments(db, c)
	})
	app.Get("/account/:id", func(c *fiber.Ctx) error {
		return getAccount(c)
	})

	app.Post("/payments", func(c *fiber.Ctx) error {
		return createPayment(db, c)
	})

	// Start server
	log.Fatal(app.Listen(":8000"))
}

// Dummy user for example
var user1 = struct {
	Email    string
	Password string
}{
	Email:    "user@example.com",
	Password: "123",
}

func login(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type LoginRequest struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var request LoginRequest
		if err := c.BodyParser(&request); err != nil {
			return err
		}

		// Check credentials - In real world, you should check against a database
		if request.Email != user1.Email || request.Password != user1.Password {
			return fiber.ErrUnauthorized
		}

		// Create token
		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["name"] = "John Doe"
		claims["admin"] = true
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

		// Generate encoded token
		t, err := token.SignedString([]byte(secretKey))
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.JSON(fiber.Map{"token": t})
	}
}
