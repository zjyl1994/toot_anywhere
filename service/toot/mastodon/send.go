package mastodon

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/zjyl1994/toot_anywhere/service/config"
)

const requestClientTimeout = 10 * time.Second

func SendToot(content string, media []string) (string, error) {
	if len(media) > 0 {
		mediaIds := make([]string, 0, len(media))
		for _, mediaFile := range media {
			mediaId, err := uploadMediaGetId(mediaFile)
			if err != nil {
				return "", err
			}
			mediaIds = append(mediaIds, mediaId)
		}

		tootId, err := sendToot(content, mediaIds)
		if err != nil {
			return "", err
		}
		return tootId, nil
	} else {
		tootId, err := sendToot(content, nil)
		if err != nil {
			return "", err
		}
		return tootId, nil
	}
}

type sendTootResponse struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

func sendToot(content string, mediaIds []string) (string, error) {
	params := url.Values{}
	if len(content) > 0 {
		params.Set("status", content)
	}
	if len(mediaIds) > 0 {
		for _, mediaId := range mediaIds {
			params.Add("media_ids[]", mediaId)
		}
	}
	body := params.Encode()

	resp, err := sendRequest("/api/v1/statuses", "application/x-www-form-urlencoded", []byte(body))
	if err != nil {
		return "", err
	}
	var response sendTootResponse
	err = jsoniter.Unmarshal(resp, &response)
	if err != nil {
		return "", err
	}
	return response.ID, nil
}

type uploadMediaResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func uploadMediaGetId(mediaFile string) (string, error) {
	inputFile, err := os.Open(mediaFile)
	if err != nil {
		return "", err
	}
	defer inputFile.Close()

	mediaFilename := filepath.Base(mediaFile)

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", mediaFilename)
	if err != nil {
		return "", err
	}

	io.Copy(fileWriter, inputFile)
	bodyWriter.Close()

	contentType := bodyWriter.FormDataContentType()
	resp, err := sendRequest("/api/v1/media", contentType, bodyBuf.Bytes())
	if err != nil {
		return "", err
	}

	var response uploadMediaResponse
	err = jsoniter.Unmarshal(resp, &response)
	if err != nil {
		return "", err
	}
	return response.ID, nil
}

func sendRequest(path, contentType string, body []byte) ([]byte, error) {
	serverURL, err := config.Get("mastodon.url")
	if err != nil {
		return nil, fmt.Errorf("get mastodon.url failed: %w", err)
	}

	accessToken, err := config.Get("mastodon.access_token")
	if err != nil {
		return nil, fmt.Errorf("get mastodon.access_token failed: %w", err)
	}

	hc := http.Client{Timeout: requestClientTimeout}
	req, err := http.NewRequest(http.MethodPost, serverURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
