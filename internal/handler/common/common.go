package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	service "github.com/lin-snow/ech0/internal/service/common"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
)

type CommonHandler struct {
	commonService service.CommonServiceInterface
}

// NewCommonHandler CommonHandler 的构造函数
func NewCommonHandler(commonService service.CommonServiceInterface) *CommonHandler {
	return &CommonHandler{
		commonService: commonService,
	}
}

// ShowImage 显示图片
// func (commonHandler *CommonHandler) ShowImage() gin.HandlerFunc {
// 	return func (ctx *gin.Context) {
// 		ctx.Header("Access-Control-Allow-Origin", "*")

// 		// 安全校验：防止路径遍历攻击
// 		filepath := ctx.Param("filepath")
// 		if filepath == "/" || filepath == ".." {
// 			ctx.AbortWithStatusJSON(http.StatusBadRequest, commonModel.INVALID_FILE_PATH)
// 		}

// 		ctx.File("./data/images/" + ctx.Param(filepath))
// 	}
// }

// UploadImage 上传图片
//
//	@Summary		上传图片
//	@Description	用户上传图片，成功后返回图片的访问 URL
//	@Tags			通用功能
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file						true	"图片文件"
//	@Success		200		{object}	res.Response{data=string}	"上传成功，返回图片URL"
//	@Failure		200		{object}	res.Response				"上传失败"
//	@Router			/images/upload [post]
func (commonHandler *CommonHandler) UploadImage() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 提取上传的 File数据
		file, err := ctx.FormFile("file")
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		// 从表单中提取source字符串
		source := ctx.PostForm("ImageSource")
		if source != string(echoModel.MediaSourceLocal) && source != string(echoModel.MediaSourceS3) {
			source = string(echoModel.MediaSourceLocal)
		}

		// 提取userid
		userId := ctx.MustGet("userid").(uint)

		// 调用 CommonService 上传文件
		imageDto, err := commonHandler.commonService.UploadImage(userId, file, source)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: imageDto,
			Msg:  commonModel.UPLOAD_SUCCESS,
		}
	})
}

// DeleteImage 删除图片
//
//	@Summary		删除图片
//	@Description	用户删除已上传的图片，需传入图片 URL 和来源信息
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Param			imageDto	body		commonModel.ImageDto	true	"图片删除请求体"
//	@Success		200			{object}	res.Response			"删除成功"
//	@Failure		200			{object}	res.Response			"删除失败"
//	@Router			/images/delete [delete]
func (commonHandler *CommonHandler) DeleteImage() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userId := ctx.MustGet("userid").(uint)

		var imageDto commonModel.ImageDto
		if err := ctx.ShouldBindJSON(&imageDto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		if err := commonHandler.commonService.DeleteImage(userId, imageDto.URL, imageDto.SOURCE, imageDto.ObjectKey); err != nil {
			ctx.JSON(
				http.StatusOK,
				commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
					Msg: "",
					Err: err,
				})),
			)
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_SUCCESS,
		}
	})
}

// GetStatus 获取Echo状态
//
//	@Summary		获取 Echo 系统状态
//	@Description	查询系统当前运行状态及初始化安装状态
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"获取状态成功"
//	@Failure		200	{object}	res.Response	"获取状态失败或未初始化"
//	@Router			/status [get]
func (commonHandler *CommonHandler) GetStatus() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		_, err := commonHandler.commonService.GetSysAdmin()
		if err != nil {
			return res.Response{
				Code: commonModel.InitInstallCode,
				Msg:  commonModel.SIGNUP_FIRST,
			}
		}

		status, err := commonHandler.commonService.GetStatus()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: status,
			Msg:  commonModel.GET_STATUS_SUCCESS,
		}
	})
}

// GetHeatMap 获取热力图数据
//
//	@Summary		获取热力图数据
//	@Description	获取系统活动热力图数据，用于展示用户活动分布情况。不传参数返回近30天数据，传year和month返回指定月份数据
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Param			year	query		int							false	"年份（可选）"
//	@Param			month	query		int							false	"月份（可选，1-12）"
//	@Success		200		{object}	res.Response{data=object}	"获取热力图数据成功"
//	@Failure		200		{object}	res.Response				"获取热力图数据失败"
//	@Router			/heatmap [get]
func (commonHandler *CommonHandler) GetHeatMap() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取可选的年月参数
		yearStr := ctx.Query("year")
		monthStr := ctx.Query("month")

		var heatMap []commonModel.Heatmap
		var err error

		if yearStr != "" && monthStr != "" {
			// 如果传了年月参数，获取指定月份的数据
			year, yearErr := strconv.Atoi(yearStr)
			month, monthErr := strconv.Atoi(monthStr)
			if yearErr != nil || monthErr != nil || month < 1 || month > 12 {
				return res.Response{
					Msg: commonModel.INVALID_QUERY_PARAMS,
					Err: nil,
				}
			}
			heatMap, err = commonHandler.commonService.GetHeatMapByMonth(year, month)
		} else {
			// 默认获取近30天数据
			heatMap, err = commonHandler.commonService.GetHeatMap()
		}

		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: heatMap,
			Msg:  commonModel.GET_HEATMAP_SUCCESS,
		}
	})
}

// GetRss 获取RSS
//
//	@Summary		获取RSS订阅源
//	@Description	获取系统的RSS订阅源（Atom格式），用于订阅最新动态
//	@Tags			通用功能
//	@Accept			json
//	@Produce		application/rss+xml
//	@Success		200	{string}	string			"返回RSS内容（xml格式）"
//	@Failure		200	{object}	res.Response	"获取RSS失败"
//	@Router			/rss [get]
func (commonHandler *CommonHandler) GetRss(ctx *gin.Context) {
	atom, err := commonHandler.commonService.GenerateRSS(ctx)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
				Msg: "",
				Err: err,
			})),
		)
		return
	}

	ctx.Data(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(atom))
}

