package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Namespace          string
	JobID              string
	Command            string
	Subscriptions      string
	Timeout            string
	RuntimeAssets      string
	SensuAPIUrl        string
	SensuAccessToken   string
	SensuTrustedCaFile string
}

// JobRequest represents a job request.
type JobRequest struct {
	Check         string   `json:"check"`
	Subscriptions []string `json:"subscriptions"`
}

var (
	config = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-runbook",
			Short:    "Sensu Runbook Automation. Execute commands on Sensu Agent nodes.",
			Keyspace: "sensu.io/plugins/sensu-runbook/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "id",
			Env:       "SENSU_RUNBOOK_JOB_ID",
			Argument:  "id",
			Shorthand: "i",
			Default:   uuid.New().String(),
			Usage:     "The ID or name to use for the job (i.e. defaults to a random UUIDv4)",
			Value:     &config.JobID,
		},
		{
			Path:      "command",
			Env:       "SENSU_RUNBOOK_COMMAND",
			Argument:  "command",
			Shorthand: "c",
			Default:   "",
			Usage:     "The command that should be executed by the Sensu Go agent(s)",
			Value:     &config.Command,
		},
		{
			Path:      "timeout",
			Env:       "SENSU_RUNBOOK_TIMEOUT",
			Argument:  "timeout",
			Shorthand: "t",
			Default:   "10",
			Usage:     "Command execution timeout, in seconds",
			Value:     &config.Command,
		},
		{
			Path:      "runtime-assets",
			Env:       "SENSU_RUNBOOK_ASSETS",
			Argument:  "runtime-assets",
			Shorthand: "a",
			Default:   "",
			Usage:     "Comma-separated list of assets to distribute with the command(s)",
			Value:     &config.RuntimeAssets,
		},
		{
			Path:      "subscriptions",
			Env:       "SENSU_RUNBOOK_SUBSCRIPTIONS",
			Argument:  "subscriptions",
			Shorthand: "s",
			Default:   "",
			Usage:     "Comma-separated list of subscriptions to execute the command(s) on",
			Value:     &config.Subscriptions,
		},
		{
			Path:      "namespace",
			Env:       "SENSU_NAMESPACE", // provided by the sensuctl command plugin execution environment
			Argument:  "namespace",
			Shorthand: "n",
			Default:   "",
			Usage:     "Sensu Namespace to perform the runbook automation (defaults to $SENSU_NAMESPACE)",
			Value:     &config.Namespace,
		},
		{
			Path:      "sensu-api-url",
			Env:       "SENSU_API_URL", // provided by the sensuctl command plugin execution environment
			Argument:  "sensu-api-url",
			Shorthand: "",
			Default:   "",
			Usage:     "Sensu API URL (defaults to $SENSU_API_URL)",
			Value:     &config.SensuAPIUrl,
		},
		{
			Path:      "sensu-access-token",
			Env:       "SENSU_ACCESS_TOKEN", // provided by the sensuctl command plugin execution environment
			Argument:  "sensu-access-token",
			Shorthand: "",
			Default:   "",
			Usage:     "Sensu API Access Token (defaults to $SENSU_ACCESS_TOKEN)",
			Value:     &config.SensuAccessToken,
		},
		{
			Path:      "sensu-trusted-ca-file",
			Env:       "SENSU_TRUSTED_CA_FILE", // provided by the sensuctl command plugin execution environment
			Argument:  "sensu-trusted-ca-file",
			Shorthand: "",
			Default:   "",
			Usage:     "Sensu API Trusted Certificate Authority File (defaults to $SENSU_TRUSTED_CA_FILE)",
			Value:     &config.SensuTrustedCaFile,
		},
	}
)

func main() {
	plugin := sensu.NewGoCheck(&config.PluginConfig, options, checkArgs, executePlaybook, false)
	plugin.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	if len(config.SensuAPIUrl) == 0 {
		return sensu.CheckStateCritical, errors.New("--sensu-api-url flag or $SENSU_API_URL environment variable must be set")
	} else if len(config.Namespace) == 0 {
		return sensu.CheckStateCritical, errors.New("--namespace flag or $SENSU_NAMESPACE environment variable must be set")
	} else if len(config.Command) == 0 {
		return sensu.CheckStateWarning, errors.New("--command flag or $SENSU_RUNBOOK_COMMAND environment variable must be set")
	} else if len(config.Subscriptions) == 0 {
		return sensu.CheckStateWarning, errors.New("--subscriptions flag or $SENSU_RUNBOOK_SUBSCRIPTIONS environment variable must be set")
	}
	return sensu.CheckStateOK, nil
}

