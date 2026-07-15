package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
}

var (
	users  = make(map[int]User)
	nextID = 1
	mu     sync.RWMutex
)

func main() {
	r := gin.Default()

	r.GET("/health", healthHandler)
	r.POST("/api/users", createUser)
	r.GET("/api/users/:id", getUser)

	r.Run(":8080")
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func createUser(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}

	mu.Lock()
	u.ID = nextID
	users[u.ID] = u
	nextID++
	mu.Unlock()

	c.JSON(http.StatusCreated, u)
}

func getUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	mu.RLock()
	u, ok := users[id]
	mu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, u)
}
