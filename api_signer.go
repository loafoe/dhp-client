package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

const (
	ALGORITHM_NAME    = "HmacSHA256"
	SECRET_KEY_PREFIX = "DHPWS"
)

type ApiSigner struct {
	secretKey string
	sharedKey string
	debug     bool
}

func (signer *ApiSigner) Init(sharedKey, secretKey string, debug bool) {
	signer.secretKey = secretKey
	signer.sharedKey = sharedKey
	signer.debug = debug
}

func (signer *ApiSigner) BuildAuthorizationHeaderValue(requestMethod, queryString string, header *http.Header, url string, body []byte) string {
	joinedHeaders := joinHeaders(header)
	signatureKey := signer.hashRequest(requestMethod, queryString, body, joinedHeaders)
	signature := signString(signatureKey, url)

	return signer.buildAuthorizationHeaderValue(joinedHeaders, signature)
}

func (signer *ApiSigner) buildAuthorizationHeaderValue(requestHeader, signature string) string {
	buffer := bytes.NewBufferString(ALGORITHM_NAME)
	buffer.WriteString(";")
	buffer.WriteString("Credential:")
	buffer.WriteString(signer.sharedKey)
	buffer.WriteString(";")
	buffer.WriteString("SignedHeaders:")
	headers := strings.Split(requestHeader, ";")
	for i, v := range headers {
		headerName := strings.Split(v, ":")[0]
		if i > 0 && len(headerName) > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(headerName)
	}
	buffer.WriteString(";")
	buffer.WriteString("Signature:")
	buffer.WriteString(signature)
	return buffer.String()
}

func (signer *ApiSigner) hashRequest(requestMethod, queryString string, body []byte, requestHeaders string) []byte {
	kSecret := []byte(SECRET_KEY_PREFIX + signer.secretKey)
	if signer.debug {
		fmt.Printf("kSecret = %v\n", kSecret)
	}
	kMethod := hash([]byte(requestMethod), kSecret)
	if signer.debug {
		fmt.Printf("kMethod = %v\n", kMethod)
	}
	kQueryString := hash([]byte(queryString), kMethod)
	if signer.debug {
		fmt.Printf("kQueryString = %v\n", kQueryString)
		fmt.Printf("body = %v\n", body)
		fmt.Printf("bodyString = %s\n", string(body))
	}
	kBody := hash(body, kQueryString)
	if signer.debug {
		fmt.Printf("kBody = %v\n", kBody)
		fmt.Printf("requestHeader = %s\n", requestHeaders)
	}
	hashed := hash([]byte(requestHeaders), kBody)
	if signer.debug {
		fmt.Printf("hashed = %v\n", hashed)
	}
	return hashed
}

func joinHeaders(header *http.Header) string {
	var headers []string
	var keys []string
	for key := range *header {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, k := range keys {
		h := k
		if h == "Signeddate" { // Ugly hack
			h = "SignedDate"
		}
		headers = append(headers, h+":"+header.Get(h))
	}
	joined := fmt.Sprintf("%s;", strings.Join(headers, ";"))
	return joined
}

func signString(signatureKey []byte, uriToBeSigned string) string {
	signatureSlice := hash([]byte(uriToBeSigned), signatureKey)
	return base64.StdEncoding.EncodeToString(signatureSlice)
}

func hash(data []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