func executePlaybook(event *types.Event) (int, error) {
	// TODO: use the sensu-plugin-sdk HTTP client (reference: https://github.com/sensu/sensu-ec2-handler/blob/master/main.go#L12)
	job, err := generateCheckConfig()
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("ERROR: %s", err)
	}
	log.Printf("registering runbook job ID %s/%s with --command %s\n", job.Namespace, job.Name, config.Command)
	err = createJob(&job)
	if err != nil {
		return sensu.CheckStateCritical, err
	}
	err = executeJob(&job)
	if err != nil {
		return sensu.CheckStateCritical, nil
	}
	return sensu.CheckStateOK, nil
}

func generateCheckConfig() (types.CheckConfig, error) {
	// Build CheckConfig object
	var timeout, _ = strconv.Atoi(config.Timeout)
	var labels = make(map[string]string)
	var job = types.CheckConfig{
		ObjectMeta: types.ObjectMeta{
			Name:      config.JobID,
			Namespace: config.Namespace,
			Labels:    labels,
		},
		Command:       config.Command,
		Publish:       false,
		Subscriptions: []string{"none"},
		Interval:      10,
		Timeout:       uint32(timeout),
	}
	if len(config.RuntimeAssets) > 0 {
		job.RuntimeAssets = strings.Split(config.RuntimeAssets, ",")
	}
	return job, nil
}

// LoadCACerts loads the system cert pool.
func LoadCACerts(path string) (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("ERROR: failed to load system cert pool: %s", err)
		rootCAs = x509.NewCertPool()
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if path != "" {
		certs, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("ERROR: failed to read CA file (%s): %s", path, err)
			return nil, err
		}
		rootCAs.AppendCertsFromPEM(certs)
	}
	return rootCAs, nil
}

func initHTTPClient() *http.Client {
	certs, err := LoadCACerts(config.SensuTrustedCaFile)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
	tlsConfig := &tls.Config{
		RootCAs: certs,
	}
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{
		Transport: tr,
	}
	return client
}

func createJob(job *types.CheckConfig) error {
	postBody, err := json.Marshal(job)
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
	body := bytes.NewReader(postBody)
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/core/v2/namespaces/%s/checks",
			config.SensuAPIUrl,
			config.Namespace,
		),
		body,
	)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
	var httpClient *http.Client = initHTTPClient()
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SensuAccessToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
		return err
	} else if resp.StatusCode == 404 {
		log.Fatalf("ERROR: %v %s (%s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), req.URL)
		return err
	} else if resp.StatusCode == 409 {
		log.Printf("runbook job \"%s\" already exists (%v: %s)\n", job.Name, resp.StatusCode, http.StatusText(resp.StatusCode))
		return err
	} else if resp.StatusCode >= 300 {
		log.Fatalf("ERROR: %v %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return err
	} else if resp.StatusCode == 201 {
		log.Printf("registered runbook Job \"%s\"", job.Name)
		return nil
	} else {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err)
		} else {
			fmt.Printf("%s\n", string(b))
		}
	}

	return err
}

func executeJob(job *types.CheckConfig) error {
	var jobRequest = JobRequest{
		Check:         job.Name,
		Subscriptions: strings.Split(config.Subscriptions, ","),
	}
	postBody, err := json.Marshal(jobRequest)
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
	body := bytes.NewReader(postBody)
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/core/v2/namespaces/%s/checks/%s/execute",
			config.SensuAPIUrl,
			config.Namespace,
			config.JobID,
		),
		body,
	)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
	var httpClient *http.Client = initHTTPClient()
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SensuAccessToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
		return err
	} else if resp.StatusCode == 404 {
		log.Fatalf("ERROR: %v %s (%s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), req.URL)
		return err
	} else if resp.StatusCode >= 300 {
		log.Fatalf("ERROR: %v %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return err
	} else if resp.StatusCode == 202 {
		log.Printf("requested runbook Job \"%s\" execution on subscriptions: %s\n", job.Name, config.Subscriptions)
		return nil
	} else {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err)
			return err
		}
		fmt.Printf("%s\n", string(b))
		return nil
	}
}
