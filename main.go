package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- Database Models ---

// User represents the first table
type User struct {
	gorm.Model
	Name  string `json:"name"`
	Email string `json:"email" gorm:"unique"`
	Posts []Post `json:"posts" gorm:"foreignKey:UserID"` // One-to-Many relationship
}

// Post represents the second table
type Post struct {
	gorm.Model
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  uint   `json:"user_id"`
}

// --- Database Setup ---

var DB *gorm.DB

func ConnectDatabase() {
	// Load connection string from .env
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set in .env file")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// AutoMigrate creates/updates tables automatically
	err = database.AutoMigrate(&User{}, &Post{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	DB = database
	log.Println("Database connected and migrated successfully!")
}

// --- Route Handlers ---

// CreateUser handles writing to the User table
// POST /users
func CreateUser(c *gin.Context) {
	var input User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := User{Name: input.Name, Email: input.Email}
	result := DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

// CreatePost handles writing to the Post table (linked to a User)
// POST /posts
func CreatePost(c *gin.Context) {
	var input Post
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post := Post{Title: input.Title, Content: input.Content, UserID: input.UserID}
	result := DB.Create(&post)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": post})
}

// GetUsers fetches records from the User table and preloads their Posts
// GET /users
func GetUsers(c *gin.Context) {
	var users []User
	// Preload("Posts") joins the tables to fetch related data
	result := DB.Preload("Posts").Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

// --- Main Entry Point ---

func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Connect to Database
	ConnectDatabase()

	// 3. Initialize Gin Router
	r := gin.Default()

	// 4. Define Routes
	r.POST("/users", CreateUser) // Write to Table 1
	r.POST("/posts", CreatePost) // Write to Table 2
	r.GET("/users", GetUsers)    // Fetch from Table 1 (and 2 via relation)

	// 5. Run Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}