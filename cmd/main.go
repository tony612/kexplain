package main

import (
	"explainx/mapper"
	"explainx/model"
	"explainx/view"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kubectl/pkg/util/openapi"
)

const (
	defaultTimeoutSeconds       = 3
	defaultGithubTimeoutSeconds = 5
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage: CMD RESOURCE, like CMD pods.spec.containers")
		os.Exit(0)
	}
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// First try fetching schema from k8s api
	doc, mapper, err := fetchFromK8s(*kubeconfig)
	if err != nil {
		// Then try remote(GitHub)
		doc, mapper, err = fetchFromRemote()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if doc == nil {
		fmt.Println("can't fetch schema from k8s and Remote like GitHub")
		os.Exit(1)
	}

	schema, err := openapi.NewOpenAPIData(doc)
	if err != nil {
		fmt.Println("can't get schema from spec")
		os.Exit(1)
	}

	err = run(schema, mapper, os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fetchFromRemote() (*openapi_v2.Document, mapper.Mapper, error) {
	client := &http.Client{Timeout: defaultGithubTimeoutSeconds * time.Second}
	resp, err := client.Get("https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json")
	if err != nil {
		return nil, nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	doc, err := openapi_v2.ParseDocument(data)
	if err != nil {
		return nil, nil, err
	}

	return doc, mapper.NewRawMapper(), nil
}

func fetchFromK8s(kubeconfig string) (*openapi_v2.Document, mapper.Mapper, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	config.Timeout = time.Second * defaultTimeoutSeconds
	config.QPS = 50
	config.Burst = 300

	client, err := discoveryClient(config)
	if err != nil {
		return nil, nil, err
	}

	schema, err := client.OpenAPISchema()
	if err != nil {
		return nil, nil, err
	}

	m, err := mapper.NewK8sMapper(client)
	if err != nil {
		return schema, nil, err
	}
	return schema, m, nil
}

func run(schema openapi.Resources, mapper mapper.Mapper, inResource string) error {
	resource, fieldsPath := splitDotNotation(inResource)
	gvk, err := mapper.KindFor(resource)
	if err != nil {
		return err
	}

	found := schema.LookupResource(gvk)
	if found == nil {
		return fmt.Errorf("couldn't find resource for %q", gvk)
	}

	err = render(fieldsPath, found, gvk)
	if err != nil {
		fmt.Printf("failed to render: %s", err)
	}

	return nil
}

func render(fieldsPath []string, schema proto.Schema, gvk schema.GroupVersionKind) error {
	doc, err := model.NewDoc(schema, fieldsPath, gvk)
	if err != nil {
		return err
	}

	app := tview.NewApplication()
	page := view.NewPage(doc)
	page.SetStopFn(func() { app.Stop() })
	if err := app.SetRoot(page, true).Run(); err != nil {
		return err
	}

	return nil
}

func splitDotNotation(model string) (string, []string) {
	var fieldsPath []string

	// ignore trailing period
	model = strings.TrimSuffix(model, ".")

	dotModel := strings.Split(model, ".")
	if len(dotModel) >= 1 {
		fieldsPath = dotModel[1:]
	}
	return dotModel[0], fieldsPath
}
