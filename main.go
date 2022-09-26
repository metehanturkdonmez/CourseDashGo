package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username      string
	Password      string
	DisplayName   string
	Email         string
	UserUrl       string
	ActivationKey string
	Status        int
}

func main() {
	r := gin.Default()
	r.POST("/user_register", func(c *gin.Context) {

		user := User{
			Username:      c.PostForm("username"),
			Password:      c.PostForm("password"),
			DisplayName:   c.PostForm("display_name"),
			Email:         c.PostForm("email"),
			UserUrl:       c.PostForm("user_url"),
			ActivationKey: c.PostForm("activation_key"),
			Status:        0,
		}

		registerResponse := userRegister(user)

		c.JSON(http.StatusOK, gin.H{
			"created_user_id": registerResponse,
		})
	})
	r.Run()
}

func userRegister(user User) uint {

	DB, _ := gorm.Open(mysql.New(mysql.Config{
		DSN: "root:root@tcp(localhost:8889)/speedrest?charset=utf8&parseTime=True&loc=Local",
	}), &gorm.Config{})

	DB.Create(&user)
	return user.ID
}
