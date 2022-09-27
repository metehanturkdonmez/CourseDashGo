package main

import (
	"fmt"

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

var user User
var users []User

func dbConnect() *gorm.DB {

	DB, _ := gorm.Open(mysql.New(mysql.Config{
		DSN: "root:root@tcp(localhost:8889)/speedrest?charset=utf8&parseTime=True&loc=Local",
	}), &gorm.Config{})

	return DB
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

		DB := dbConnect()

		if err := DB.Create(&user).Error; err != nil {
			c.AbortWithStatus(404)
			fmt.Println(err)
		} else {
			c.JSON(200, user.ID)
		}

	})

	r.GET("/users", func(c *gin.Context) {

		DB := dbConnect()

		if err := DB.Omit("password", "activation_key").Find(&users).Error; err != nil {
			c.AbortWithStatus(404)
			fmt.Println(err)
		} else {
			c.JSON(200, users)
		}

	})
	r.Run()
}

func userRegister(user User) uint {

	DB := dbConnect()

	DB.Create(&user)
	return user.ID
}
