package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"size:64;not null"`
}

var (
	db  *gorm.DB
	rdb *redis.Client
	ctx = context.Background()
)

func main() {
	var err error
	dsn := "root:root123@tcp(127.0.0.1:3307)/study?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("MySQL: " + err.Error())
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}
	fmt.Println("MySQL connected")

	rdb = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("Redis: " + err.Error())
	}
	fmt.Println("Redis ok")

	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(Logger())

	r.GET("/health", healthHandler)
	r.GET("/api/users", listUsers)
	r.POST("/api/users", createUser)
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

func userKey(id uint64) string {
	return fmt.Sprintf("user:%d", id)
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

func getUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	key := userKey(id)
	val, err := rdb.Get(ctx, key).Result()
	if err == nil {
		var u User
		if json.Unmarshal([]byte(val), &u) == nil {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, u)
			return
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Println("redis get error: ", err)
	}

	var u User
	if err := db.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if b, err := json.Marshal(u); err == nil {
		_ = rdb.Set(ctx, key, b, 5*time.Minute).Err()
	}

	c.Header("X-Cache", "MISS")
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

	_ = rdb.Del(ctx, userKey(id)).Err()
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

	_ = rdb.Del(ctx, userKey(id)).Err()
	c.Status(http.StatusNoContent)
}
