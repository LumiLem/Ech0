package model

import (
	"encoding/json"
	"time"
)

// Echo 定义Echo实体
type Echo struct {
	ID            uint      `gorm:"primaryKey"                                       json:"id"`
	Content       string    `gorm:"type:text;not null"                               json:"content"`
	Username      string    `gorm:"type:varchar(100)"                                json:"username,omitempty"`
	Media         []Media   `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE" json:"media,omitempty"`
	Images        []Media   `gorm:"-"                                                json:"images,omitempty"` // 兼容旧版客户端，不存数据库
	Layout        string    `gorm:"type:varchar(50);default:'waterfall'"             json:"layout,omitempty"`
	Private       bool      `gorm:"default:false"                                    json:"private"`
	UserID        uint      `gorm:"not null;index"                                   json:"user_id"`
	Extension     string    `gorm:"type:text"                                        json:"extension,omitempty"`
	ExtensionType string    `gorm:"type:varchar(100)"                                json:"extension_type,omitempty"`
	Tags          []Tag     `gorm:"many2many:echo_tags;"                             json:"tags,omitempty"`
	FavCount      int       `gorm:"default:0"                                        json:"fav_count"`
	CreatedAt     time.Time `                                                        json:"created_at"`
	User          User      `gorm:"foreignKey:UserID"                                json:"user,omitempty"` // 关联用户信息
}

// User 用户信息（用于Echo关联查询）
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// Message 定义Message实体 (注意⚠️: 该模型为旧版Echo模型,新版已经弃用)
// type Message struct {
// 	ID            uint      `gorm:"primaryKey"           json:"id"`
// 	Content       string    `gorm:"type:text;not null"   json:"content"`
// 	Username      string    `gorm:"type:varchar(100)"    json:"username,omitempty"`
// 	ImageURL      string    `gorm:"type:text"            json:"image_url,omitempty"`
// 	ImageSource   string    `gorm:"type:varchar(20)"     json:"image_source,omitempty"`
// 	Images        []Image   `gorm:"foreignKey:MessageID" json:"images,omitempty"`
// 	Private       bool      `gorm:"default:false"        json:"private"`
// 	UserID        uint      `gorm:"not null;index"       json:"user_id"`
// 	Extension     string    `gorm:"type:text"            json:"extension,omitempty"`
// 	ExtensionType string    `gorm:"type:varchar(100)"    json:"extension_type,omitempty"`
// 	CreatedAt     time.Time `                            json:"created_at"`
// }

// Media 定义Media实体（原Image实体）
type Media struct {
	ID          uint   `gorm:"primaryKey"       json:"id"`
	MessageID   uint   `gorm:"index;not null"   json:"message_id"`           // 关联的Echo ID(注意⚠️: 该字段名为MessageID, 但实际关联的是Echo表,因为为了兼容旧版Echo用户)
	MediaURL    string `gorm:"type:text"        json:"media_url"`            // 媒体URL（原image_url）
	MediaType   string `gorm:"type:varchar(20)" json:"media_type"`           // 媒体类型: image/video
	MediaSource string `gorm:"type:varchar(20)" json:"media_source"`         // 媒体来源: local/url/s3（原image_source）
	ObjectKey   string `gorm:"type:text"        json:"object_key,omitempty"` // 对象存储的Key (如果是本地存储则为空)
	Width       int    `gorm:"default:0"        json:"width,omitempty"`      // 媒体宽度
	Height      int    `gorm:"default:0"        json:"height,omitempty"`     // 媒体高度
	LiveVideoID *uint  `gorm:"index"            json:"live_video_id,omitempty"` // 实况照片关联的视频Media ID（仅图片类型有效）
	LivePairID  string `gorm:"-"                json:"live_pair_id,omitempty"`  // 实况照片配对ID（仅用于请求，不持久化）
}

// Image 旧版兼容结构体，用于 JSON 序列化时提供 images 字段（仅包含图片，不含视频）
type Image struct {
	ID          uint   `json:"id"`
	MessageID   uint   `json:"message_id"`
	ImageURL    string `json:"image_url"`            // 图片URL
	ImageSource string `json:"image_source"`         // 图片来源: local/url/s3
	ObjectKey   string `json:"object_key,omitempty"` // 对象存储的Key (如果是本地存储则为空)
	Width       int    `json:"width,omitempty"`      // 图片宽度
	Height      int    `json:"height,omitempty"`     // 图片高度
}

