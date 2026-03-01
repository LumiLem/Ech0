package util

import (
	"bytes"
	"fmt"
	"image"
	stdDraw "image/draw"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/chai2010/webp"
	"golang.org/x/image/draw"
)

// ImageProcessOptions 图片处理选项（与阿里云 OSS / 腾讯云 COS 参数体系语义兼容）
type ImageProcessOptions struct {
	Width   int    // 目标宽度 (0=不限制)
	Height  int    // 目标高度 (0=不限制)
	Quality int    // 质量 (1-100, 默认85, 仅对 JPEG/WebP 有效)
	Format  string // 输出格式 ("webp"/"png"/"jpg"/"jpeg", 空=保持原格式)
	Mode    string // 缩放模式: "lfit"(等比缩小,默认) / "mfit"(等比放大) / "fill"(裁剪填充)
}

// ProcessImage 图片处理核心函数：缩放 + 格式转换 + 质量压缩
// 返回处理后的字节流、Content-Type 和错误
func ProcessImage(src io.Reader, srcFormat string, opts ImageProcessOptions) ([]byte, string, error) {
	// 0. 读取所有数据到内存 (文件并不大，大的是解压后的像素组，因此这里读入内存是安全的)
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, "", fmt.Errorf("read image source failed: %w", err)
	}

	// 1. 尝试只读取头部信息获取图片尺寸，进行防爆检测
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decode image config failed: %w", err)
	}

	// 如果前端未提供图片原格式，我们优先尝试根据文件头嗅探
	if srcFormat == "" && format != "" {
		srcFormat = format
	}

	// 防 OOM 熔断保护：限制最高处理像素为 3600 万（例如 6000x6000）
	// 大于该值的很可能是炸弹图，如果此时进入 image.Decode() 会瞬间吃掉 140MB 以上的峰值内存。
	const maxPixels = 36000000
	if int64(config.Width)*int64(config.Height) > maxPixels {
		return nil, "", fmt.Errorf("image pixels too large: %dx%d exceeds absolute limit %d pixels", config.Width, config.Height, maxPixels)
	}

	// 2. 真正进行像素级解码 (安全的)
	srcImg, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decode image failed: %w", err)
	}

	bounds := srcImg.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	// 1. 计算目标尺寸
	dstW, dstH := calcDimensions(srcW, srcH, opts.Width, opts.Height, opts.Mode)

	// 2. 缩放或裁剪
	var resultImg image.Image
	if dstW == srcW && dstH == srcH {
		// 无需缩放
		resultImg = srcImg
	} else if opts.Mode == "fill" {
		// 裁剪填充模式：先缩放到覆盖目标区域，再居中裁剪
		resultImg = resizeFill(srcImg, srcW, srcH, dstW, dstH)
	} else {
		// lfit / mfit：等比缩放
		resultImg = resizeImage(srcImg, dstW, dstH)
	}

	// 3. 确定输出格式
	outFormat := opts.Format
	if outFormat == "" {
		outFormat = normalizeFormat(srcFormat)
	}

	// 4. 质量参数
	quality := opts.Quality
	if quality <= 0 || quality > 100 {
		quality = 75
	}

	// 5. 编码输出
	return encodeImage(resultImg, outFormat, quality)
}

// calcDimensions 根据源尺寸、目标尺寸和缩放模式计算最终尺寸
func calcDimensions(srcW, srcH, targetW, targetH int, mode string) (int, int) {
	if targetW <= 0 && targetH <= 0 {
		return srcW, srcH
	}

	switch mode {
	case "fill":
		// fill: 必须同时指定宽高，否则退化为 lfit
		if targetW > 0 && targetH > 0 {
			return targetW, targetH
		}
		return calcLfit(srcW, srcH, targetW, targetH)

	case "mfit":
		// mfit: 等比缩放，保证长边匹配目标尺寸（可能放大）
		return calcMfit(srcW, srcH, targetW, targetH)

	default:
		// lfit (默认): 等比缩小到不超过目标尺寸，但不放大
		return calcLfit(srcW, srcH, targetW, targetH)
	}
}

// calcLfit 等比缩小（不放大）到不超过目标尺寸
func calcLfit(srcW, srcH, targetW, targetH int) (int, int) {
	ratioW := 1.0
	ratioH := 1.0

	if targetW > 0 && srcW > targetW {
		ratioW = float64(targetW) / float64(srcW)
	}
	if targetH > 0 && srcH > targetH {
		ratioH = float64(targetH) / float64(srcH)
	}

	ratio := ratioW
	if ratioH < ratio {
		ratio = ratioH
	}

	if ratio >= 1.0 {
		return srcW, srcH // 不放大
	}

	dstW := int(float64(srcW) * ratio)
	dstH := int(float64(srcH) * ratio)
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}
	return dstW, dstH
}

// calcMfit 等比缩放，使得短边匹配目标尺寸（但不超过原图尺寸）
func calcMfit(srcW, srcH, targetW, targetH int) (int, int) {
	if targetW <= 0 {
		targetW = srcW
	}
	if targetH <= 0 {
		targetH = srcH
	}

	ratioW := float64(targetW) / float64(srcW)
	ratioH := float64(targetH) / float64(srcH)

	// 取较大的缩放比，确保覆盖目标区域
	ratio := ratioW
	if ratioH > ratio {
		ratio = ratioH
	}

	// 核心优化：避免小图被强行放大。如果 ratio > 1.0 说明目标尺寸大于原图
	if ratio >= 1.0 {
		return srcW, srcH
	}

	dstW := int(float64(srcW) * ratio)
	dstH := int(float64(srcH) * ratio)
	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}
	return dstW, dstH
}

// resizeImage 使用 BiLinear 进行高质量等比缩放
func resizeImage(srcImg image.Image, dstW, dstH int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	draw.BiLinear.Scale(dst, dst.Bounds(), srcImg, srcImg.Bounds(), stdDraw.Over, nil)
	return dst
}

// resizeFill 裁剪填充模式：先 mfit 缩放到覆盖目标区域，再居中裁剪
func resizeFill(srcImg image.Image, srcW, srcH, dstW, dstH int) image.Image {
	// 先缩放到覆盖目标区域
	scaleW, scaleH := calcMfit(srcW, srcH, dstW, dstH)
	scaled := resizeImage(srcImg, scaleW, scaleH)

	// 居中裁剪到目标尺寸
	offsetX := (scaleW - dstW) / 2
	offsetY := (scaleH - dstH) / 2
	cropRect := image.Rect(offsetX, offsetY, offsetX+dstW, offsetY+dstH)

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	stdDraw.Draw(dst, dst.Bounds(), scaled, cropRect.Min, stdDraw.Src)
	return dst
}

// encodeImage 编码图片为指定格式
func encodeImage(img image.Image, format string, quality int) ([]byte, string, error) {
	var buf bytes.Buffer
	var contentType string

	switch normalizeFormat(format) {
	case "webp":
		if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: float32(quality)}); err != nil {
			return nil, "", err
		}
		contentType = "image/webp"

	case "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, "", err
		}
		contentType = "image/jpeg"

	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		contentType = "image/png"

	default:
		// 默认用 JPEG (压缩率更好)
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, "", err
		}
		contentType = "image/jpeg"
	}

	return buf.Bytes(), contentType, nil
}

// normalizeFormat 统一格式名称
func normalizeFormat(format string) string {
	switch format {
	case "jpeg", "jpg", "JPEG", "JPG":
		return "jpg"
	case "png", "PNG":
		return "png"
	case "gif", "GIF":
		return "gif"
	case "webp", "WEBP":
		return "webp"
	default:
		return format
	}
}
