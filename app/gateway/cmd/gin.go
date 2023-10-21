package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func InitRouter(r *gin.Engine) {
	gateway, err := InitGateway()
	if err != nil {
		log.Fatalf("failed to init handler %v", err.Error())
	}
	r.POST("/user/register", gateway.userHandler.Register)
	r.GET("/user/verify/:task_id", gateway.userHandler.RegisterVerify)
	r.POST("/user/login", gateway.userHandler.UserLogin)
	r.GET("/user/login/get_verification_code", gateway.userHandler.GetVerificationCode)
}

func SetMiddleWare(r *gin.Engine) {
	// 处理跨域问题
	// CORS
	r.Use(func(c *gin.Context) {
		method := c.Request.Method
		// 必须，接受指定域的请求，可以使用*不加以限制，但不安全
		//context.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		// 必须，设置服务器支持的所有跨域请求的方法
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		// 服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Token")
		// 可选，设置XMLHttpRequest的响应对象能拿到的额外字段
		c.Header("Access-Control-Expose-Headers", "Access-Control-Allow-Headers, Token")
		// 可选，是否允许后续请求携带认证信息Cookir，该值只能是true，不需要则不设置
		c.Header("Access-Control-Allow-Credentials", "true")
		// 放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
}
