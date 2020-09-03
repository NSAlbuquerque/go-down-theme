package common

import (
	"errors"
	"io"
	"net/http"
)

// Download over http.
func Download(url string, to io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("request error: " + resp.Status)
	}

	_, err = io.Copy(to, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
