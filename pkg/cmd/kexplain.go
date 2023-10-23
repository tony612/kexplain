package cmd

import (
	"fmt"
	"kexplain/pkg/mapper"
	"kexplain/pkg/model"
	"kexplain/pkg/version"
	"kexplain/pkg/view"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kubectl/pkg/util/openapi"
)

const (
	defaultKubeTimeout = "5s"
)

var (
	longDoc = `List the fields for supported resources. An interactive "kubectl explain".

Use "kubectl api-resources" for a complete list of supported resources.

Global flags are from "kubectl options", but "--request-timeout" is changed to 5s by default. Remote doc like GitHub will be used
when k8s server is not accessible.
`
	cliUsage = `%[1]s <type>.<fieldName>[.<fieldName>]`

	cliExample = `
	# Get the documentation of the resource and its fields
	%[1]s pod

	# Get the documentation of a specific field of a resource
	%[1]s pod.spec.containers
`

	versionTemplate = `%[1]s {{printf "version %%s" .Version}}
`
)

var (
	debug      = false
	k8sVersion = ""
	remote     = false
)

type KexplainOptions struct {
	// k8s
	k8sConfigFlags *genericclioptions.ConfigFlags
	mapper         mapper.Mapper
	schema         openapi.Resources
	version        string

	args []string

	genericclioptions.IOStreams
}

func NewCmdKexplain(streams genericclioptions.IOStreams) *cobra.Command {
	o := newKexplainOptions(streams)
	cmdName := "kexplain"
	cmdBaseName := filepath.Base(os.Args[0])
	if strings.HasPrefix(cmdBaseName, "kubectl-") {
		cmdName = "kubectl " + strings.TrimPrefix(cmdBaseName, "kubectl-")
	}

	cmd := &cobra.Command{
		Use:          fmt.Sprintf(cliUsage, cmdName),
		Short:        "List the fields for supported resources",
		Long:         longDoc,
		Example:      fmt.Sprintf(cliExample, cmdName),
		SilenceUsage: true,
		Version:      version.FullVersion(),
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return c.Help()
			}
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.SetVersionTemplate(fmt.Sprintf(versionTemplate, strings.Replace(cmdName, " ", "-", 1)))
	o.k8sConfigFlags.AddFlags(cmd.InheritedFlags())
	cmd.Flags().BoolVar(&debug, "debug", false, "output debug log")
	cmd.Flags().BoolVar(&remote, "remote", false, "force to use remote doc instead of k8s server")
	cmd.Flags().StringVar(&k8sVersion, "k8s-version", "", "custom k8s version for fetching remote doc. Use latest by default")

	return cmd
}

func newKexplainOptions(streams genericclioptions.IOStreams) *KexplainOptions {
	opts := genericclioptions.NewConfigFlags(true)
	// Use "" to represent unset to distinguish with 0
	timeout := ""
	opts.Timeout = &timeout
	opts.Namespace = nil
	return &KexplainOptions{
		k8sConfigFlags: opts,

		IOStreams: streams,
	}
}

func (o *KexplainOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	var schema openapi.Resources
	var mapper mapper.Mapper
	var k8sErr error
	if remote {
		if debug {
			log.Println("get doc from remote directly")
		}
		var err error
		schema, mapper, err = getFromRemote()
		if debug {
			if err == nil {
				log.Println("done get schema from remote")
			} else {
				log.Printf("done get schema from remote, error: %s\n", err)
			}
		}
		if err != nil {
			return fmt.Errorf("fail to get schema\nfrom k8s: %v,\nand from remote: %w", k8sErr, err)
		}
	} else {
		if debug {
			log.Println("fetching k8s resources")
		}
		schema, mapper, k8sErr = o.getK8sResources()
		if k8sErr != nil {
			var err error
			if debug {
				log.Printf("fail to get k8s resources and get from remote: %s\n", k8sErr)
			}
			schema, mapper, err = getFromRemote()
			if debug {
				if err == nil {
					log.Println("done get schema from remote")
				} else {
					log.Printf("done get schema from remote, error: %s\n", err)
				}
			}
			if err != nil {
				return fmt.Errorf("fail to get schema\nfrom k8s: %v,\nand from remote: %w", k8sErr, err)
			}
		}
	}

	o.schema = schema
	o.mapper = mapper

	return nil
}

func (o *KexplainOptions) Validate() error {
	if len(o.args) == 0 {
		return fmt.Errorf("resource is needed like CMD pod.spec, you can use `kubectl api-resources` to get resources list")
	}
	if len(o.args) > 1 {
		return fmt.Errorf("either one or no arguments are allowed")
	}
	return nil
}

func (o *KexplainOptions) Run() error {
	resource, fieldsPath := splitDotNotation(o.args[0])
	gvk, err := o.mapper.KindFor(resource)
	if err != nil {
		return err
	}

	found := o.schema.LookupResource(gvk)
	if found == nil {
		return fmt.Errorf("couldn't find resource for %q", gvk)
	}

	v := o.version
	if v == "" {
		v = k8sVersion
	}
	err = render(fieldsPath, found, gvk, v)
	if err != nil {
		fmt.Printf("failed to render: %s", err)
	}
	return nil
}

func (o *KexplainOptions) getK8sResources() (openapi.Resources, mapper.Mapper, error) {
	if o.k8sConfigFlags.Timeout != nil && *o.k8sConfigFlags.Timeout == "" {
		timeout := defaultKubeTimeout
		o.k8sConfigFlags.Timeout = &timeout
	}
	discovery, err := o.k8sConfigFlags.ToDiscoveryClient()
	if err != nil {
		return nil, nil, fmt.Errorf("fail to get client: %w", err)
	}
	if v, err := discovery.ServerVersion(); err == nil {
		o.version = v.String()
	}
	schema, err := discovery.OpenAPISchema()
	if err != nil {
		return nil, nil, fmt.Errorf("fail to get schema: %w", err)
	}
	resources, err := openapi.NewOpenAPIData(schema)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to get resources from schema: %w", err)
	}
	k8sMapper, err := o.k8sConfigFlags.ToRESTMapper()
	if err != nil {
		return nil, nil, fmt.Errorf("fail to get rest mapper: %w", err)
	}
	return resources, mapper.NewK8sMapper(k8sMapper), nil
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

func render(fieldsPath []string, schema proto.Schema, gvk schema.GroupVersionKind, version string) error {
	doc, err := model.NewDoc(schema, fieldsPath, gvk)
	if err != nil {
		return err
	}

	app := tview.NewApplication()
	page := view.NewPage(doc)
	page.SetStopFn(func() { app.Stop() })
	page.SetVersion(version)
	if err := app.SetRoot(page, true).Run(); err != nil {
		return err
	}

	return nil
}
