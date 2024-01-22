package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DropboxAPIArg struct {
	Autorename      bool   `json:"autorename"`
	Mode            string `json:"mode"`
	Mute            bool   `json:"mute"`
	Path            string `json:"path"`
	Strict_conflict bool   `json:"strict_conflict"`
}

func Upload(localFilePath, dropboxFilePath string) error {
	token := os.Getenv("ANALYTICS_DROPBOX_ACCESS_TOKEN")

	arg := DropboxAPIArg{
		Autorename:      false,
		Mode:            "add",
		Mute:            false,
		Path:            dropboxFilePath,
		Strict_conflict: false,
	}

	argJson, _ := json.Marshal(arg)

	fileContent, err := os.ReadFile(localFilePath)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", bytes.NewBuffer(fileContent))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Dropbox-API-Arg", string(argJson))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error while reading response body: %s", err)
		} else {
			return fmt.Errorf("error uploading file to dropbox: %s", string(body))
		}
	}

	return nil
}
