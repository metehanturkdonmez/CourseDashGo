package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	jwt "github.com/appleboy/gin-jwt/v2"
)

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

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

	var identityKey = "id"

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {

				DB := dbConnect()
				result := DB.Where("username = ? AND password = ?", v.Username, v.Password).First(&user)

				if result.RowsAffected > 0 {

					return jwt.MapClaims{
						"id":       user.ID,
						"username": user.Username,
						"email":    user.Email,
					}

				}

			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &User{
				Username: claims["username"].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			Username := loginVals.Username
			Password := loginVals.Password
			DB := dbConnect()
			result := DB.Where("username = ? AND password = ?", Username, Password).First(&user)

			if result.RowsAffected > 0 {
				return &User{
					Username: user.Username,
					Password: user.Password,
					Status:   user.Status,
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*User); ok {

				DB := dbConnect()

				zator := DB.Where("username = ?", v.Username).First(&user)

				if zator.RowsAffected > 0 && user.Status > 0 {
					return true
				}
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	errInit := authMiddleware.MiddlewareInit()

	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

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

	r.POST("/login", authMiddleware.LoginHandler)

	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	auth := r.Group("/auth")
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	auth.Use(authMiddleware.MiddlewareFunc())
	{

	}

	admin := r.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc())
	{

		admin.GET("/users", func(c *gin.Context) {

			DB := dbConnect()

			if err := DB.Omit("password", "activation_key").Find(&users).Error; err != nil {
				c.AbortWithStatus(404)

				fmt.Println(err)
			} else {
				c.JSON(200, users)
			}

		})

	}

	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal(err)
	}

	r.Run()
}
