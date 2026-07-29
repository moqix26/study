package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"size:64;not null"`
}

type BatchCreateRequest struct {
	Names []string `json:"names"`
}

var db *gorm.DB

func main() {
	var err error
	dsn := "root:root123@tcp(127.0.0.1:3307)/study?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Unable to connect MySQL: " + err.Error())
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}
	fmt.Println("database ok")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(Logger())

	r.GET("/health", healthHandler)
	r.GET("/api/users", listUsers)
	r.POST("/api/users", createUser)
	r.POST("/api/users/batch", createUsersBatch)
	r.GET("/api/users/:id", getUser)
	r.PUT("/api/users/:id", updateUser)
	r.DELETE("/api/users/:id", deleteUser)

	fmt.Println(":8080 is on")
	r.Run(":8080")
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		fmt.Println("[IN ]", method, path)
		c.Next()
		fmt.Println("[OUT]", method, path, "->", c.Writer.Status())
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func listUsers(c *gin.Context) {
	var list []User
	if err := db.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func createUser(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	u.ID = 0

	if err := db.Create(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, u)
}

func createUsersBatch(c *gin.Context) {
	var req BatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	if len(req.Names) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "names required"})
		return
	}

	var created []User
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, name := range req.Names {
			name = strings.TrimSpace(name)
			if name == "" {
				return errors.New("empty name")
			}
			u := User{Name: name}
			if err := tx.Create(&u).Error; err != nil {
				return err
			}
			created = append(created, u)
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func getUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var u User
	if err := db.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func updateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var body User
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}

	var u User
	if err := db.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	u.Name = body.Name
	if err := db.Save(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func deleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	res := db.Delete(&User{}, id)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
