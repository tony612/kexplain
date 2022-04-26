package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"kexplain/pkg/mapper"
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
	defaultRemoteURL            = "https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json"
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
		// TODO: what to do when write fails
		file.Write(d)
		return d, nil
	}
	stat, err := file.Stat()
	// can't get mod time or Now - ModTime > cacheTime
	if err != nil || time.Since(stat.ModTime()) > cacheTime {
		return fetchFromRemoteAndSaveCache()
	}

	d, err := ioutil.ReadAll(file)
	if err != nil {
		return fetchFromRemoteAndSaveCache()
	}
	if len(d) == 0 {
		return fetchFromRemoteAndSaveCache()
	}
	return d, nil
}

func cacheName() string {
	sum := md5.Sum([]byte(defaultRemoteURL))
	return cacheFilePrefix + hex.EncodeToString(sum[:])
}

func fetchFromRemote() ([]byte, error) {
	client := &http.Client{Timeout: defaultRemoteTimeoutSeconds * time.Second}
	resp, err := client.Get(defaultRemoteURL)
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
