package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"xh-prase-backend/models"
)

// extractJSArray 从 HTML 中提取 JavaScript 数组字符串
func extractJSArray(htmlStr string) (string, error) {
	startMarker := "var array = "
	startIdx := strings.Index(htmlStr, startMarker)
	if startIdx == -1 {
		startMarker = "var array="
		startIdx = strings.Index(htmlStr, startMarker)
	}
	if startIdx == -1 {
		return "", fmt.Errorf("未找到消息数组变量")
	}

	startIdx += len(startMarker)

	runes := []rune(htmlStr[startIdx:])
	bracketCount := 0
	endIdx := -1
	inString := false
	var prev rune

	for i, ch := range runes {
		if ch == '\'' && prev != '\\' {
			inString = !inString
		}
		if !inString {
			if ch == '[' {
				bracketCount++
			} else if ch == ']' {
				bracketCount--
				if bracketCount == 0 {
					endIdx = i + 1
					break
				}
			}
		}
		prev = ch
	}

	if endIdx == -1 {
		return "", fmt.Errorf("未找到数组结束位置")
	}

	return string(runes[:endIdx]), nil
}

// parseHTMLToImagesDirect 直接从 HTML 解析图片消息（字段级提取，无需完整 JSON 转换）
func parseHTMLToImagesDirect(htmlStr string) ([]models.ParsedImage, map[string]bool, map[string]bool, error) {
	jsArray, err := extractJSArray(htmlStr)
	if err != nil {
		return nil, nil, nil, err
	}

	objects := splitJSObjects(jsArray)
	if len(objects) == 0 {
		return nil, nil, nil, fmt.Errorf("未找到消息对象")
	}

	var images []models.ParsedImage
	senders := make(map[string]bool)
	recipients := make(map[string]bool)

	for objIdx, objStr := range objects {
		msgType := extractJSField(objStr, "type")
		if msgType != "GROUP_IMAGE" {
			continue
		}

		fromName := extractJSField(objStr, "fromName")
		datetime := extractJSField(objStr, "datetime")
		msgIDStr := extractJSField(objStr, "msgId")
		sessionIDStr := extractJSField(objStr, "sessionId")
		contentStr := extractJSField(objStr, "content")

		msgID, _ := strconv.ParseInt(msgIDStr, 10, 64)
		sessionID, _ := strconv.ParseInt(sessionIDStr, 10, 64)

		// 从 content 的 JSON 中提取图片信息
		imgHeight := extractJSONField(contentStr, "img_height")
		imgWidth := extractJSONField(contentStr, "img_width")
		serverImgURL := extractJSONField(contentStr, "serverimg_url")

		if serverImgURL == "" {
			continue
		}

		height, _ := strconv.Atoi(imgHeight)
		width, _ := strconv.Atoi(imgWidth)

		datePart := strings.ReplaceAll(datetime, "-", "")
		if len(datePart) > 8 {
			datePart = datePart[:8]
		}
		fileName := fmt.Sprintf("%s_%s_%d.jpg",
			sanitizeFileName(fromName),
			datePart,
			msgID,
		)

		parsed := models.ParsedImage{
			Index:        objIdx,
			MsgID:        msgID,
			FromName:     fromName,
			SessionID:    sessionID,
			Datetime:     datetime,
			ImgWidth:     width,
			ImgHeight:    height,
			ServerImgURL: serverImgURL,
			FileName:     fileName,
		}

		images = append(images, parsed)
		senders[fromName] = true
		recipients[fmt.Sprintf("%d", sessionID)] = true
	}

	return images, senders, recipients, nil
}

// splitJSObjects 将 JS 数组拆分为单个对象字符串
func splitJSObjects(jsArray string) []string {
	var objects []string

	runes := []rune(jsArray)
	depth := 0
	start := -1
	inSingleQuote := false
	var prev rune

	for i, ch := range runes {
		if ch == '\'' && prev != '\\' {
			inSingleQuote = !inSingleQuote
		}
		if !inSingleQuote {
			if ch == '{' {
				if depth == 0 {
					start = i
				}
				depth++
			} else if ch == '}' {
				depth--
				if depth == 0 && start >= 0 {
					objects = append(objects, string(runes[start:i+1]))
					start = -1
				}
			}
		}
		prev = ch
	}

	return objects
}

// extractJSField 从 JS 对象字符串中提取字段值
// 格式: 'fieldName': 'value' 或 'fieldName': value
func extractJSField(objStr, fieldName string) string {
	pattern := fmt.Sprintf(`'%s'\s*:\s*`, regexp.QuoteMeta(fieldName))
	re := regexp.MustCompile(pattern)
	loc := re.FindStringIndex(objStr)
	if loc == nil {
		return ""
	}

	rest := objStr[loc[1]:]
	rest = strings.TrimSpace(rest)

	if len(rest) == 0 {
		return ""
	}

	// 转为 rune 切片以正确处理 UTF-8 多字节字符
	runes := []rune(rest)

	if runes[0] == '\'' {
		// 字符串值：找到结束单引号
		end := findClosingQuoteRunes(runes[1:])
		if end < 0 {
			return ""
		}
		val := string(runes[1 : end+1])
		val = strings.ReplaceAll(val, "\\/", "/")
		val = strings.ReplaceAll(val, "\\'", "'")
		return val
	}

	// 非字符串值：数字、布尔等
	end := strings.IndexAny(rest, ",}")
	if end < 0 {
		end = len(rest)
	}
	return strings.TrimSpace(rest[:end])
}

// findClosingQuoteRunes 在 rune 切片中找到匹配的结束单引号（考虑转义）
func findClosingQuoteRunes(runes []rune) int {
	var prev rune
	for i, ch := range runes {
		if ch == '\'' && prev != '\\' {
			return i
		}
		prev = ch
	}
	return -1
}

// extractJSONField 从 JSON 字符串中提取字段值
// 格式: "fieldName": "value" 或 "fieldName": value
func extractJSONField(jsonStr, fieldName string) string {
	pattern := fmt.Sprintf(`"%s"\s*:\s*`, regexp.QuoteMeta(fieldName))
	re := regexp.MustCompile(pattern)
	loc := re.FindStringIndex(jsonStr)
	if loc == nil {
		return ""
	}

	rest := jsonStr[loc[1]:]
	rest = strings.TrimSpace(rest)

	if len(rest) == 0 {
		return ""
	}

	runes := []rune(rest)

	if runes[0] == '"' {
		end := findClosingDoubleQuoteRunes(runes[1:])
		if end < 0 {
			return ""
		}
		val := string(runes[1 : end+1])
		val = strings.ReplaceAll(val, "\\/", "/")
		return val
	}

	// 数字值
	end := strings.IndexAny(rest, ",}")
	if end < 0 {
		end = len(rest)
	}
	return strings.TrimSpace(rest[:end])
}

// findClosingDoubleQuoteRunes 在 rune 切片中找到匹配的结束双引号
func findClosingDoubleQuoteRunes(runes []rune) int {
	var prev rune
	for i, ch := range runes {
		if ch == '"' && prev != '\\' {
			return i
		}
		prev = ch
	}
	return -1
}
