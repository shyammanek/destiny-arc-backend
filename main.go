package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User represents the structure of the user data
type User struct {
	ID             string `json:"id" gorm:"primaryKey"` // Matches 'id' column
	Email          string `json:"email"`
	Name           string `json:"name"`
	Photo          string `json:"photo"`
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
	ReadingDate    string `json:"reading_date"`
}

var db *gorm.DB

// initDB initializes the database connection
func initDB() {
	var err error
	dsn := "postgresql://shyam:zAmST0MYGH3YeFwYrVGNkg@jumbo-auk-6240.j77.aws-ap-southeast-1.cockroachlabs.cloud:26257/defaultdb?sslmode=verify-full"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	db.AutoMigrate(&User{})
}

// GoogleAuthMiddleware handles Google OAuth token verification
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func GoogleAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization Header"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		accessToken = strings.TrimPrefix(accessToken, "Bearer ")

		// Fetch user info from Google UserInfo endpoint
		userInfo, err := fetchGoogleUserInfo(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired access token"})
			c.Abort()
			return
		}

		// Set user information in the context
		c.Set("userID", userInfo.ID)
		c.Set("email", userInfo.Email)
		c.Set("name", userInfo.Name)
		c.Set("photo", userInfo.Picture)

		// Continue with the request
		c.Next()
	}
}

func fetchGoogleUserInfo(accessToken string) (*GoogleUser, error) {
	// Call Google's UserInfo endpoint
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	// Parse the response body
	var user GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func main() {
	r := gin.Default()

	// Initialize the database
	initDB()

	// Apply the GoogleAuthMiddleware to simulate authentication
	r.Use(GoogleAuthMiddleware())

	// Protected route to store user data
	// Apply GoogleAuthMiddleware to protect this route
	r.POST("/saveProfile", GoogleAuthMiddleware(), func(c *gin.Context) {
		var user User

		// Bind JSON to User struct
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Println("Bind Error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Extract email from the verified token
		tokenEmail, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Ensure the email in the body matches the token email
		if user.Email != tokenEmail {
			c.JSON(http.StatusForbidden, gin.H{"error": "Email mismatch with token"})
			return
		}

		log.Println("Incoming Data:", user)

		// Upsert logic: Update existing row or insert new data
		if err := db.Model(&User{}).
			Where("email = ?", user.Email).
			Assign(User{
				Name:           user.Name,
				Photo:          user.Photo,
				DOB:            user.DOB,
				Hobbies:        user.Hobbies,
				MobileNumber:   user.MobileNumber,
				LocationCity:   user.LocationCity,
				Country:        user.Country,
				TarotReading:   user.TarotReading,
				CardsSelection: user.CardsSelection,
				PersonName:     user.PersonName,
				Coins:          user.Coins,
				Relationship:   user.Relationship,
				Occupation:     user.Occupation,
				BirthTime:      user.BirthTime,
				BirthCity:      user.BirthCity,
				BirthState:     user.BirthState,
				BirthCountry:   user.BirthCountry,
				ReadingDate:    user.ReadingDate,
			}).
			FirstOrCreate(&user).Error; err != nil {
			log.Println("DB Save/Update Error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save or update user data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User data saved/updated successfully"})
	})

	// Get user profile from the database based on email
	r.GET("/getProfile", func(c *gin.Context) {
		email := c.Query("email") // Get the email from query parameters
		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email query parameter is required"})
			return
		}

		var user User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			log.Println("DB Fetch Error:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		log.Println("Fetched User:", user)
		c.JSON(http.StatusOK, gin.H{"user": user})
	})

	// Start the server
	r.Run(":8080")
}
