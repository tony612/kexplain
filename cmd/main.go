package main

import (
	"explainx/model"
	"explainx/view"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kubectl/pkg/explain"
	"k8s.io/kubectl/pkg/util/openapi"
)

const defaultTimeoutSeconds = 3

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
	doc, mapper := fetchFromK8s(*kubeconfig)
	if doc == nil || mapper == nil {
		// Then try remote(GitHub)
		doc = fetchFromRemote()
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

func fetchFromRemote() *openapi_v2.Document {
	// TODO: fetch from https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json

	return nil
}

func fetchFromK8s(kubeconfig string) (*openapi_v2.Document, meta.RESTMapper) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil
	}

	config.Timeout = time.Second * defaultTimeoutSeconds
	config.QPS = 100
	config.Burst = 100

	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, nil
	}

	schema, err := client.OpenAPISchema()
	if err != nil {
		return nil, nil
	}

	apiresources, err := restmapper.GetAPIGroupResources(client)
	if err != nil {
		return schema, nil
	}
	mapper := restmapper.NewDiscoveryRESTMapper(apiresources)
	mapper = restmapper.NewShortcutExpander(mapper, client)
	return schema, mapper
}

func run(schema openapi.Resources, mapper meta.RESTMapper, resource string) error {
	fullySpecifiedGVR, fieldsPath, err := explain.SplitAndParseResourceRequest(resource, mapper)
	if err != nil {
		return err
	}

	gvk, _ := mapper.KindFor(fullySpecifiedGVR)
	if gvk.Empty() {
		gvk, err = mapper.KindFor(fullySpecifiedGVR.GroupResource().WithVersion(""))
		if err != nil {
			return err
		}
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
	if err := app.SetRoot(page, true).EnableMouse(true).Run(); err != nil {
		return err
	}

	return nil
}