// MediaToImage 将 Media 转换为兼容旧版的 Image（仅用于图片类型）
func MediaToImage(m Media) Image {
	return Image{
		ID:          m.ID,
		MessageID:   m.MessageID,
		ImageURL:    m.MediaURL,
		ImageSource: m.MediaSource,
		ObjectKey:   m.ObjectKey,
		Width:       m.Width,
		Height:      m.Height,
	}
}

// UnsupportedMediaHintImage 旧版客户端不支持的媒体类型提示占位图
// 使用 Base64 Data URI 内嵌
const UnsupportedMediaHintImage = "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MDAiIGhlaWdodD0iMzAwIiB2aWV3Qm94PSIwIDAgNDAwIDMwMCI+CiAgPGRlZnM+CiAgICA8bGluZWFyR3JhZGllbnQgaWQ9ImJnIiB4MT0iMCUiIHkxPSIwJSIgeDI9IjEwMCUiIHkyPSIxMDAlIj4KICAgICAgPHN0b3Agb2Zmc2V0PSIwJSIgc3R5bGU9InN0b3AtY29sb3I6I2Y1ZjVmNCIvPgogICAgICA8c3RvcCBvZmZzZXQ9IjEwMCUiIHN0eWxlPSJzdG9wLWNvbG9yOiNlN2U1ZTQiLz4KICAgIDwvbGluZWFyR3JhZGllbnQ+CiAgPC9kZWZzPgogIAogIDwhLS0g6IOM5pmvIC0tPgogIDxyZWN0IHdpZHRoPSI0MDAiIGhlaWdodD0iMzAwIiBmaWxsPSJ1cmwoI2JnKSIgcng9IjEyIi8+CiAgCiAgPCEtLSDop4bpopHlm77moIcgLS0+CiAgPGcgdHJhbnNmb3JtPSJ0cmFuc2xhdGUoMjAwLCA5MCkiPgogICAgPGNpcmNsZSBjeD0iMCIgY3k9IjAiIHI9IjQwIiBmaWxsPSIjYThhMjllIiBvcGFjaXR5PSIwLjMiLz4KICAgIDxwYXRoIGQ9Ik0tMTIgLTE1IEwtMTIgMTUgTDE1IDAgWiIgZmlsbD0iIzc4NzE2YyIvPgogIDwvZz4KICAKICA8IS0tIOS4u+aWh+WtlyAtLT4KICA8dGV4dCB4PSIyMDAiIHk9IjE2MCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZm9udC1mYW1pbHk9InN5c3RlbS11aSwgLWFwcGxlLXN5c3RlbSwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxOCIgZm9udC13ZWlnaHQ9IjYwMCIgZmlsbD0iIzU3NTM0ZSI+CiAgICDmraTlhoXlrrnljIXlkKvop4bpopHmiJblrp7lhrXnhafniYcKICA8L3RleHQ+CiAgCiAgPCEtLSDkuI3mlK/mjIHmj5DnpLogLS0+CiAgPHRleHQgeD0iMjAwIiB5PSIxODUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZvbnQtZmFtaWx5PSJzeXN0ZW0tdWksIC1hcHBsZS1zeXN0ZW0sIHNhbnMtc2VyaWYiIGZvbnQtc2l6ZT0iMTMiIGZpbGw9IiNhOGEyOWUiPgogICAg5b2T5YmN54mI5pys5LiN5pSv5oyB5pi+56S65q2k57G75Z6L5YaF5a65CiAgPC90ZXh0PgogIAogIDwhLS0g5Ymv5paH5a2X5bim5Zu+5qCHIC0g5pW05L2T5bGF5LitIC0tPgogIDxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKDkzLCAyMTUpIj4KICAgIDx0ZXh0IHg9IjAiIHk9IjAiIGZvbnQtZmFtaWx5PSJzeXN0ZW0tdWksIC1hcHBsZS1zeXN0ZW0sIHNhbnMtc2VyaWYiIGZvbnQtc2l6ZT0iMTQiIGZpbGw9IiM3ODcxNmMiPuivt+eCueWHu+S4i+aWueeahDwvdGV4dD4KICAgIDwhLS0gbGlua3RvIOWbvuaghyAtLT4KICAgIDxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKDk2LCAtMTMpIHNjYWxlKDAuNjUpIj4KICAgICAgPHBhdGggZmlsbD0iIzc4NzE2YyIgZD0iTTExIDZhMSAxIDAgMSAxIDAgMkg1djExaDExdi02YTEgMSAwIDEgMSAyIDB2NmEyIDIgMCAwIDEtMiAySDVhMiAyIDAgMCAxLTItMlY4YTIgMiAwIDAgMSAyLTJ6bTktM2ExIDEgMCAwIDEgMSAxdjVhMSAxIDAgMSAxLTIgMFY2LjQxNGwtOC4yOTMgOC4yOTNhMSAxIDAgMCAxLTEuNDE0LTEuNDE0TDE3LjU4NiA1SDE1YTEgMSAwIDEgMSAwLTJaIi8+CiAgICA8L2c+CiAgICA8dGV4dCB4PSIxMjAiIHk9IjAiIGZvbnQtZmFtaWx5PSJzeXN0ZW0tdWksIC1hcHBsZS1zeXN0ZW0sIHNhbnMtc2VyaWYiIGZvbnQtc2l6ZT0iMTQiIGZpbGw9IiM3ODcxNmMiPuafpeeci+WujOaVtOWGheWuuTwvdGV4dD4KICA8L2c+CiAgCiAgPCEtLSDlupXpg6jmj5DnpLogLS0+CiAgPHRleHQgeD0iMjAwIiB5PSIyNjUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZvbnQtZmFtaWx5PSJzeXN0ZW0tdWksIC1hcHBsZS1zeXN0ZW0sIHNhbnMtc2VyaWYiIGZvbnQtc2l6ZT0iMTIiIGZpbGw9IiNhOGEyOWUiPgogICAgRWNoMCDCtyDljYfnuqflrqLmiLfnq6/ku6Xojrflvpfmm7Tlpb3kvZPpqowKICA8L3RleHQ+Cjwvc3ZnPg=="

