package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultTemplate = `	fetch logs //, scanLimitGBytes: 500, samplingRatio: 1000
	| filter matchesValue(k8s.pod.uid, "{{ .PodUID }}")
	| sort timestamp desc`

	defaultLogQueryPath = "ui/logs-events"
)

type query struct {
	PodUID string
}

func main() {
	viper.SetDefault("Flag", "Flag Value")

	command := &cobra.Command{
		Run: func(c *cobra.Command, args []string) {

			var query = query{
				PodUID: args[0],
			}

			fmt.Printf("Query: %+v\n", query)
			u, err := buildLogURL(query)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Print(u)
		},
	}

	command.Execute()
}

func buildLogURL(queryData query) (string, error) {
	var s string

	host, err := getConfig()
	if err != nil {
		return s, err
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", host))
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
	q, err := parseTemplate(defaultTemplate, queryData)
	if err != nil {
		return s, err
	}

	// fmt.Println(q.String())

	eq := url.PathEscape(q.String())
	// fmt.Println(eq)

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

func getConfig() (string, error) {
	var host string

	viper.SetEnvPrefix("podtolog")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/podtolog")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Printf("WARNING: %s\n", err)
		} else {
			// Config file was found but another error was produced
			return host, err
		}
	}

	viper.AutomaticEnv()

	host = viper.GetString("host")

	if host == "" {
		return host, fmt.Errorf("cannot find 'PODTOLOG_HOST' environment variable or 'host' value in %s", viper.ConfigFileUsed())
	}

	return host, nil
}
