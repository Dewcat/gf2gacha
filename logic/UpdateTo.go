package logic

import (
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

func UpdateTo(version string) error {
	downloadURL, err := releaseDownloadURL(version)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		return err
	}

	err = restart()
	if err != nil {
		return err
	}
	os.Exit(0)

	return nil
}

func releaseDownloadURL(version string) (string, error) {
	if strings.TrimSpace(version) == "" {
		return "", errors.New("更新版本不能为空")
	}
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/gf2gacha.exe", updateRepository, url.PathEscape(version)), nil
}

func restart() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(execPath)
	return cmd.Start()
}
