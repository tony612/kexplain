package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"kexplain/pkg/mapper"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	"github.com/mitchellh/go-homedir"
	"k8s.io/kubectl/pkg/util/openapi"
)

const (
	defaultRemoteTimeoutSeconds = 5
	defaultRemoteURL            = "https://raw.githubusercontent.com/kubernetes/kubernetes/%s/api/openapi-spec/swagger.json"
	defaultCacheDir             = "~/.config/kexplain/cache"
	cacheFilePrefix             = "remote-"
	cacheTime                   = time.Hour * 24 * 7
)

func getFromRemote() (openapi.Resources, mapper.Mapper, error) {
	data, err := cacheOrFetch()
	if err != nil {
		return nil, nil, err
	}
	doc, err := openapi_v2.ParseDocument(data)
	if err != nil {
		return nil, nil, err
	}

	schema, err := openapi.NewOpenAPIData(doc)
	if err != nil {
		return nil, nil, err
	}

	return schema, mapper.NewRawMapper(), nil
}

func cacheOrFetch() ([]byte, error) {
	cacheDir, err := homedir.Expand(defaultCacheDir)
	if err != nil {
		return fetchFromRemote()
	}
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return fetchFromRemote()
	}

	filename := cacheName()
	p := path.Join(cacheDir, filename)
	file, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return fetchFromRemote()
	}

	defer file.Close()

	fetchFromRemoteAndSaveCache := func() ([]byte, error) {
		d, err := fetchFromRemote()
		if err != nil {
			return nil, err
		}
		if debug {
			log.Println("write to local cache using remote data")
		}
		_, err = file.Write(d)
		if err != nil {
			return nil, err
		}
		return d, nil
	}
	stat, err := file.Stat()
	// can't get mod time or Now - ModTime > cacheTime
	if err != nil || time.Since(stat.ModTime()) > cacheTime {
		return fetchFromRemoteAndSaveCache()
	}

	d, err := io.ReadAll(file)
	if err != nil {
		return fetchFromRemoteAndSaveCache()
	}
	if len(d) == 0 {
		return fetchFromRemoteAndSaveCache()
	}
	if debug {
		log.Println("use local cache as remote data")
	}
	return d, nil
}

func cacheName() string {
	sum := md5.Sum([]byte(remoteUrl()))
	return cacheFilePrefix + hex.EncodeToString(sum[:])
}

func fetchFromRemote() ([]byte, error) {
	if debug {
		log.Println("fetching doc from remote")
	}
	client := &http.Client{Timeout: defaultRemoteTimeoutSeconds * time.Second}
	resp, err := client.Get(remoteUrl())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func remoteUrl() string {
	if k8sVersion == "" {
		return fmt.Sprintf(defaultRemoteURL, "master")
	}

	return fmt.Sprintf(defaultRemoteURL, k8sVersion)
}
