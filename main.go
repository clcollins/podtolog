package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultTemplate = `fetch logs //, scanLimitGBytes: 500, samplingRatio: 1000
	| filter matchesValue(k8s.pod.uid, "{{ .PodUID }}")
	| sort timestamp desc`

	defaultLogQueryPath = "ui/logs-events"
)

var kubeClient *dynamic.DynamicClient

type query struct {
	PodUID    string
	PodName   string
	Namespace string
	Shard     string
}

var podResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
var dynaKubeResource = schema.GroupVersionResource{Group: "dynatrace.com", Version: "v1alpha1", Resource: "dynakubes"}

const (
	dynaKubeNamespace = "dynatrace"
	dynaKubeName      = "request-serving"
	dynaKubeApiURL    = ".spec.apiUrl"
)

// Pass a pod name as argument; accept optional namespace flag
// If flag is not passed, use default namespace?
func main() {
	pflag.StringP("namespace", "n", "", "namespace {default: current namespace}")

	command := &cobra.Command{
		Use: "podtolog [-n NAMESPACE] (POD)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
			return nil
		},
		Run: func(c *cobra.Command, args []string) {

			pflag.Parse()

			viper.BindPFlags(pflag.CommandLine)
			namespace := viper.GetString("namespace")

			var query = query{
				PodName:   args[0],
				Namespace: namespace,
			}

			u, err := buildLogURL(query)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Print(u)
		},
	}

	command.Execute()
}

func buildLogURL(query query) (string, error) {
	var s string

	home, err := os.UserHomeDir()
	if err != nil {
		return s, err
	}

	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})

	currentNamespace, _, err := loader.Namespace()
	if err != nil {
		return s, err
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", home+"/.kube/config")
	if err != nil {
		return s, err
	}

	kubeClient, err = dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return s, err
	}

	if query.Namespace == "" {
		query.Namespace = currentNamespace
	}

	pod, err := kubeClient.Resource(podResource).Namespace(query.Namespace).Get(context.TODO(), query.PodName, metav1.GetOptions{})
	if err != nil {
		return s, err
	}

	query.PodUID = string(pod.GetUID())

	query.Shard, err = func(kubeClient *dynamic.DynamicClient) (string, error) {
		dk, err := kubeClient.Resource(dynaKubeResource).Namespace(dynaKubeNamespace).Get(context.TODO(), dynaKubeName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		u, err := url.Parse(dk.Object["spec"].(map[string]interface{})["apiUrl"].(string))
		if err != nil {
			return "", err
		}

		return u.Hostname(), nil
	}(kubeClient)
	if err != nil {
		return s, err
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", query.Shard))
	if err != nil {
		return s, err
	}

	u.Path = defaultLogQueryPath

	params := url.Values{}
	params.Add("gtf", "-2h")
	params.Add("gf", "all")
	params.Add("sortDirection", "desc")
	params.Add("visibleColumns", "timestamp")
	params.Add("visibleColumns", "status")
	params.Add("visibleColumns", "content")
	params.Add("advancedQueryMode", "true")
	params.Add("visualizationType", "table")
	params.Add("isDefaultQuery", "true")
	u.RawQuery = params.Encode()

	// Create the log query string
	q, err := parseTemplate(defaultTemplate, query)
	if err != nil {
		return s, err
	}

	// URL encode the query string
	eq := url.PathEscape(q.String())

	// Base64 encode the query string
	bq := base64.RawStdEncoding.EncodeToString([]byte(eq))

	return fmt.Sprintf("%s#%s\n", u.String(), bq), err

}

func parseTemplate(templateString string, query query) (*bytes.Buffer, error) {
	var queryString bytes.Buffer

	t, err := template.New("default").Parse(templateString)
	if err != nil {
		return &queryString, err
	}
	t.Execute(&queryString, query)
	return &queryString, nil
}
