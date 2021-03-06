package service

import (
	"fmt"
	"reflect"

	"gopkg.in/go-playground/validator.v8"
)

// user 注册 json 验证
/*
validate 坑之一
千万不能在认证格式里用空格分开，如下面这种
binding:"required, min=10, max=30"
正确格式是下面
binding:"required,min=10,max=30"
*/
type UserRegisterService struct {
	UserName        string `json:"username" binding:"required,min=5,max=30"`
	Password        string `json:"password" binding:"required,min=8,max=30"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8,max=30,eqfield=Password"`
	Email           string `json:"email" binding:"required,email"`
	Code            string `json:"code" binding:"required"`
}

// 处理 user login 登录处理
type UserLoginService struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 构建用户错误的序列化器，用于将错误进行翻译
func ValidateTrans(err error) *Response {
	resp := &Response{}
	resp.Code = 51001
	resp.Msg = "post error"
	// 添加错误验证信息
	if val, ok := err.(validator.ValidationErrors); ok {
		for _, e := range val {
			// 翻译字段为中文
			actualTag := dictionary["ActualTag"][e.ActualTag]
			field := dictionary["Field"][e.Field]
			reflect.ValueOf(resp)
			resp.Error = fmt.Sprintf("%s: %s", field, actualTag)
			break
		}
	}
	return resp
}
