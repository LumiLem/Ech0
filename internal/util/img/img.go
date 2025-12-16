package util

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
)

// var (
// 	vipsOnce    sync.Once
// 	vipsInitErr error
// )

// func vipsInit() error {
// 	vipsOnce.Do(func() {
// 		// Startup 会检查版本并初始化 libvips
// 		vips.Startup(nil)
// 	})
// 	return vipsInitErr
// }

// GetImageSize 只读取图片头部获取尺寸，避免加载整图与 CGO 依赖
func GetImageSizeFromPath(path string) (width, height int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = f.Close()
	}()

	return GetImageSizeFromReader(f)
}

// GetImageSizeFromFile 从文件获取图片尺寸
func GetImageSizeFromFile(file *multipart.FileHeader) (width, height int, err error) {
	reader, err := file.Open()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = reader.Close()
	}()
	return GetImageSizeFromReader(reader)
}

// GetImageSizeFromReader 从 Reader 获取图片尺寸
func GetImageSizeFromReader(reader io.Reader) (width, height int, err error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return 0, 0, err
	}
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty image data")
	}

	// 先尝试标准库（支持 png/jpeg/gif）
	if cfg, _, stdErr := image.DecodeConfig(bytes.NewReader(data)); stdErr == nil {
		return cfg.Width, cfg.Height, nil
	}

	return 0, 0, nil

	// 回退用 libvips 支持 webp/avif 等
	// if err := vipsInit(); err != nil {
	// 	return 0, 0, err
	// }
	// img, err := vips.NewImageFromBuffer(data, nil)
	// if err != nil {
	// 	return 0, 0, err
	// }
	// defer img.Close()

	// return img.Width(), img.Height(), nil
}

// // ConvertImage 转换图片格式
// func ConvertImage(path, outputFormat string) error {
// 	if err := vipsInit(); err != nil {
// 		return err
// 	}

// 	img, err := vips.NewImageFromFile(path, nil)
// 	if err != nil {
// 		return err
// 	}
// 	defer img.Close()

// 	format := strings.TrimPrefix(strings.ToLower(outputFormat), ".")
// 	if format == "" {
// 		return fmt.Errorf("output format required")
// 	}

// 	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
// 	outPath := filepath.Join(filepath.Dir(path), base+"."+format)

// 	switch format {
// 	case "webp":
// 		opts := vips.DefaultWebpsaveOptions()
// 		opts.Q = 85
// 		return img.Webpsave(outPath, opts)
// 	case "avif":
// 		opts := vips.DefaultHeifsaveOptions()
// 		opts.Q = 80
// 		opts.Compression = vips.HeifCompressionAv1
// 		opts.Encoder = vips.HeifEncoderAom
// 		return img.Heifsave(outPath, opts)
// 	case "png":
// 		return img.Pngsave(outPath, nil)
// 	case "jpeg", "jpg":
// 		opts := vips.DefaultJpegsaveOptions()
// 		opts.Q = 85
// 		return img.Jpegsave(outPath, opts)
// 	default:
// 		return fmt.Errorf("unsupported format: %s", outputFormat)
// 	}
// }

// // ConvertImageFromFile 转换图片格式并返回新的文件
// func ConvertImageFromFile(file *multipart.FileHeader, outputFormat string) (newFile *multipart.FileHeader, err error) {
// 	if err := vipsInit(); err != nil {
// 		return nil, err
// 	}

// 	src, err := file.Open()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer func() {
// 		_ = src.Close()
// 	}()

// 	data, err := io.ReadAll(src)
// 	if err != nil {
// 		return nil, err
// 	}
// 	img, err := vips.NewImageFromBuffer(data, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer img.Close()

// 	format := strings.TrimPrefix(strings.ToLower(outputFormat), ".")
// 	if format == "" {
// 		return nil, fmt.Errorf("output format required")
// 	}

// 	var (
// 		buf []byte
// 		ct  string
// 	)

// 	switch format {
// 	case "webp":
// 		opts := vips.DefaultWebpsaveBufferOptions()
// 		opts.Q = 85
// 		buf, err = img.WebpsaveBuffer(opts)
// 		ct = "image/webp"
// 	case "avif":
// 		opts := vips.DefaultHeifsaveBufferOptions()
// 		opts.Q = 80
// 		opts.Compression = vips.HeifCompressionAv1
// 		opts.Encoder = vips.HeifEncoderAom
// 		buf, err = img.HeifsaveBuffer(opts)
// 		ct = "image/avif"
// 	case "png":
// 		buf, err = img.PngsaveBuffer(nil)
// 		ct = "image/png"
// 	case "jpeg", "jpg":
// 		opts := vips.DefaultJpegsaveBufferOptions()
// 		opts.Q = 85
// 		buf, err = img.JpegsaveBuffer(opts)
// 		ct = "image/jpeg"
// 	default:
// 		return nil, fmt.Errorf("unsupported format: %s", outputFormat)
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	filename := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename)) + "." + format

// 	// 将 buffer 包装成 multipart.FileHeader，便于复用上传逻辑
// 	var b bytes.Buffer
// 	writer := multipart.NewWriter(&b)
// 	hdr := make(textproto.MIMEHeader)
// 	hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename))
// 	hdr.Set("Content-Type", ct)
// 	part, err := writer.CreatePart(hdr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if _, err = part.Write(buf); err != nil {
// 		return nil, err
// 	}
// 	if err = writer.Close(); err != nil {
// 		return nil, err
// 	}

// 	reader := multipart.NewReader(&b, writer.Boundary())
// 	form, err := reader.ReadForm(int64(len(buf) + 1024))
// 	if err != nil {
// 		return nil, err
// 	}
// 	fhs := form.File["file"]
// 	if len(fhs) == 0 {
// 		return nil, fmt.Errorf("converted file header missing")
// 	}
// 	return fhs[0], nil
// }
