package lib

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
)

func ChunkArray[T any](array []T, size int) [][]T {
	finalArray := [][]T{}
	tempArr := []T{}
	for _, item := range array {
		if len(tempArr) == size {
			finalArray = append(finalArray, tempArr)
			tempArr = []T{}

		}
		tempArr = append(tempArr, item)
	}
	finalArray = append(finalArray, tempArr)

	return finalArray
}

func Handlepanic(errorContext string) {
	if a := recover(); a != nil {
		stck := debug.Stack()
		CaptureSentryException(fmt.Sprintf("RECOVER from error at %s:  %s", errorContext, stck))
	}
}

func GenerateRandomUUID() string {
	// Create a buffer to hold the random bytes
	buf := make([]byte, 16)

	// Read random bytes from the crypto/rand package
	_, err := rand.Read(buf)
	if err != nil {
		CaptureSentryException(fmt.Sprintf("Error reading from buffer %s", err))
	}

	// Set the UUID version (4 for random) and variant bits
	// According to the UUID format, version 4 has the 4 most significant bits set to 0100 (0x40),
	// and the variant bits are set to 10xx (0x80, 0x90, 0xA0, or 0xB0).
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80

	// Format the bytes as a string representation of a UUID
	uuidStr := fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:])

	return uuidStr
}

func StringLenGtZero(str string) bool {
	return len(strings.TrimSpace(str)) > 0
}

func PlaceGetReq(req *Request, url string, params map[string]string, token string) (*[]byte, error) {

	httpReq, httpErr := http.NewRequest("GET", url, nil)
	if httpErr != nil {
		CaptureSentryException(fmt.Sprintf("ERR: %s URL %s failed with error %s", req.ID, url, httpErr))
		return nil, errors.New("Something went wrong")
	}

	httpReq.Header.Set("Authorization", token)

	q := httpReq.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	httpReq.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, doErr := client.Do(httpReq)

	if doErr != nil || resp.StatusCode != 200 {
		CaptureSentryException(fmt.Sprintf("%s Error: Portfolio service could be down", req.ID))
		CaptureSentryException(fmt.Sprintf("%s Error: client.do failed on url %s with status.code %d with error %s", req.ID, url, resp.StatusCode, doErr))
		return nil, errors.New("Something went wrong")
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		CaptureSentryException(fmt.Sprintf("%s ioutil.ReadAll failed with error %s", req.ID, readErr))
		return nil, errors.New("Something went wrong")
	}

	return &body, nil
}

func GoFuncWrapper(title string, callback func()) {
	go func() {
		Handlepanic(title)
		callback()
	}()
}

func SnakeToUpperCamel(input string) string {
	words := strings.Split(input, "_")

	for i := 0; i < len(words); i++ {
		words[i] = strings.Title(words[i])
	}

	result := strings.Join(words, "")

	return result
}
