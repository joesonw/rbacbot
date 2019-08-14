package github

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
)

var (
	bytesNewline   = []byte("\n")
	bytesDiffMinus = []byte("--- ")
	bytesDiffPlus  = []byte("+++ ")
	bytesSlash     = []byte("/")
)

func readChangedFiles(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fileMap := make(map[string]bool)
	lines := bytes.Split(bodyBytes, bytesNewline)
	for _, line := range lines {
		if bytes.HasPrefix(line, bytesDiffMinus) {
			parts := bytes.Split(line[4:], bytesSlash)
			fileMap[string(bytes.Join(parts[1:], bytesSlash))] = true
		} else if bytes.HasPrefix(line, bytesDiffPlus) {
			parts := bytes.Split(line[4:], bytesSlash)
			fileMap[string(bytes.Join(parts[1:], bytesSlash))] = true
		}
	}

	files := make([]string, 0, len(fileMap))
	for file := range fileMap {
		files = append(files, file)
	}
	return files, nil
}