// MarshalJSON 自定义 Echo 的 JSON 序列化，自动填充 images 字段以兼容旧版客户端
// images 只包含图片类型（包括实况照片的图片部分），过滤掉视频
// 如果有视频或实况照片，会在末尾添加一张提示占位图
func (e Echo) MarshalJSON() ([]byte, error) {
	// 将 Media 中的图片转换为旧版 Image 列表（过滤掉视频）
	var images []Image
	hasUnsupportedMedia := false

	for _, m := range e.Media {
		if m.MediaType == MediaTypeImage {
			// 图片类型，包括实况照片的图片部分
			images = append(images, MediaToImage(m))
			// 检查是否是实况照片（有关联视频）
			if m.LiveVideoID != nil {
				hasUnsupportedMedia = true
			}
		} else if m.MediaType == MediaTypeVideo {
			// 视频类型，旧版不支持
			hasUnsupportedMedia = true
		}
	}

	// 如果有旧版不支持的媒体类型，添加提示占位图
	if hasUnsupportedMedia {
		images = append(images, Image{
			ImageURL:    UnsupportedMediaHintImage,
			ImageSource: MediaSourceURL, // 直链
		})
	}

	// 使用匿名结构体避免递归调用 MarshalJSON
	type EchoAlias Echo
	return json.Marshal(&struct {
		EchoAlias
		Images []Image `json:"images,omitempty"` // 覆盖原有的 Images 字段
	}{
		EchoAlias: EchoAlias(e),
		Images:    images,
	})
}

// Tag 定义Tag实体
type Tag struct {
	ID         uint      `gorm:"primaryKey"                            json:"id"`
	Name       string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`        // 标签名称
	UsageCount int       `gorm:"default:0"                             json:"usage_count"` // 使用计数
	CreatedAt  time.Time `                                             json:"created_at"`  // 创建时间
}

// EchoTag 纯关系表，联合主键
type EchoTag struct {
	EchoID uint `gorm:"primaryKey;autoIncrement:false"` // Echo ID
	TagID  uint `gorm:"primaryKey;autoIncrement:false"` // Tag ID
}

const (
	Extension_MUSIC      = "MUSIC"      // 扩展附加内容--音乐
	Extension_VIDEO      = "VIDEO"      // 扩展附加内容--视频
	Extension_GITHUBPROJ = "GITHUBPROJ" // 扩展附加内容--GitHub项目
	Extension_WEBSITE    = "WEBSITE"    // 扩展附加内容--网站

	MediaTypeImage = "image" // 媒体类型--图片
	MediaTypeVideo = "video" // 媒体类型--视频

	MediaSourceLocal = "local" // 本地媒体（原ImageSourceLocal）
	MediaSourceURL   = "url"   // 直链媒体（原ImageSourceURL）
	MediaSourceS3    = "s3"    // S3 媒体（原ImageSourceS3）

	LayoutWaterfall  = "waterfall"  // 瀑布流布局
	LayoutGrid       = "grid"       // 九宫格布局
	LayoutHorizontal = "horizontal" // 横向布局
	LayoutCarousel   = "carousel"   // 单图轮播布局

)
