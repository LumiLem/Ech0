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
const UnsupportedMediaHintImage = "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0MDAiIGhlaWdodD0iMzQwIiB2aWV3Qm94PSIwIDAgNDAwIDM0MCI+CiAgPGRlZnM+CiAgICA8bGluZWFyR3JhZGllbnQgaWQ9ImJnIiB4MT0iMCUiIHkxPSIwJSIgeDI9IjEwMCUiIHkyPSIxMDAlIj4KICAgICAgPHN0b3Agb2Zmc2V0PSIwJSIgc3R5bGU9InN0b3AtY29sb3I6I2Y1ZjVmNCIvPgogICAgICA8c3RvcCBvZmZzZXQ9IjEwMCUiIHN0eWxlPSJzdG9wLWNvbG9yOiNlN2U1ZTQiLz4KICAgIDwvbGluZWFyR3JhZGllbnQ+CiAgPC9kZWZzPgogIAogIDwhLS0g6IOM5pmvIC0tPgogIDxyZWN0IHdpZHRoPSI0MDAiIGhlaWdodD0iMzQwIiBmaWxsPSJ1cmwoI2JnKSIgcng9IjEyIi8+CiAgCiAgPCEtLSDop4bpopHlm77moIcgLS0+CiAgPGcgdHJhbnNmb3JtPSJ0cmFuc2xhdGUoMjAwLCA2NSkiPgogICAgPGNpcmNsZSBjeD0iMCIgY3k9IjAiIHI9IjMyIiBmaWxsPSIjYThhMjllIiBvcGFjaXR5PSIwLjMiLz4KICAgIDxwYXRoIGQ9Ik0tOSAtMTEgTC05IDExIEwxMSAwIFoiIGZpbGw9IiM3ODcxNmMiLz4KICA8L2c+CiAgCiAgPCEtLSDkuLvmloflrZcgLS0+CiAgPHRleHQgeD0iMjAwIiB5PSIxMjAiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZvbnQtZmFtaWx5PSJzeXN0ZW0tdWksIC1hcHBsZS1zeXN0ZW0sIHNhbnMtc2VyaWYiIGZvbnQtc2l6ZT0iMTYiIGZvbnQtd2VpZ2h0PSI2MDAiIGZpbGw9IiM1NzUzNGUiPgogICAg5q2k5YaF5a655YyF5ZCr6KeG6aKR5oiW5a6e5Ya154Wn54mHCiAgPC90ZXh0PgogIAogIDwhLS0g5LiN5pSv5oyB5o+Q56S6IC0tPgogIDx0ZXh0IHg9IjIwMCIgeT0iMTQyIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmb250LWZhbWlseT0ic3lzdGVtLXVpLCAtYXBwbGUtc3lzdGVtLCBzYW5zLXNlcmlmIiBmb250LXNpemU9IjEyIiBmaWxsPSIjYThhMjllIj4KICAgIOW9k+WJjeeJiOacrOS4jeaUr+aMgeaYvuekuuatpOexu+Wei+WGheWuuQogIDwvdGV4dD4KICAKICA8IS0tIOWJr+aWh+Wtl+W4puWbvuaghyAtIOaVtOS9k+WxheS4rSAtLT4KICA8ZyB0cmFuc2Zvcm09InRyYW5zbGF0ZSgyMDAsIDE3MCkiPgogICAgPGcgdHJhbnNmb3JtPSJ0cmFuc2xhdGUoLTEzOCwgMCkiPgogICAgICA8dGV4dCB4PSIwIiB5PSIwIiBmb250LWZhbWlseT0ic3lzdGVtLXVpLCAtYXBwbGUtc3lzdGVtLCBzYW5zLXNlcmlmIiBmb250LXNpemU9IjEzIiBmaWxsPSIjNzg3MTZjIj7or7fngrnlh7vkuIvmlrk8L3RleHQ+CiAgICAgIDwhLS0gbGlua3RvIOWbvuaghyAtLT4KICAgICAgPGcgdHJhbnNmb3JtPSJ0cmFuc2xhdGUoNzAsIC0xMikgc2NhbGUoMC42KSI+CiAgICAgICAgPHBhdGggZmlsbD0iIzc4NzE2YyIgZD0iTTExIDZhMSAxIDAgMSAxIDAgMkg1djExaDExdi02YTEgMSAwIDEgMSAyIDB2NmEyIDIgMCAwIDEtMiAySDVhMiAyIDAgMCAxLTItMlY4YTIgMiAwIDAgMSAyLTJ6bTktM2ExIDEgMCAwIDEgMSAxdjVhMSAxIDAgMSAxLTIgMFY2LjQxNGwtOC4yOTMgOC4yOTNhMSAxIDAgMCAxLTEuNDE0LTEuNDE0TDE3LjU4NiA1SDE1YTEgMSAwIDEgMSAwLTJaIi8+CiAgICAgIDwvZz4KICAgICAgPHRleHQgeD0iOTAiIHk9IjAiIGZvbnQtZmFtaWx5PSJzeXN0ZW0tdWksIC1hcHBsZS1zeXN0ZW0sIHNhbnMtc2VyaWYiIGZvbnQtc2l6ZT0iMTMiIGZpbGw9IiM3ODcxNmMiPuaIluS9v+eUqCBjdXN0b20g54mI5p+l55yL5a6M5pW05YaF5a65PC90ZXh0PgogICAgPC9nPgogIDwvZz4KICAKICA8IS0tIEVjaDAgY3VzdG9tIOeJiOS/oeaBr+ahhiAtLT4KICA8ZyB0cmFuc2Zvcm09InRyYW5zbGF0ZSgyMDAsIDI2NSkiPgogICAgPCEtLSDovrnmoYbvvJrnlKggcGF0aCDnlLvluKbnvLrlj6PnmoTlnIbop5Lnn6nlvaLvvIzmoIfpopjljLrln5/nlZnnqbogLS0+CiAgICA8cGF0aCBkPSJNLTUwLC00NSBMLTE0MiwtNDUgUS0xNTAsLTQ1IC0xNTAsLTM3IEwtMTUwLDM3IFEtMTUwLDQ1IC0xNDIsNDUgTDE0Miw0NSBRMTUwLDQ1IDE1MCwzNyBMMTUwLC0zNyBRMTUwLC00NSAxNDIsLTQ1IEw1MCwtNDUiIAogICAgICAgICAgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjZDZkM2QxIiBzdHJva2Utd2lkdGg9IjEuNSIvPgogICAgCiAgICA8IS0tIOagh+mimCAtIOWcqOi+ueahhue8uuWPo+WkhO+8jOWujOWFqOmAj+aYjuiDjOaZryAtLT4KICAgIDx0ZXh0IHg9IjAiIHk9Ii00MCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZm9udC1mYW1pbHk9InN5c3RlbS11aSwgLWFwcGxlLXN5c3RlbSwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxMiIgZm9udC13ZWlnaHQ9IjUwMCIgZmlsbD0iIzc4NzE2YyI+CiAgICAgIEVjaDAgY3VzdG9tIOeJiAogICAgPC90ZXh0PgogICAgCiAgICA8IS0tIEdpdEh1YiDooYwgLS0+CiAgICA8ZyB0cmFuc2Zvcm09InRyYW5zbGF0ZSgtODUsIC0xMCkiPgogICAgICA8IS0tIEdpdEh1YiDlm77moIflrrnlmajvvJrlm7rlrpogMTZ4MTbvvIzlsYXkuK3lr7npvZDmloflrZcgLS0+CiAgICAgIDxzdmcgeD0iMCIgeT0iLTEyIiB3aWR0aD0iMTYiIGhlaWdodD0iMTYiIHZpZXdCb3g9IjAgMCAyNCAyNCI+CiAgICAgICAgPHBhdGggZmlsbD0iIzU3NTM0ZSIgZD0iTTEyIDBDNS4zNyAwIDAgNS4zNyAwIDEyYzAgNS4zMSAzLjQzNSA5Ljc5NSA4LjIwNSAxMS4zODUuNi4xMDUuODI1LS4yNTUuODI1LS41NyAwLS4yODUtLjAxNS0xLjIzLS4wMTUtMi4yMzUtMy4wMTUuNTU1LTMuNzk1LS43MzUtNC4wMzUtMS40MS0uMTM1LS4zNDUtLjcyLTEuNDEtMS4yMy0xLjY5NS0uNDItLjIyNS0xLjAyLS43OC0uMDE1LS43OTUuOTQ1LS4wMTUgMS42Mi44NyAxLjg0NSAxLjIzIDEuMDggMS44MTUgMi44MDUgMS4zMDUgMy40OTUuOTkuMTA1LS43OC40Mi0xLjMwNS43NjUtMS42MDUtMi42Ny0uMy01LjQ2LTEuMzM1LTUuNDYtNS45MjUgMC0xLjMwNS40NjUtMi4zODUgMS4yMy0zLjIyNS0uMTItLjMtLjU0LTEuNTMuMTItMy4xOCAwIDAgMS4wMDUtLjMxNSAzLjMgMS4yMy45Ni0uMjcgMS45OC0uNDA1IDMtLjQwNXMyLjA0LjEzNSAzIC40MDVjMi4yOTUtMS41NiAzLjMtMS4yMyAzLjMtMS4yMy42NiAxLjY1LjI0IDIuODguMTIgMy4xOC43NjUuODQgMS4yMyAxLjkwNSAxLjIzIDMuMjI1IDAgNC42MDUtMi44MDUgNS42MjUtNS40NzUgNS45MjUuNDM1LjM3NS44MSAxLjA5NS44MSAyLjIyIDAgMS42MDUtLjAxNSAyLjg5NS0uMDE1IDMuMyAwIC4zMTUuMjI1LjY5LjgyNS41N0ExMi4wMiAxMi4wMiAwIDAgMCAyNCAxMmMwLTYuNjMtNS4zNy0xMi0xMi0xMnoiLz4KICAgICAgPC9zdmc+CiAgICAgIDx0ZXh0IHg9IjI2IiB5PSIwIiBmb250LWZhbWlseT0ic3lzdGVtLXVpLCAtYXBwbGUtc3lzdGVtLCBzYW5zLXNlcmlmIiBmb250LXNpemU9IjExIiBmaWxsPSIjNTc1MzRlIj4KICAgICAgICBnaXRodWIuY29tL0x1bWlMZW0vRWNoMAogICAgICA8L3RleHQ+CiAgICA8L2c+CiAgICAKICAgIDwhLS0gRG9ja2VyIOihjCAtLT4KICAgIDxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKC04NSwgMjApIj4KICAgICAgPCEtLSBEb2NrZXIg5Zu+5qCH5a655Zmo77ya5Zu65a6aIDE2eDE277yM5bGF5Lit5a+56b2Q5paH5a2XIC0tPgogICAgICA8c3ZnIHg9IjAiIHk9Ii0xMiIgd2lkdGg9IjE2IiBoZWlnaHQ9IjE2IiB2aWV3Qm94PSIwIDAgMjQgMjQiPgogICAgICAgIDxwYXRoIGZpbGw9IiMwZGI3ZWQiIGQ9Ik0xMy45ODMgMTEuMDc4aDIuMTE5YS4xODYuMTg2IDAgMCAwIC4xODYtLjE4NVY5LjAwNmEuMTg2LjE4NiAwIDAgMC0uMTg2LS4xODZoLTIuMTE5YS4xODYuMTg2IDAgMCAwLS4xODUuMTg2djEuODg3YzAgLjEwMi4wODMuMTg1LjE4NS4xODVtLTIuOTU0LTUuNDNoMi4xMThhLjE4Ni4xODYgMCAwIDAgLjE4Ni0uMTg2VjMuNTc1YS4xODYuMTg2IDAgMCAwLS4xODYtLjE4NmgtMi4xMThhLjE4Ni4xODYgMCAwIDAtLjE4NS4xODZ2MS44ODdjMCAuMTAyLjA4My4xODYuMTg1LjE4Nm0wIDIuNzE2aDIuMTE4YS4xODYuMTg2IDAgMCAwIC4xODYtLjE4NVY2LjI5MWEuMTg2LjE4NiAwIDAgMC0uMTg2LS4xODZoLTIuMTE4YS4xODYuMTg2IDAgMCAwLS4xODUuMTg2djEuODg4YzAgLjEwMi4wODMuMTg1LjE4NS4xODVtLTIuOTMgMGgyLjEyYS4xODYuMTg2IDAgMCAwIC4xODQtLjE4NVY2LjI5MWEuMTg2LjE4NiAwIDAgMC0uMTg0LS4xODZoLTIuMTJhLjE4Ni4xODYgMCAwIDAtLjE4NC4xODZ2MS44ODhjMCAuMTAyLjA4My4xODUuMTg0LjE4NW0tMi45NjQgMGgyLjExOWEuMTg2LjE4NiAwIDAgMCAuMTg1LS4xODVWNi4yOTFhLjE4Ni4xODYgMCAwIDAtLjE4NS0uMTg2SDUuMTM2YS4xODYuMTg2IDAgMCAwLS4xODYuMTg2djEuODg4YzAgLjEwMi4wODQuMTg1LjE4Ni4xODVtNS44OTMgMi43MTVoMi4xMThhLjE4Ni4xODYgMCAwIDAgLjE4Ni0uMTg1VjkuMDA2YS4xODYuMTg2IDAgMCAwLS4xODYtLjE4NmgtMi4xMThhLjE4Ni4xODYgMCAwIDAtLjE4NS4xODZ2MS44ODdjMCAuMTAyLjA4My4xODUuMTg1LjE4NW0tMi45MyAwaDIuMTJhLjE4Ni4xODYgMCAwIDAgLjE4NC0uMTg1VjkuMDA2YS4xODYuMTg2IDAgMCAwLS4xODQtLjE4NmgtMi4xMmEuMTg2LjE4NiAwIDAgMC0uMTg0LjE4NnYxLjg4N2MwIC4xMDIuMDgzLjE4NS4xODQuMTg1bS0yLjk2NCAwaDIuMTE5YS4xODYuMTg2IDAgMCAwIC4xODUtLjE4NVY5LjAwNmEuMTg2LjE4NiAwIDAgMC0uMTg1LS4xODZINS4xMzZhLjE4Ni4xODYgMCAwIDAtLjE4Ni4xODZ2MS44ODdjMCAuMTAyLjA4NC4xODUuMTg2LjE4NW0tMi45MiAwaDIuMTJhLjE4Ni4xODYgMCAwIDAgLjE4NC0uMTg1VjkuMDA2YS4xODYuMTg2IDAgMCAwLS4xODQtLjE4NmgtMi4xMmEuMTg2LjE4NiAwIDAgMC0uMTg1LjE4NnYxLjg4N2MwIC4xMDIuMDgzLjE4NS4xODUuMTg1TTIzLjc2MyA5Ljg5Yy0uMDY1LS4wNTEtLjY3Mi0uNTEtMS45NTQtLjUxLS4zMzguMDAxLS42NzYuMDMtMS4wMS4wODctLjI0OC0xLjctMS42NTMtMi41My0xLjcxNi0yLjU2NmwtLjM0NC0uMTk5LS4yMjYuMzI3Yy0uMjg0LjQzOC0uNDkuOTIyLS42MTIgMS40My0uMjMuOTctLjA5IDEuODgyLjQwMyAyLjY2MS0uNTk1LjMzMi0xLjU1LjQxMy0xLjc0NC40MkguNzUxYS43NTEuNzUxIDAgMCAwLS43NS43NDggMTEuMzc2IDExLjM3NiAwIDAgMCAuNjkyIDQuMDYyYy41NDUgMS40MjggMS4zNTUgMi40OCAyLjQxIDMuMTI0IDEuMTguNzIzIDMuMSAxLjEzNyA1LjI3NSAxLjEzNy45ODMuMDAzIDEuOTYzLS4wODYgMi45My0uMjY2YTEyLjI0OCAxMi4yNDggMCAwIDAgMy44MjMtMS4zODljLjk4LS41NjcgMS44Ni0xLjI4OCAyLjYxLTIuMTM2IDEuMjUyLTEuNDE4IDEuOTk4LTIuOTk3IDIuNTUzLTQuNGguMjIxYzEuMzcyIDAgMi4yMTUtLjU0OSAyLjY4LTEuMDA5LjMwOS0uMjkzLjU1LS42NS43MDctMS4wNDZsLjA5OC0uMjg4WiIvPgogICAgICA8L3N2Zz4KICAgICAgPHRleHQgeD0iMjYiIHk9IjAiIGZvbnQtZmFtaWx5PSJzeXN0ZW0tdWksIC1hcHBsZS1zeXN0ZW0sIHNhbnMtc2VyaWYiIGZvbnQtc2l6ZT0iMTEiIGZpbGw9IiMwZGI3ZWQiPgogICAgICAgIGRvY2tlciBwdWxsIGx1bWxpbWUvZWNoMDpsYXRlc3QKICAgICAgPC90ZXh0PgogICAgPC9nPgogIDwvZz4KPC9zdmc+Cg=="

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
