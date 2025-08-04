package dockerinternal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func fetchManifestDigestWithToken(url, token string) (string, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("User-Agent", "Go-Docker-Client/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("401 Unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch manifest: %s", resp.Status)
	}

	digest := resp.Header.Get("Docker-Content-Digest")
	if digest != "" {
		return digest, nil
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, resp.Body); err != nil {
		return "", fmt.Errorf("failed to read manifest body for hashing: %v", err)
	}

	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func getRegistryToken(manifestURL string) (string, error) {
	// First request to get the WWW-Authenticate header
	req, _ := http.NewRequest("GET", manifestURL, nil)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("User-Agent", "Go-Docker-Client/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	authHeader := resp.Header.Get("WWW-Authenticate")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("unexpected auth header format: %s", authHeader)
	}

	// Parse realm, service, scope
	params := map[string]string{}
	for _, field := range strings.Split(authHeader[len("Bearer "):], ",") {
		kv := strings.SplitN(strings.TrimSpace(field), "=", 2)
		if len(kv) == 2 {
			k := strings.Trim(kv[0], `"`)
			v := strings.Trim(kv[1], `"`)
			params[k] = v
		}
	}

	tokenURL := fmt.Sprintf("%s?service=%s&scope=%s", params["realm"], params["service"], params["scope"])
	tokenResp, err := http.Get(tokenURL)
	if err != nil {
		return "", err
	}
	defer tokenResp.Body.Close()

	var tokenData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		return "", err
	}

	return tokenData.Token, nil
}
