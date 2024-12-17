package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User represents the structure of the user data
type User struct {
	ID             string `gorm:"primaryKey"`
	Email          string `gorm:"unique"`
	Name           string
	Photo          string
	DOB            string
	Hobbies        string
	MobileNumber   string
	LocationCity   string
	Country        string
	TarotReading   string
	CardsSelection string
	PersonName     string
	Coins          int    `json:"coins"`
	Relationship   string `json:"relationship"`
	Occupation     string `json:"occupation"`
	BirthTime      string `json:"birth_time"`
	BirthCity      string `json:"birth_city"`
	BirthState     string `json:"birth_state"`
	BirthCountry   string `json:"birth_country"`
}

var db *gorm.DB

// Initialize the database connection
func initDB() {
	var err error
	dsn := "postgresql://shyam:zAmST0MYGH3YeFwYrVGNkg@jumbo-auk-6240.j77.aws-ap-southeast-1.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full"

	// Connect to CockroachDB
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Auto-migrate the User struct to create the table
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("Failed to migrate the database schema: %v", err)
	}
}

// Middleware to verify Google ID token
func GoogleAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Verify the token using Google's ID Token verifier
		ctx := context.Background()
		payload, err := idtoken.Validate(ctx, token, "YOUR_GOOGLE_CLIENT_ID")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Pass user data into context
		c.Set("userID", payload.Claims["sub"])
		c.Set("email", payload.Claims["email"])
		c.Set("name", payload.Claims["name"])
		c.Set("photo", payload.Claims["picture"])

		c.Next()
	}
}

func main() {
	r := gin.Default()

	// Initialize the database
	initDB()

	// Protected route to store user data
	r.POST("/saveProfile", GoogleAuthMiddleware(), func(c *gin.Context) {
		// Extract user information from context
		userID, _ := c.Get("userID")
		email, _ := c.Get("email")
		name, _ := c.Get("name")
		photo, _ := c.Get("photo")

		// Bind JSON body to additional fields
		var additionalData struct {
			DOB            string `json:"dob"`
			Hobbies        string `json:"hobbies"`
			MobileNumber   string `json:"mobile_number"`
			LocationCity   string `json:"location_city"`
			Country        string `json:"country"`
			TarotReading   string `json:"tarot_reading"`
			CardsSelection string `json:"cards_selection"`
			PersonName     string `json:"person_name"`
			Coins          int    `json:"coins"`
			Relationship   string `json:"relationship"`
			Occupation     string `json:"occupation"`
			BirthTime      string `json:"birth_time"`
			BirthCity      string `json:"birth_city"`
			BirthState     string `json:"birth_state"`
			BirthCountry   string `json:"birth_country"`
		}

		if err := c.ShouldBindJSON(&additionalData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create a user object
		user := User{
			ID:             userID.(string),
			Email:          email.(string),
			Name:           name.(string),
			Photo:          photo.(string),
			DOB:            additionalData.DOB,
			Hobbies:        additionalData.Hobbies,
			MobileNumber:   additionalData.MobileNumber,
			LocationCity:   additionalData.LocationCity,
			Country:        additionalData.Country,
			TarotReading:   additionalData.TarotReading,
			CardsSelection: additionalData.CardsSelection,
			PersonName:     additionalData.PersonName,
			Coins:          additionalData.Coins,
			Relationship:   additionalData.Relationship,
			Occupation:     additionalData.Occupation,
			BirthTime:      additionalData.BirthTime,
			BirthCity:      additionalData.BirthCity,
			BirthState:     additionalData.BirthState,
			BirthCountry:   additionalData.BirthCountry,
		}

		// Save user to the database
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User profile saved successfully!"})
	})

	// Health check route
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Start the server
	r.Run(":8080")
}
