package api

import "C"
import (
	"io/ioutil"
	"os"
	"strconv"
	"vgo/cache"
	"vgo/model"
	"vgo/service"

	"github.com/gin-contrib/sessions"

	"github.com/gin-gonic/gin"
)

// 用户注册与登录
// 注册
func HandleUserRegister(c *gin.Context) {
	register := service.UserRegisterService{}
	// BindJSON 处理 post json 的请求
	// 验证用户字段是否合法，合法则发送验证邮件
	if err := c.ShouldBindJSON(&register); err != nil {
		resp := service.ValidateTrans(err)
		c.JSON(200, resp)
	} else {
		// json 数据解析成功，需要做进一步验证
		// 先确定验证码对应的邮箱是否是提交注册内容的邮箱
		// 对邮箱和用户名进行判断
		// 插入数据库，并且对数据库插入能否成功作进一步验证
		if code := cache.Get(register.Email); code != register.Code {
			resp := service.Response{}
			resp.Code = 53003
			resp.Msg = "验证码不正确或邮箱填写错误"
			c.JSON(200, resp)
		} else {
			checkResp := model.CheckNameAndEmail(&register)
			if checkResp.Error != "" {
				// 邮箱或学工号已被注册
				c.JSON(200, checkResp)
			} else {
				user := model.User{}
				user.Create(&register)
				cache.DelCache(register.Code)
				c.JSON(200, service.Response{
					Code:  20000,
					Data:  nil,
					Msg:   "注册成功",
					Error: "",
				})
			}
		}
	}
}

// 用户登录
func HandleUserLogin(c *gin.Context) {
	loginUser := service.UserLoginService{}
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		c.JSON(200, gin.H{
			"message": "请填写必要的字段",
		})

	} else {
		user := model.User{}
		resp := user.LoginCheck(&loginUser)
		if resp.Code != 20000 {
			c.JSON(200, resp)
		} else {
			s := sessions.Default(c)
			s.Clear()
			s.Set("user_id", user.ID)
			_ = s.Save()
			c.JSON(200, resp)
		}
	}
}

// 用户退出登录状态
func HandleUserLogout(c *gin.Context) {
	s := sessions.Default(c)
	userID := s.Get("user_id")
	if uid, ok := userID.(int); ok {
		// 清空缓存
		cache.DelCache(strconv.Itoa(uid))
		s.Delete(uid)
		s.Clear()
		_ = s.Save()
		c.JSON(200, gin.H{
			"code": 20000,
			"msg":  "退出登录成功",
		})
	}
}

// 获取此用户的详细信息
func HandleGetUserInfo(c *gin.Context) {
	user := getUser(c)

	userResp := service.UserResponse{}
	userResp.Username = user.Username
	userResp.Status = user.Status
	userResp.Email = user.Email
	userResp.CreateTime = user.CreatedAt.Format("2006年1月2号 15:04:05")
	resp := &service.Response{
		Code:  20000,
		Data:  userResp,
		Msg:   "成功获取信息",
		Error: "",
	}
	c.JSON(200, resp)

}

// 用户在登录状态下修改密码
func HandleLoginStatusChangePassword(c *gin.Context) {
	type Form struct {
		OldPassword string `json:"oldpassword"`
		Password    string `json:"password"`
	}
	form := Form{}
	_ = c.BindJSON(&form)

	user := getUser(c)
	if err := user.CheckPassword(form.OldPassword); err == nil {
		resp := user.AdminUpdatePassword(form.Password)
		c.JSON(200, resp)
	} else {
		resp := &service.Response{
			Code:  510002,
			Data:  nil,
			Msg:   "旧密码输入错误，请重新输入",
			Error: "",
		}
		c.JSON(200, resp)
	}
}

// 检查单一字段
func HandleCheckFieldRepeat(c *gin.Context) {
	check := model.CheckFieldRepeat{}
	_ = c.BindJSON(&check)
	if check.Name == "email" || check.Name == "username" {
		resp := check.NameOrEmail()
		c.JSON(200, resp)
	}
}

// 添加头像
func HandleUploadAvatar(c *gin.Context) {
	basePath := os.Getenv("AVATAR_PATH")
	file, _ := c.FormFile("file")

	user := getUser(c)
	avatarName := strconv.Itoa(user.ID) + ".jpg"
	filepath := basePath + avatarName
	e := c.SaveUploadedFile(file, filepath)

	// 头像 URL 链接
	// /api/v1/user/avatar
	avatarURL := "/api/v1/user/avatar/" + avatarName

	if e != nil {
		c.String(200, e.Error())
	} else {
		user.HandleUpdateAvatar(&avatarName)
		cache.UpdateUserCache(user)
		c.String(200, avatarURL)
	}
}

// 获取头像
// 从数据库中获取当前头像的名带有文件名
func HandleGetAvatar(c *gin.Context) {
	avatarPath := os.Getenv("AVATAR_PATH")
	filename := c.Param("filename")
	// 返回类型
	contentType := "image/jpeg"
	//user, _ := c.Get("user")

	//if u, ok := user.(*model.User); ok {
	filepath := avatarPath + filename
	file, _ := os.Open(filepath)
	content, _ := ioutil.ReadAll(file)
	c.Data(200, contentType, content)
	//}

}

// 获取头像
// 获取头像，不带文件名
func HandleGetAvatarNoFileName(c *gin.Context) {
	avatarPath := os.Getenv("AVATAR_PATH")
	// 返回类型
	contentType := "image/jpeg"

	user := getUser(c)

	filepath := avatarPath + user.Avatar
	file, _ := os.Open(filepath)
	content, _ := ioutil.ReadAll(file)
	c.Data(200, contentType, content)
}
