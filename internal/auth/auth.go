package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"gopkg.in/gin-gonic/gin.v1"
	"tuohai/im_api/models"
	"tuohai/internal/convert"
	httplib "tuohai/internal/http"
)

func LoginAuth(host string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("session_token")
		url := GetUserInfoUrl(token, host)
		if user, err := ValidationToken(url); err != nil || user == nil {
			ctx.Abort()
			fmt.Println(err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"err_code": 1, "data": "无权限访问"})
		} else {
			//每一个新的连接会创建一个goroutine
			//每一个新的请求都会新创建Context
			//所以Context类下面的key是不需要加锁的
			//这里user可以放心使用
			fmt.Println("++++++++++")
			models.ValidAndCreate(&models.User{
				Uuid:  user.Uid,
				Uname: user.Nickname,
				Token: token,
			})
			ctx.Set("user", user)
			ctx.Set("token", token)
			ctx.Next()
		}
	}
}

//主站用户体系
type MainUser struct {
	Id       int64  `json:"id"`
	Uid      string `json:"uuid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
}

func ValidationToken(url string) (*MainUser, error) {
	var result struct {
		Msg       string    `json:"msg"`
		MainUser  *MainUser `json:"user"`
		ErrorCode float64   `json:"error_code"`
	}
	fmt.Println("auth URL: ", url)
	if url == "" {
		return nil, fmt.Errorf("%v", "url is empty")
	}
	err := httplib.Get(url).ToJson(&result)
	if err != nil {
		return nil, err
	}

	if result.ErrorCode == 0 {
		return result.MainUser, nil
	}
	return nil, fmt.Errorf("%s", result.Msg)
}

//获取个人profile URL
func GetUserInfoUrl(token, url string) string {
	return fmt.Sprintf("%s/api/i/session?%s", url, SignStr(token))
}

//获取更新用户信息URL
func GetUpdateUserInfoUrl(token, url string, params []string) string {
	return fmt.Sprintf("%s/api/i/profile?%s", url, SignStr(token, params...))
}

//获取主站好友列表URL
func GetFriendsUrl(token, url string) string {
	return fmt.Sprintf("%s/api/i/friends?%s", url, SignStr(token))
}

//生成签名参数
func SignStr(token string, params ...string) (session_token string) {
	var (
		par_str = strings.Join(params, "&")
		stamp   = convert.ToStr(time.Now().Unix())
	)
	if par_str != "" {
		par_str = "&" + par_str
	}
	params = append(
		params,
		fmt.Sprintf("session_token=%s", token),
		fmt.Sprintf("stamp=%s", stamp),
	)

	sort.Strings(params)
	sign_str := fmt.Sprintf("%scloudwork", strings.Join(params, ""))
	fmt.Println("签名字符串: ", sign_str)
	session_token = fmt.Sprintf("session_token=%s&stamp=%s&sign=%s%s", token, stamp, getSign(sign_str), par_str)
	return
}

func getSign(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return hex.EncodeToString(m.Sum(nil))
}
