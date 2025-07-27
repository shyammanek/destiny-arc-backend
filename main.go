package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
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

// NumerologyPrediction represents the structure for daily numerology predictions
type NumerologyPrediction struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserEmail   string    `json:"user_email" gorm:"index"`
	Date        time.Time `json:"date" gorm:"index"`
	LifePath    int       `json:"life_path"`
	Destiny     int       `json:"destiny"`
	Soul        int       `json:"soul"`
	Personality int       `json:"personality"`
	DailyNumber int       `json:"daily_number"`
	Prediction  string    `json:"prediction"`
	LuckyColor  string    `json:"lucky_color"`
	LuckyNumber int       `json:"lucky_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Affirmation   string `json:"affirmation"`
	LuckyActivity string `json:"lucky_activity"`
	Quote         string `json:"quote"`
	FocusArea     string `json:"focus_area"`
}

var db *gorm.DB

// initDB initializes the database connection
func initDB() {
	var err error
	// Use SQLite for local development
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	db.AutoMigrate(&User{}, &NumerologyPrediction{})
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

// calculateLifePath calculates the life path number from date of birth
func calculateLifePath(dob string) int {
	// Remove any non-digit characters
	dob = strings.ReplaceAll(dob, "-", "")
	dob = strings.ReplaceAll(dob, "/", "")
	dob = strings.ReplaceAll(dob, " ", "")
	
	sum := 0
	for _, char := range dob {
		if char >= '0' && char <= '9' {
			sum += int(char - '0')
		}
	}
	
	// Reduce to single digit (except master numbers 11, 22, 33)
	for sum > 9 && sum != 11 && sum != 22 && sum != 33 {
		sum = (sum / 10) + (sum % 10)
	}
	
	return sum
}

// calculateDailyNumber calculates the daily number based on current date and life path
func calculateDailyNumber(lifePath int, date time.Time) int {
	day := date.Day()
	month := int(date.Month())
	year := date.Year()
	
	sum := day + month + year + lifePath
	
	// Reduce to single digit
	for sum > 9 {
		sum = (sum / 10) + (sum % 10)
	}
	
	return sum
}

// generatePrediction generates a prediction based on the daily number
func generatePrediction(dailyNumber int) (string, string, int, string, string, string, string) {
	predictions := map[int]map[string]interface{}{
		1: {
			"prediction":     "Today is a day for new beginnings and taking initiative. Trust your instincts and step confidently into the spotlight.",
			"color":          "Red",
			"number":         1,
			"affirmation":    "I am bold and courageous. I lead with confidence.",
			"lucky_activity": "Start something new or take the lead in a group.",
			"quote":          "The future belongs to those who believe in the beauty of their dreams. – Eleanor Roosevelt",
			"focus_area":     "Initiative",
		},
		2: {
			"prediction":     "Cooperation and harmony are your strengths today. Focus on building relationships and supporting others.",
			"color":          "Orange",
			"number":         2,
			"affirmation":    "I am a source of peace and understanding.",
			"lucky_activity": "Mediate a conflict or help someone in need.",
			"quote":          "Peace begins with a smile. – Mother Teresa",
			"focus_area":     "Partnership",
		},
		3: {
			"prediction":     "Creativity and self-expression are highlighted. Share your ideas and let your imagination soar.",
			"color":          "Yellow",
			"number":         3,
			"affirmation":    "I express myself freely and joyfully.",
			"lucky_activity": "Write, paint, or engage in a creative hobby.",
			"quote":          "Creativity is intelligence having fun. – Albert Einstein",
			"focus_area":     "Expression",
		},
		4: {
			"prediction":     "Discipline and organization will bring you success. Focus on building strong foundations and completing tasks.",
			"color":          "Green",
			"number":         4,
			"affirmation":    "I am grounded, organized, and reliable.",
			"lucky_activity": "Organize your workspace or plan your week.",
			"quote":          "Success is the sum of small efforts, repeated day in and day out. – Robert Collier",
			"focus_area":     "Stability",
		},
		5: {
			"prediction":     "Embrace change and seek out new experiences. Flexibility will bring unexpected opportunities.",
			"color":          "Blue",
			"number":         5,
			"affirmation":    "I welcome change and adventure into my life.",
			"lucky_activity": "Try something you’ve never done before.",
			"quote":          "Life is either a daring adventure or nothing at all. – Helen Keller",
			"focus_area":     "Freedom",
		},
		6: {
			"prediction":     "Family and responsibility are in focus. Offer support and care to those around you.",
			"color":          "Pink",
			"number":         6,
			"affirmation":    "I nurture myself and others with love.",
			"lucky_activity": "Spend quality time with loved ones or help a friend.",
			"quote":          "The best way to find yourself is to lose yourself in the service of others. – Mahatma Gandhi",
			"focus_area":     "Care",
		},
		7: {
			"prediction":     "Reflection and introspection will bring insight. Take time for spiritual or intellectual pursuits.",
			"color":          "Purple",
			"number":         7,
			"affirmation":    "I trust my intuition and seek deeper understanding.",
			"lucky_activity": "Meditate, read, or spend time in nature.",
			"quote":          "Knowing yourself is the beginning of all wisdom. – Aristotle",
			"focus_area":     "Insight",
		},
		8: {
			"prediction":     "Ambition and determination are rewarded. Focus on your goals and take practical steps toward success.",
			"color":          "Gold",
			"number":         8,
			"affirmation":    "I am strong, capable, and successful.",
			"lucky_activity": "Set financial or career goals and take action.",
			"quote":          "The only limit to our realization of tomorrow will be our doubts of today. – Franklin D. Roosevelt",
			"focus_area":     "Achievement",
		},
		9: {
			"prediction":     "Compassion and generosity are your guides. Give back and complete unfinished business.",
			"color":          "Silver",
			"number":         9,
			"affirmation":    "I am compassionate and make a positive difference.",
			"lucky_activity": "Volunteer or help someone in need.",
			"quote":          "The best way to find yourself is to lose yourself in the service of others. – Mahatma Gandhi",
			"focus_area":     "Compassion",
		},
	}

	if data, exists := predictions[dailyNumber]; exists {
		return data["prediction"].(string), data["color"].(string), data["number"].(int),
			data["affirmation"].(string), data["lucky_activity"].(string), data["quote"].(string), data["focus_area"].(string)
	}

	return "Today brings balanced energy. Trust the universe and stay positive.", "Gray", dailyNumber,
		"I am balanced and open to the flow of life.", "Reflect and recharge.", "Balance is not something you find, it’s something you create.", "Balance"
}

func main() {
	r := gin.Default()

	// Initialize the database
	initDB()

	// Apply the GoogleAuthMiddleware to simulate authentication
	// r.Use(GoogleAuthMiddleware()) // This line applies the middleware to ALL routes:

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

	// Get daily numerology prediction
	r.GET("/numerology/daily", func(c *gin.Context) {
		userEmail := "test@example.com" // Hardcoded for testing

		// Accept dob as a query parameter
		dob := c.Query("dob")
		if dob == "" {
			// If dob not provided, try to get from user profile (legacy behavior)
			var user User
			if err := db.Where("email = ?", userEmail).First(&user).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide dob as a query parameter, or create a user profile first."})
				return
			}
			dob = user.DOB
			if dob == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Date of birth is required for numerology calculations"})
				return
			}
		}

		// Get date parameter (default to today if not provided)
		dateStr := c.Query("date")
		var targetDate time.Time
		var err error
		if dateStr == "" {
			targetDate = time.Now()
		} else {
			targetDate, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
				return
			}
		}

		// Calculate numerology numbers
		lifePath := calculateLifePath(dob)
		dailyNumber := calculateDailyNumber(lifePath, targetDate)
		prediction, luckyColor, luckyNumber, affirmation, luckyActivity, quote, focusArea := generatePrediction(dailyNumber)

		// Create new prediction (not saved to DB, just returned)
		result := NumerologyPrediction{
			UserEmail:   userEmail,
			Date:        targetDate,
			LifePath:    lifePath,
			Destiny:     lifePath, // Simplified
			Soul:        lifePath, // Simplified
			Personality: lifePath, // Simplified
			DailyNumber: dailyNumber,
			Prediction:  prediction,
			LuckyColor:  luckyColor,
			LuckyNumber: luckyNumber,
			Affirmation: affirmation,
			LuckyActivity: luckyActivity,
			Quote: quote,
			FocusArea: focusArea,
		}

		c.JSON(http.StatusOK, gin.H{
			"prediction": result,
			"message":   "Daily prediction generated successfully",
		})
	})

	// Get numerology prediction history
	r.GET("/numerology/history", func(c *gin.Context) {
		userEmail := "test@example.com" // Hardcoded for testing

		// Get limit parameter (default to 30 days)
		limitStr := c.Query("limit")
		limit := 30
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		var predictions []NumerologyPrediction
		if err := db.Where("user_email = ?", userEmail).
			Order("date DESC").
			Limit(limit).
			Find(&predictions).Error; err != nil {
			log.Println("DB Fetch Error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prediction history"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"predictions": predictions,
			"count":       len(predictions),
			"message":     "Prediction history retrieved successfully",
		})
	})

	// Get numerology prediction for a specific date
	r.GET("/numerology/date/:date", func(c *gin.Context) {
		userEmail := "test@example.com" // Hardcoded for testing

		// Get date from URL parameter
		dateStr := c.Param("date")
		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}

		var prediction NumerologyPrediction
		if err := db.Where("user_email = ? AND date = ?", userEmail, targetDate.Format("2006-01-02")).First(&prediction).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Prediction not found for this date"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"prediction": prediction,
			"message":    "Prediction retrieved successfully",
		})
	})

	// Start the server
	r.Run(":8080")
}
