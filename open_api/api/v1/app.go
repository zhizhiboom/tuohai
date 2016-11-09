package v1

import (
	// "bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gopkg.in/gin-gonic/gin.v1"
	"tuohai/internal/console"
	"tuohai/internal/convert"
	httplib "tuohai/internal/http"
	"tuohai/internal/pb/IM_Message"
	"tuohai/internal/uuid"
	"tuohai/models"
)

func BotList(api_host string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.MustGet("token").(string)
		u := api_host + "/v1/groups?session_token=" + token
		gs, err := httplib.Groups(u)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, []int{}, 1, "远程服务器错误")
			return
		}

		var id []string
		for _, g := range gs {
			id = append(id, g.Gid)
		}
		bots, err := models.GetBots(id)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, []int{}, 1, "远程服务器错误")
			return
		}
		renderJSON(ctx, bots)
	}
}

func Apps() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apps, err := models.Apps()
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, []int{}, 1, "远程服务器错误")
			return
		}
		renderJSON(ctx, apps)
	}
}

func CreateBot(WebHookHOST, ConnLogicRPCAddress, ApiHost string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			bot          models.Bot
			msg_type     = "message"
			msg_sub_type = "bot_add"
		)
		if err := ctx.Bind(&bot); err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, err)
			return
		}

		token := ctx.MustGet("token").(string)
		if bot.CreatorId != token {
			console.StdLog.Error(errors.New("操作者必须等于创建者"))
			renderJSON(ctx, []int{}, 1, "操作者必须等于创建者")
			return
		}

		//去im_api获取用户信息
		user, err := httplib.Users(fmt.Sprintf("%s/v1/user/%s?session_token=%s", ApiHost, token, token))
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, []int{}, 1, "无效的token")
			return
		}

		bappid, _ := models.GetAppById(bot.AppId)
		if bappid.Id == "" {
			console.StdLog.Error(errors.New("无效的app_id"))
			renderJSON(ctx, []int{}, 1, "无效的app_id")
			return
		}

		u := ApiHost + "/v1/groups?session_token=" + token
		gs, err := httplib.Groups(u)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, []int{}, 1, "远程服务器错误")
			return
		}

		isgroup := false
		for _, g := range gs {
			if bot.ChannelId == g.Gid {
				isgroup = true
			}
		}

		if !isgroup {
			//用户不属于群主
			console.StdLog.Error(errors.New("用户当前不在这个群组"))
			renderJSON(ctx, []int{}, 1, "用户当前不在这个群组")
			return
		}

		b := &models.Bot{
			Id:         uuid.NewV4().String(),
			Name:       bot.Name,
			Icon:       bot.Icon,
			CreatorId:  user.Uuid,
			ChannelId:  bot.ChannelId,
			AppId:      bot.AppId,
			State:      1,
			CreateTime: time.Now(),
			UpTime:     time.Now(),
			IsPub:      bot.IsPub,
		}

		if err := models.CreateBot(b); err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "创建失败")
			return
		}

		app, err := models.GetAppById(b.AppId)
		if err != nil {
			console.StdLog.Error(err)
		}

		msg := &IM_Message.IMMsgData{
			Type:       msg_type,
			Subtype:    msg_sub_type,
			From:       b.Id,
			To:         b.ChannelId,
			MsgData:    []byte(fmt.Sprintf("%s 创建了 %s 服务", user.Uname, app.Name)),
			CreateTime: convert.ToStr(time.Now().Unix()),
		}

		if _, err := httplib.SendLogicMsg(ConnLogicRPCAddress, msg); err != nil {
			console.StdLog.Error(err)
		}

		renderJSON(ctx, gin.H{
			"web_hook":    WebHookHOST + b.Id,
			"id":          bot.Id,
			"name":        bot.Name,
			"icon":        bot.Icon,
			"creator_id":  bot.CreatorId,
			"channel_id":  bot.ChannelId,
			"app_id":      bot.AppId,
			"create_time": bot.CreateTime,
			"is_pub":      bot.IsPub,
		})
	}
}

func UpdateBot() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func DeleteBot() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func PushMsg(ConnLogicRPCAddress string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			bot_id       = ctx.Param("bot_id")
			msg_type     = "message"
			msg_sub_type = "bot_msg"
		)

		bot, err := models.GetBotById(bot_id)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "没有找到数据")
			return
		}

		if bot.Id == "" {
			renderJSON(ctx, struct{}{}, 1, "没有找到数据")
			return
		}

		if bot.ChannelId == "" {
			renderJSON(ctx, struct{}{}, 1, "没有找到数据")
			return
		}

		data, err := ioutil.ReadAll(ctx.Request.Body)
		defer ctx.Request.Body.Close()
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "读取body失败")
			return
		}

		msg := &IM_Message.IMMsgData{
			Type:       msg_type,
			Subtype:    msg_sub_type,
			From:       bot.Id,
			To:         bot.ChannelId,
			MsgData:    data,
			CreateTime: convert.ToStr(time.Now().Unix()),
		}

		if _, err := httplib.SendLogicMsg(ConnLogicRPCAddress, msg); err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "push消息失败")
			return
		}

		renderJSON(ctx, "ok")
	}
}

func PushHook() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			bot_id = ctx.Param("bot_id")
		)

		bot, err := models.GetBotById(bot_id)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "没有找到数据")
			return
		}

		app, err := models.GetAppById(bot.AppId)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "没有找到数据")
			return
		}

		log.Println(app.AppURL)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "远程服务器错误")
			return
		}

		bot_info, err := json.Marshal(bot)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "远程服务器错误")
			return
		}

		tarbody, err := ioutil.ReadAll(ctx.Request.Body)
		defer ctx.Request.Body.Close()
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "远程服务器错误")
			return
		}

		val := url.Values{
			"bot_info": []string{string(bot_info)},
			"content":  []string{string(tarbody)},
		}

		payload := strings.NewReader(val.Encode())
		req, err := http.NewRequest("POST", app.AppURL, payload)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "远程服务器错误")
			return
		}

		req.Header.Add("content-type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "远程服务器错误")
			return
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			renderJSON(ctx, "ok")
		} else {
			renderJSON(ctx, struct{}{}, 1, res.Status)
		}
	}
}

func renderJSON(ctx *gin.Context, json interface{}, err_status ...interface{}) {
	switch len(err_status) {
	case 0:
		ctx.JSON(http.StatusOK, gin.H{"err_code": 0, "data": json})
		break
	case 1:
		ctx.JSON(http.StatusOK, gin.H{"err_code": err_status[0], "data": json})
		break
	case 2:
		ctx.JSON(http.StatusOK, gin.H{"err_code": err_status[0], "err_msg": err_status[1], "data": json})
		break
	}
}

func renderSysError(ctx *gin.Context, err error) {
	if err != nil {
		console.StdLog.Error(err)
		renderJSON(ctx, struct{}{}, 1, "远程服务器错误!")
	}
}
