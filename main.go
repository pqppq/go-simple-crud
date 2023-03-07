package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Repo struct {
	db *sql.DB
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := Repo{db: db}

	// create table if not exists
	_, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)",
	)
	if err != nil {
		log.Fatal(err)
	}

	// router
	r := gin.Default()
	r.GET("/users", repo.getUsers)
	r.GET("/users/:id", repo.getUser)
	r.POST("/users", repo.createUser)
	r.PUT("/users/:id", repo.updateUser)
	r.DELETE("/users/:id", repo.deleteUser)

	r.Run(":8000")
}

func (rp *Repo) getUsers(c *gin.Context) {
	rows, err := rp.db.Query("SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (rp *Repo) getUser(c *gin.Context) {
	id := c.Param("id")
	var u User
	err := rp.db.QueryRow("SELECT * FROM users where id = $1", id).Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": u})
}

func (rp *Repo) createUser(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")

	u := User{Name: name, Email: email}

	err := rp.db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", name, email).Scan(&u.ID)
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, gin.H{"user": u})
}

func (rp *Repo) updateUser(c *gin.Context) {
	i := c.Param("id")
	id, err := strconv.Atoi(i)
	u := User{ID: id}
	err = rp.db.QueryRow("SELECT * FROM users where id = $1", id).Scan(&u.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	u.Name = c.PostForm("name")
	u.Email = c.PostForm("email")

	_, err = rp.db.Exec("INSERT INTO users (name, email) VALUES ($1, $2) WHERE id = $3", u.Name, u.Email, u.ID)
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, gin.H{"user": u})
}

func (rp *Repo) deleteUser(c *gin.Context) {
	i := c.Param("id")
	id, err := strconv.Atoi(i)
	u := User{ID: id}
	err = rp.db.QueryRow("SELECT * FROM users where id = $1", id).Scan(&u.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	_, err = rp.db.Exec("DELETE users WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
