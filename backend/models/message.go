package models

// ImageMessage 图片消息结构体
type ImageMessage struct {
	FromID      int64  `json:"fromId"`
	SessionID   int64  `json:"sessionId"`
	ToID        int64  `json:"toId"`
	SessionType string `json:"sessionType"`
	MsgID       int64  `json:"msgId"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	CreatedTime int64  `json:"createdTime"`
	Status      string `json:"status"`
	Extend      string `json:"extend"`
	FromName    string `json:"fromName"`
	Datetime    string `json:"datetime"`
}

// ImageContent 图片消息的 content JSON 结构
type ImageContent struct {
	ImgHeight    int    `json:"img_height"`
	ImgWidth     int    `json:"img_width"`
	LocalImgURL  string `json:"localimg_url"`
	ServerImgURL string `json:"serverimg_url"`
}

// ParsedImage 解析后的图片信息
type ParsedImage struct {
	Index        int    `json:"index"`
	MsgID        int64  `json:"msgId"`
	FromName     string `json:"fromName"`
	SessionID    int64  `json:"sessionId"`
	Datetime     string `json:"datetime"`
	ImgWidth     int    `json:"imgWidth"`
	ImgHeight    int    `json:"imgHeight"`
	ServerImgURL string `json:"serverImgUrl"`
	FileSize     int64  `json:"fileSize"`
	FileName     string `json:"fileName"`
}

// ParseResult 解析结果
type ParseResult struct {
	TaskID     string        `json:"taskId"`
	Sender     string        `json:"sender"`
	Recipients string        `json:"recipients"`
	TotalCount int           `json:"totalCount"`
	Images     []ParsedImage `json:"images"`
}

// DownloadRequest 下载请求
type DownloadRequest struct {
	TaskID  string `json:"taskId"`
	Indices []int  `json:"indices"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	Version   string `json:"version"`
	CommitID  string `json:"commitId"`
	BuildTime string `json:"buildTime"`
}
