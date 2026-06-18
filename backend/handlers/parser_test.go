package handlers

import (
	"fmt"
	"os"
	"testing"
)

func TestParseHTML(t *testing.T) {
	sampleFiles := []string{
		"126.html",
		"28f2c174-2eef-490f-93c2-d09aa192136f.html",
		"6d2be2af-fc4f-4ea2-8601-48bab47586d4.html",
		"853c9c5f-2cee-4d2a-9c80-56e9684525bb.html",
		"8854bc5f-67e4-44c4-9df1-ecbed5c9045c.html",
	}

	for _, fname := range sampleFiles {
		t.Run(fname, func(t *testing.T) {
			testParseFile(t, fname)
		})
	}
}

func testParseFile(t *testing.T, fname string) {
	content, err := os.ReadFile("../../sample/" + fname)
	if err != nil {
		t.Skip("Sample file not available")
		return
	}

	htmlStr := string(content)

	// 测试提取 JS 数组
	jsArray, err := extractJSArray(htmlStr)
	if err != nil {
		t.Fatalf("extractJSArray failed: %v", err)
	}
	t.Logf("JS Array length: %d", len(jsArray))

	// 测试拆分对象
	objects := splitJSObjects(jsArray)
	t.Logf("Objects count: %d", len(objects))
	if len(objects) == 0 {
		t.Fatal("No objects found")
	}

	// 测试完整解析
	images, senders, recipients, err := parseHTMLToImagesDirect(htmlStr)
	if err != nil {
		t.Fatalf("parseHTMLToImagesDirect failed: %v", err)
	}

	t.Logf("Found %d images", len(images))
	t.Logf("Senders: %v", senders)
	t.Logf("Recipients: %v", recipients)

	if len(images) == 0 {
		t.Error("No images found - file may have no image messages or parsing may have issues")
	}

	for i, img := range images {
		if i >= 3 {
			break
		}
		t.Logf("Image %d: fromName=%s, datetime=%s, size=%dx%d, fileName=%s",
			i, img.FromName, img.Datetime,
			img.ImgWidth, img.ImgHeight, img.FileName)
	}

	// 验证必要字段
	for i, img := range images {
		if img.ServerImgURL == "" {
			t.Errorf("Image %d has empty URL", i)
		}
		if img.FromName == "" {
			t.Errorf("Image %d has empty sender", i)
		}
		if img.Datetime == "" {
			t.Errorf("Image %d has empty datetime", i)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestExtractJSField(t *testing.T) {
	// 测试基本字段提取
	obj := `{'fromId': 179378, 'sessionId': 855539, 'type': 'GROUP_IMAGE', 'fromName': '陈欢', 'datetime': '2026-06-10 06:16:54'}`

	tests := []struct {
		field string
		want  string
	}{
		{"type", "GROUP_IMAGE"},
		{"fromName", "陈欢"},
		{"datetime", "2026-06-10 06:16:54"},
		{"fromId", "179378"},
		{"sessionId", "855539"},
	}

	for _, tt := range tests {
		got := extractJSField(obj, tt.field)
		if got != tt.want {
			t.Errorf("extractJSField(%q) = %q, want %q", tt.field, got, tt.want)
		}
	}
}

func TestExtractJSONField(t *testing.T) {
	// 测试 content JSON 字段提取
	jsonStr := `{"img_height":717,"localimg_url":"/var/mobile/test.jpeg","img_width":1280,"serverimg_url":"https://example.com/image.jpeg"}`

	tests := []struct {
		field string
		want  string
	}{
		{"img_height", "717"},
		{"img_width", "1280"},
		{"serverimg_url", "https://example.com/image.jpeg"},
	}

	for _, tt := range tests {
		got := extractJSONField(jsonStr, tt.field)
		if got != tt.want {
			t.Errorf("extractJSONField(%q) = %q, want %q", tt.field, got, tt.want)
		}
	}
}

func Example_extractJSArray() {
	html := `<html><script>var array = [{'type': 'GROUP_IMAGE'}];</script></html>`
	arr, err := extractJSArray(html)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(arr)
	// Output: [{'type': 'GROUP_IMAGE'}]
}
