package i18n

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"
)

// Header is the JWT header data structure.
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// Payload is the JWT payload data structure and should keep sync with the
// starling server-side definition.
type Payload struct {
	ProjectName string `json:"project_name"`
	ExpiresAt   int64  `json:"expires_at"`
	Operator    string `json:"operator"`
	AccessType  string `json:"access_type"`
	UserAgent   string `json:"user_agent"`
}

// CreateAuthToken creates the jwt token based on the app key of the project.
func CreateAuthToken(key, projectName, operator string) string {
	header := &Header{
		Algorithm: "HS256",
		Type:      "JWT",
	}
	payload := &Payload{
		ProjectName: projectName,
		ExpiresAt:   time.Now().Unix() + 60,
		Operator:    operator,
		AccessType:  "SDK",
		UserAgent:   Platform,
	}
	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)
	headerStr := base64.URLEncoding.EncodeToString(headerBytes)
	payloadStr := base64.URLEncoding.EncodeToString(payloadBytes)

	signature := calcSignature(headerStr, payloadStr, key)
	return headerStr + "." + payloadStr + "." + signature
}

func calcSignature(header, payload, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(header))
	h.Write([]byte("."))
	h.Write([]byte(payload))
	h.Write([]byte("."))
	signature := h.Sum(nil)
	return hex.EncodeToString(signature)
}
