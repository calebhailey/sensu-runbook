package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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
	Command string
	Timeout string
	Assets string 
	Subscriptions string 
	Namespace string
	JobID string 
	SensuApiUrl string
	SensuAccessToken string 
	SensuTrustedCaFile string
}

type JobRequest struct {
	Check string `json:"check"`
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
		&sensu.PluginConfigOption{
			Path:      "command",
			Env:       "SENSU_RUNBOOK_COMMAND",
			Argument:  "command",
			Shorthand: "c",
			Default:   "",
			Usage:     "The command that should be executed by the Sensu Go agent(s)",
			Value:     &config.Command,
		},
		&sensu.PluginConfigOption{
			Path:      "timeout",
			Env:       "SENSU_RUNBOOK_TIMEOUT",
			Argument:  "timeout",
			Shorthand: "t",
			Default:   "10",
			Usage:     "Command execution timeout, in seconds",
			Value:     &config.Command,
		},
		&sensu.PluginConfigOption{
			Path:      "assets",
			Env:       "SENSU_RUNBOOK_ASSETS",
			Argument:  "assets",
			Shorthand: "a",
			Default:   "",
			Usage:     "Comma-separated list of assets to distribute with the command(s)",
			Value:     &config.Assets,
		},
		&sensu.PluginConfigOption{
			Path:      "subscriptions",
			Env:       "SENSU_RUNBOOK_SUBSCRIPTIONS",
			Argument:  "subscriptions",
			Shorthand: "s",
			Default:   "",
			Usage:     "Comma-separated list of subscriptions to execute the command(s) on",
			Value:     &config.Subscriptions,
		},
		&sensu.PluginConfigOption{
			Path:      "namespace",
			Env:       "SENSU_NAMESPACE", // provided by the sensuctl command plugin execution environment
			Argument:  "namespace",
			Shorthand: "n",
			Default:   "",
			Usage:     "Sensu Namespace to perform the runbook automation (defaults to $SENSU_NAMESPACE)",
			Value:     &config.Namespace,
		},
		&sensu.PluginConfigOption{
			Path:      "sensu-api-url",
			Env:       "SENSU_API_URL", // provided by the sensuctl command plugin execution environment
			Argument:  "sensu-api-url",
			Shorthand: "",
			Default:   "",
			Usage:     "Sensu API URL (defaults to $SENSU_API_URL)",
			Value:     &config.SensuApiUrl,
		},
		&sensu.PluginConfigOption{
			Path:      "sensu-access-token",
			Env:       "SENSU_ACCESS_TOKEN", // provided by the sensuctl command plugin execution environment
			Argument:  "sensu-access-token",
			Shorthand: "",
			Default:   "",
			Usage:     "Sensu API Access Token (defaults to $SENSU_ACCESS_TOKEN)",
			Value:     &config.SensuAccessToken,
		},
		&sensu.PluginConfigOption{
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
	if len(config.Command) == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--command or SENSU_RUNBOOK_COMMAND environment variable is required")
	}
	if len(config.Subscriptions) == 0 {
		return sensu.CheckStateWarning, fmt.Errorf("--subscriptions or SENSU_RUNBOOK_SUBSCRIPTIONS environment variable is required")
	}
	return sensu.CheckStateOK, nil
}

func executePlaybook(event *types.Event) (int, error) {
	// TODO: use the sensu-plugin-sdk HTTP client (reference: https://github.com/sensu/sensu-ec2-handler/blob/master/main.go#L12)
	job, err := generateCheckConfig()
	if err != nil {
		fmt.Errorf("ERROR: %s\n", err)
		return sensu.CheckStateCritical, err
	} else {
		log.Printf("registering runbook job ID %s/%s with --command %s", job.Namespace, job.Name, config.Command)
		err = createJob(&job)
		if err != nil {
			return sensu.CheckStateCritical, err
		} else {
			err = executeJob(&job)
		}
		if err != nil {
			return sensu.CheckStateCritical, nil
		} else {
			return sensu.CheckStateOK, nil
		}
	}
}

func generateCheckConfig() (types.CheckConfig, error) {
	// Build CheckConfig object 
	timeout, _ := strconv.Atoi(config.Timeout)
	labels := make(map[string]string)
	config.JobID = uuid.New().String()
	var job = types.CheckConfig{
		Publish: false,
		Subscriptions: []string{"none"},
		Command: config.Command,
		Interval: 10,
		Timeout: uint32(timeout),
		ObjectMeta: types.ObjectMeta{
			Name: config.JobID,
			Namespace: config.Namespace,
			Labels: labels,
		},
	}
	job.ObjectMeta.Labels["check_type"] = "runbook"
	job.ObjectMeta.Labels["source"] = "Sensu Runbook"
	return job, nil
}

func LoadCACerts(path string) (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("ERROR: failed to load system cert pool: %s", err)
		return nil, err
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if path != "" {
		certs, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("ERROR: failed to read CA file (%s): %s", path, err)
			return nil, err
		} else {
			rootCAs.AppendCertsFromPEM(certs)
		}
	}
	return rootCAs, nil
}

func initHttpClient() *http.Client {
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
			config.SensuApiUrl,
			config.Namespace,
		),
		body,
	)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
	var httpClient *http.Client = initHttpClient()
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
	var job_request = JobRequest{
		Check: job.Name,
		Subscriptions: strings.Split(config.Subscriptions, ","),
	}
	postBody, err := json.Marshal(job_request)
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
	body := bytes.NewReader(postBody)
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/core/v2/namespaces/%s/checks/%s/execute",
			config.SensuApiUrl,
			config.Namespace,
			config.JobID,
		),
		body,
	)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
	var httpClient *http.Client = initHttpClient()
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
		} else {
			fmt.Printf("%s\n", string(b))
			return nil
		}
	}

	return err
}

