package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"xh-prase-backend/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 通过 ldflags 在构建时注入
var (
	commitID  = "unknown"
	buildTime = "unknown"
)

func main() {
	// 设置版本信息到 handlers
	handlers.SetBuildInfo(commitID, buildTime)

	// 确保上传目录存在
	os.MkdirAll("uploads", 0755)
	os.MkdirAll("saved_html", 0755)

	// 根据环境变量设置 gin 模式（默认开发模式）
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug" // 默认开发模式，可以看到详细日志
	}
	gin.SetMode(ginMode)
	log.Printf("Gin 运行模式: %s", ginMode)

	r := gin.Default()

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
	}))

	// 限制上传文件大小
	r.MaxMultipartMemory = 100 << 20 // 100 MB

	// API 路由
	api := r.Group("/api")
	{
		api.POST("/upload", handlers.UploadHTML)
		api.GET("/images/:id", handlers.GetImageInfo)
		api.GET("/proxy-image", handlers.ProxyImage)
		api.POST("/download", handlers.DownloadImages)
		api.GET("/version", handlers.GetVersion)
	}

	// 静态文件服务（生产模式）
	frontendDist := "frontend/dist"
	if _, err := os.Stat(frontendDist); err == nil {
		log.Println("检测到前端构建产物，启用静态文件服务")
		r.Use(staticFileMiddleware(frontendDist))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8282"
	}
	log.Printf("服务启动在端口 %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// staticFileMiddleware 提供前端静态文件服务，支持 SPA 路由回退
func staticFileMiddleware(frontendDir string) gin.HandlerFunc {
	fileServer := http.FileServer(http.Dir(frontendDir))

	return func(c *gin.Context) {
		// 跳过 API 路由
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		fullPath := filepath.Join(frontendDir, path)

		// 如果文件存在，直接提供服务
		if _, err := os.Stat(fullPath); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// SPA 回退：所有非 API 路由返回 index.html
		c.File(filepath.Join(frontendDir, "index.html"))
		c.Abort()
	}
}
