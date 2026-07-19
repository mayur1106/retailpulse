package aws

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Credentials struct {
	AccessKey    string
	SecretKey    string
	SessionToken string
	Region       string
	Service      string
}

type SigV4Signer struct {
	credentials Credentials
}

func NewSigV4Signer(credentials Credentials) SigV4Signer {
	if credentials.Service == "" {
		credentials.Service = "execute-api"
	}
	return SigV4Signer{credentials: credentials}
}

func (s SigV4Signer) Sign(request *http.Request, now time.Time) error {
	if s.credentials.AccessKey == "" || s.credentials.SecretKey == "" || s.credentials.Region == "" {
		return errors.New("AWS credentials for SP-API request signing are not configured")
	}
	requestTime := now.UTC()
	amzDate := requestTime.Format("20060102T150405Z")
	dateStamp := requestTime.Format("20060102")

	request.Header.Set("Host", request.URL.Host)
	request.Header.Set("X-Amz-Date", amzDate)
	if s.credentials.SessionToken != "" {
		request.Header.Set("X-Amz-Security-Token", s.credentials.SessionToken)
	}

	signedHeaders := signedHeaderNames(request)
	canonicalRequest := strings.Join([]string{
		request.Method,
		canonicalURI(request.URL.EscapedPath()),
		request.URL.RawQuery,
		canonicalHeaders(request, signedHeaders),
		strings.Join(signedHeaders, ";"),
		hexSHA256(""),
	}, "\n")

	scope := strings.Join([]string{dateStamp, s.credentials.Region, s.credentials.Service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		hexSHA256(canonicalRequest),
	}, "\n")

	signingKey := deriveSigningKey(s.credentials.SecretKey, dateStamp, s.credentials.Region, s.credentials.Service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	authorization := "AWS4-HMAC-SHA256 Credential=" + s.credentials.AccessKey + "/" + scope +
		", SignedHeaders=" + strings.Join(signedHeaders, ";") +
		", Signature=" + signature
	request.Header.Set("Authorization", authorization)
	return nil
}

func canonicalURI(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func signedHeaderNames(request *http.Request) []string {
	headers := make([]string, 0, len(request.Header))
	for name := range request.Header {
		headers = append(headers, strings.ToLower(name))
	}
	sort.Strings(headers)
	return headers
}

func canonicalHeaders(request *http.Request, signedHeaders []string) string {
	var builder strings.Builder
	for _, name := range signedHeaders {
		values := request.Header.Values(name)
		normalized := make([]string, 0, len(values))
		for _, value := range values {
			normalized = append(normalized, strings.Join(strings.Fields(value), " "))
		}
		builder.WriteString(name)
		builder.WriteString(":")
		builder.WriteString(strings.Join(normalized, ","))
		builder.WriteString("\n")
	}
	return builder.String()
}

func hexSHA256(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func deriveSigningKey(secret string, dateStamp string, region string, service string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secret), dateStamp)
	regionKey := hmacSHA256(dateKey, region)
	serviceKey := hmacSHA256(regionKey, service)
	return hmacSHA256(serviceKey, "aws4_request")
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))
	return mac.Sum(nil)
}
