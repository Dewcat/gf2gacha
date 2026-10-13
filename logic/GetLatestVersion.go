package logic

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

const updateRepository = "Dewcat/gf2gacha"

func GetLatestVersion() (string, error) {
	return getLatestVersion(&http.Client{Timeout: 15 * time.Second}, "https://api.github.com/repos/"+updateRepository+"/releases/latest")
}

func getLatestVersion(client *http.Client, endpoint string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gf2gacha-updater")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 新 fork 尚未发布 Release 时，GitHub 返回 404。
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name  string `json:"name"`
			State string `json:"state"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	// 仅在 Windows 程序已上传完成后提示更新。
	for _, asset := range release.Assets {
		if asset.Name == "gf2gacha.exe" && asset.State == "uploaded" {
			return release.TagName, nil
		}
	}
	return "", nil
}
