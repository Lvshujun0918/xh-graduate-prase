package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"xh-prase-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	taskStore = sync.Map{}
	// 构建信息，由 main 包设置
	buildCommitID  = "unknown"
	buildTimestamp = "unknown"
)

// SetBuildInfo 设置构建信息
func SetBuildInfo(commitID, buildTime string) {
	if commitID != "" {
		buildCommitID = commitID
	}
	if buildTime != "" {
		buildTimestamp = buildTime
	}
}

// UploadHTML 上传并解析 HTML 文件
func UploadHTML(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传文件"})
		return
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".html") &&
		!strings.HasSuffix(strings.ToLower(file.Filename), ".htm") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传 HTML 文件"})
		return
	}

	tmpPath := filepath.Join("uploads", uuid.New().String()+".html")
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	htmlStr := string(content)

	images, senders, recipients, err := parseHTMLToImagesDirect(htmlStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析文件失败: " + err.Error()})
		return
	}

	if len(images) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "未找到图片消息"})
		return
	}

	taskID := uuid.New().String()

	senderList := make([]string, 0, len(senders))
	for s := range senders {
		senderList = append(senderList, s)
	}
	recipientList := make([]string, 0, len(recipients))
	for r := range recipients {
		recipientList = append(recipientList, r)
	}

	savedDir := "saved_html"
	os.MkdirAll(savedDir, 0755)
	savedName := fmt.Sprintf("%s_%s_%s.html",
		sanitizeFileName(strings.Join(senderList, "_")),
		strings.Join(recipientList, "_"),
		taskID[:8],
	)
	savedPath := filepath.Join(savedDir, savedName)
	if err := os.WriteFile(savedPath, content, 0644); err != nil {
		log.Printf("保存重命名文件失败: %v", err)
	}

	result := models.ParseResult{
		TaskID:     taskID,
		Sender:     strings.Join(senderList, ", "),
		Recipients: strings.Join(recipientList, ", "),
		TotalCount: len(images),
		Images:     images,
	}

	taskStore.Store(taskID, result)
	os.Remove(tmpPath)

	c.JSON(http.StatusOK, result)
}

// GetImageInfo 获取已解析的图片信息
func GetImageInfo(c *gin.Context) {
	taskID := c.Param("id")
	val, ok := taskStore.Load(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, val)
}

// ProxyImage 代理获取远程图片
func ProxyImage(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少图片 URL"})
		return
	}

	decodedURL, err := url.QueryUnescape(imageURL)
	if err != nil {
		decodedURL = imageURL
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", decodedURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取图片失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Cache-Control", "public, max-age=86400")
	io.Copy(c.Writer, resp.Body)
}

// DownloadImages 下载图片
func DownloadImages(c *gin.Context) {
	var req models.DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	val, ok := taskStore.Load(req.TaskID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	result := val.(models.ParseResult)

	var selectedImages []models.ParsedImage
	if len(req.Indices) == 0 {
		selectedImages = result.Images
	} else {
		indexSet := make(map[int]bool)
		for _, idx := range req.Indices {
			indexSet[idx] = true
		}
		for i, img := range result.Images {
			if indexSet[i] {
				selectedImages = append(selectedImages, img)
			}
		}
	}

	if len(selectedImages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未选择任何图片"})
		return
	}

	if len(selectedImages) == 1 {
		serveSingleImage(c, selectedImages[0])
		return
	}

	serveZipArchive(c, selectedImages)
}

func serveSingleImage(c *gin.Context, img models.ParsedImage) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("GET", img.ServerImgURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下载图片失败"})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", "image/jpeg")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, img.FileName))
	io.Copy(c.Writer, resp.Body)
}

func serveZipArchive(c *gin.Context, images []models.ParsedImage) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	client := &http.Client{Timeout: 60 * time.Second}

	for _, img := range images {
		req, err := http.NewRequest("GET", img.ServerImgURL, nil)
		if err != nil {
			log.Printf("创建请求失败 %s: %v", img.FileName, err)
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("下载失败 %s: %v", img.FileName, err)
			continue
		}

		entry, err := zipWriter.Create(img.FileName)
		if err != nil {
			resp.Body.Close()
			log.Printf("创建 ZIP 条目失败 %s: %v", img.FileName, err)
			continue
		}

		io.Copy(entry, resp.Body)
		resp.Body.Close()
	}

	zipWriter.Close()

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="images_%s.zip"`, time.Now().Format("20060102_150405")))
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"\\", "_", "/", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// GetVersion 返回版本信息
func GetVersion(c *gin.Context) {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}
	commitID := buildCommitID
	buildTime := buildTimestamp

	c.JSON(http.StatusOK, models.VersionInfo{
		Version:   version,
		CommitID:  commitID,
		BuildTime: buildTime,
	})
}
