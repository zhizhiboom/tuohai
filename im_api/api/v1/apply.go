package v1

import (
	"fmt"
	"strconv"

	"gopkg.in/gin-gonic/gin.v1"
	"tuohai/im_api/models"
	"tuohai/internal/auth"
	"tuohai/internal/console"
	"tuohai/internal/convert"
)

func ApplyFriends(url string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		main_user := ctx.MustGet("user").(*auth.MainUser)
		token := ctx.MustGet("token").(string)
		pageindex, _ := strconv.Atoi(ctx.Param("pageindex"))
		pagesize, _ := strconv.Atoi(ctx.Param("pagesize"))

		apply, err := models.FriendApplys(main_user.Uid, true, pageindex, pagesize)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "远程服务器错误")
			return
		}

		list := RenderApplyFriends(apply, token, url)
		if len(list) == 0 {
			renderJSON(ctx, []int{})
			return
		}

		renderJSON(ctx, list)
	}
}

func RenderApplyFriends(apply []models.FriendApply, token, url string) []gin.H {
	var list []gin.H
	for i, _ := range apply {
		users, err := auth.GetBatchUsers(token, url, []string{fmt.Sprintf("user_ids=%s", apply[i].ApplyUid)})
		if err != nil {
			console.StdLog.Error(err)
		}
		avatar, name := "", ""
		if len(users) != 0 {
			name = users[0].Uname
			avatar = users[0].Avatar
		}

		list = append(list, gin.H{
			"id":     apply[i].Fid,
			"uuid":   apply[i].ApplyUid,
			"way":    apply[i].Way,
			"attach": apply[i].Attach,
			"status": apply[i].Status,
			"avatar": avatar,
			"name":   name,
			"time":   apply[i].LaunchTime,
		})
	}
	return list
}

func UnApplyFriends(url string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		main_user := ctx.MustGet("user").(*auth.MainUser)
		token := ctx.MustGet("token").(string)
		pageindex, _ := strconv.Atoi(ctx.Param("pageindex"))
		pagesize, _ := strconv.Atoi(ctx.Param("pagesize"))

		apply, err := models.FriendApplys(main_user.Uid, false, pageindex, pagesize)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "远程服务器错误")
			return
		}

		list := RenderApplyFriends(apply, token, url)
		if len(list) == 0 {
			renderJSON(ctx, []int{})
			return
		}

		renderJSON(ctx, list)
	}
}

func AgreeApplyFriend() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// main_user := ctx.MustGet("user").(*auth.MainUser)
		id := ctx.PostForm("id")
		status := ctx.PostForm("status")
		if status == "" || id == "" {
			renderJSON(ctx, struct{}{}, 1, "id 或 status不能为空")
			return
		}
		status_int := convert.StrTo(status).MustInt()
		err := models.SaveFriendApply(&models.FriendApply{Fid: id, Status: models.ApplyType(status_int)})
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, struct{}{}, 1, "远程服务器错误")
			return
		}

		renderJSON(ctx, true)
	}
}