// UploadAudio 上传音频
//
//	@Summary		上传音频
//	@Description	用户上传音频文件，成功后返回音频的访问 URL
//	@Tags			通用功能
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file						true	"音频文件"
//	@Success		200		{object}	res.Response{data=string}	"上传成功，返回音频URL"
//	@Failure		200		{object}	res.Response				"上传失败"
//	@Router			/audios/upload [post]
func (commonHandler *CommonHandler) UploadAudio() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 提取userid
		userId := ctx.MustGet("userid").(uint)

		// 提取上传的 File数据
		file, err := ctx.FormFile("file")
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		audioUrl, err := commonHandler.commonService.UploadMusic(userId, file)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: audioUrl,
			Msg:  commonModel.UPLOAD_SUCCESS,
		}
	})
}

// DeleteAudio 删除音频
//
//	@Summary		删除音频
//	@Description	用户删除已上传的音频文件
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"删除成功"
//	@Failure		200	{object}	res.Response	"删除失败"
//	@Router			/audios/delete [delete]
func (commonHandler *CommonHandler) DeleteAudio() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 提取userid
		userId := ctx.MustGet("userid").(uint)

		if err := commonHandler.commonService.DeleteMusic(userId); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_SUCCESS,
		}
	})
}

// GetPlayMusic 获取可播放的音乐
//
//	@Summary		获取可播放的音乐
//	@Description	获取当前可供播放的音乐文件URL
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=string}	"获取音乐URL成功"
//	@Failure		200	{object}	res.Response				"获取音乐URL失败"
//	@Router			/getmusic [get]
func (commonHandler *CommonHandler) GetPlayMusic() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		musicUrl := commonHandler.commonService.GetPlayMusicUrl()

		return res.Response{
			Data: musicUrl,
			Msg:  commonModel.GET_MUSIC_URL_SUCCESS,
		}
	})
}

// PlayMusic 播放音乐
//
//	@Summary		播放音乐
//	@Description	以流的方式播放当前可用的音乐文件
//	@Tags			通用功能
//	@Accept			json
//	@Produce		audio/mpeg
//	@Success		200	{string}	string			"音频流"
//	@Failure		200	{object}	res.Response	"播放失败"
//	@Router			/playmusic [get]
func (commonHandler *CommonHandler) PlayMusic(ctx *gin.Context) {
	commonHandler.commonService.PlayMusic(ctx)
}

// HelloEch0 处理HelloEch0请求
//
//	@Summary		Hello Ech0
//	@Description	获取 Ech0 系统欢迎信息、版本号和 GitHub 地址
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=object}	"获取欢迎信息成功"
//	@Router			/hello [get]
func (commonHandler *CommonHandler) HelloEch0() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		hello := struct {
			Hello   string `json:"hello"`
			Version string `json:"version"`
			Github  string `json:"github"`
		}{
			Hello:   "Hello, Ech0! 👋",
			Version: commonModel.FullVersion,
			Github:  "https://github.com/LumiLem/Ech0",
		}

		return res.Response{
			Msg:  commonModel.GET_HELLO_SUCCESS,
			Data: hello,
		}
	})
}

// Healthz 健康检查
//
//	@Summary		健康检查
//	@Description	健康检查
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response	"健康检查成功"
//	@Router			/healthz [get]
func (commonHandler *CommonHandler) Healthz() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		return res.Response{
			Msg: commonModel.GET_HEALTHZ_SUCCESS,
			Data: struct {
				Status string `json:"status"`
			}{
				Status: "ok",
			},
		}
	})
}

// GetS3PresignURL 获取 S3 预签名 URL
//
//	@Summary		获取 S3 预签名 URL
//	@Description	获取用于上传文件到 S3 的预签名 URL
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Param			s3Dto	body		commonModel.GetPresignURLDto	true	"S3 预签名请求体"
//	@Success		200		{object}	res.Response{data=object}		"获取预签名 URL 成功"
//	@Failure		200		{object}	res.Response					"获取预签名 URL 失败"
//	@Router			/s3/presign [put]
func (commonHandler *CommonHandler) GetS3PresignURL() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userId := ctx.MustGet("userid").(uint)
		// 解析请求体中的参数
		var s3Dto commonModel.GetPresignURLDto
		if err := ctx.ShouldBindJSON(&s3Dto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		presignDto, err := commonHandler.commonService.GetS3PresignURL(userId, &s3Dto, "PUT")
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: presignDto,
			Msg:  commonModel.GET_S3_PRESIGN_URL_SUCCESS,
		}
	})
}

// GetWebsiteTitle 获取网站标题
//
//	@Summary		获取网站标题
//	@Description	获取网站标题
//	@Tags			通用功能
//	@Accept			json
//	@Produce		json
//	@Param			website_url	query		string						true	"网站URL"
//	@Success		200			{object}	res.Response{data=string}	"获取网站标题成功"
//	@Failure		200			{object}	res.Response				"获取网站标题失败"
//	@Router			/website/title [get]
func (commonHandler *CommonHandler) GetWebsiteTitle() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var dto commonModel.GetWebsiteTitleDto
		if err := ctx.ShouldBindQuery(&dto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_QUERY_PARAMS,
				Err: err,
			}
		}
		title, err := commonHandler.commonService.GetWebsiteTitle(dto.WebSiteURL)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}
		return res.Response{
			Data: title,
			Msg:  commonModel.GET_WEBSITE_TITLE_SUCCESS,
		}
	})
}
