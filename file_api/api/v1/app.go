package v1

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"gopkg.in/gin-gonic/gin.v1"
	"tuohai/file_api/models"
	"tuohai/internal/auth"
	"tuohai/internal/console"
	"tuohai/internal/file"
	"tuohai/internal/uuid"
)

func Upload() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		f, h, err := ctx.Request.FormFile("file")
		if err != nil {
			console.StdLog.Error(err)
			ctx.JSON(http.StatusOK, gin.H{"code": 1, "data": struct{}{}, "message": "解析file文件失败"})
			return
		}
		cid := ctx.PostForm("cid")
		user := ctx.MustGet("user").(*auth.MainUser)
		creator := user.Uid
		if creator == "" || cid == "" {
			ctx.JSON(http.StatusOK, gin.H{"code": 1, "data": struct{}{}, "message": "创建者或者cid不允许为空"})
			return
		}

		suffix := filepath.Ext(h.Filename)
		buf := &bytes.Buffer{}
		buf.ReadFrom(f)
		fmt.Println(buf.Len())

		finfo := &models.FileInfo{
			Id:       uuid.NewV4().StringMd5(),
			To:       cid,
			Name:     h.Filename,
			Size:     len(buf.Bytes()),
			Type:     0,
			Ext:      suffix,
			Category: h.Header.Get("Content-Type"),
			Meta:     nil,
			Creator:  creator,
			Updated:  time.Now().Unix(),
			Created:  time.Now().Unix(),
		}

		if err := models.WriteFileToDB(finfo, file.UploadFile(suffix, buf)); err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, false)
		} else {
			renderJSON(ctx, true)
		}
		return
	}
}

func Files() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//http获取用户相关的所有toid
		cid := ctx.Query("cid")
		tos := []string{cid}
		//获得to_id 查询文件
		info, err := models.GetFilesInfo(tos)
		if err != nil {
			console.StdLog.Error(err)
			renderJSON(ctx, []int{})
			return
		}
		renderJSON(ctx, info)
	}
}

func renderJSON(ctx *gin.Context, json interface{}, err_status ...interface{}) {
	switch len(err_status) {
	case 0:
		ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": json})
		break
	case 1:
		ctx.JSON(http.StatusOK, gin.H{"code": err_status[0], "data": json})
		break
	case 2:
		ctx.JSON(http.StatusOK, gin.H{"code": err_status[0], "msg": err_status[1], "data": json})
		break
	}
}
