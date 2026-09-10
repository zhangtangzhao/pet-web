package svc

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"pet/backend/internal/common"
)

// COSUploadToken 按 COS XML API 规范生成 PUT 对象的预签名直传地址（免 SDK 依赖）
func (sc *ServiceContext) COSUploadToken(dir, contentType string) (uploadURL, fileURL string, err error) {
	c := sc.Config.COS
	if c.Bucket == "" || c.Region == "" || c.SecretID == "" || c.SecretKey == "" {
		return "", "", common.NewErr(500, 50004, "对象存储未配置")
	}
	key := fmt.Sprintf("%s/%d%s", strings.Trim(dir, "/"), time.Now().UnixNano(), extOf(contentType))
	host := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", c.Bucket, c.Region)

	start := time.Now().Add(-time.Minute).Unix()
	end := time.Now().Add(15 * time.Minute).Unix()
	keyTime := fmt.Sprintf("%d;%d", start, end)

	signKey := hmacSha1Hex(c.SecretKey, keyTime)
	httpString := fmt.Sprintf("put\n/%s\n\n\n", key)
	stringToSign := fmt.Sprintf("sha1\n%s\n%s\n", keyTime, sha1Hex(httpString))
	signature := hmacSha1Hex(signKey, stringToSign)

	q := fmt.Sprintf(
		"q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=&q-url-param-list=&q-signature=%s",
		c.SecretID, keyTime, keyTime, signature)

	uploadURL = host + "/" + key + "?" + q
	fileURL = host + "/" + key
	return uploadURL, fileURL, nil
}

func extOf(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	default:
		return ".jpg"
	}
}

func hmacSha1Hex(key, data string) string {
	m := hmac.New(sha1.New, []byte(key))
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))
}

func sha1Hex(data string) string {
	h := sha1.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}
