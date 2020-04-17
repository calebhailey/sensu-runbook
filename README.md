[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/calebhailey/sensu-runbook)
![Go Test](https://github.com/calebhailey/sensu-runbook/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/calebhailey/sensu-runbook/workflows/goreleaser/badge.svg)

# Sensu Runbook

## Table of Contents

- [Overview](#overview)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Sensuctl command installation](#sensuctl-command-installation)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

Execute commands and custom scripts on Sensu Agent nodes. Sensu Go offers several unique attributes that make it a compelling runbook automation platform, including:

- **Namespaces & RBAC**. Sensu Go is a multi-tenant platform, with SSO and API
  Keys, enabling _self-service_ runbook automation 
- **Secure transport**. Sensu offers TLS certificate-based auth & TLS encrypted 
  transport
- **Publish/subscribe model of communication**. No centralized system with "keys to 
  the kingdom" to remotely access target nodes and execute arbitrary commands
- [Sensu Assets][assets]. Use Sensu to securely automate distribution of custom 
  runbook code (e.g. shell scripts or python scripts)
- [Sensu Tokens][tokens]. Use Sensu's powerful templating system to create dynamic
  runbook automations that can scale to thousands of nodes
- **Event processing**. Integrate with Slack, Pagerduty, ServiceNow, VictorOps, 
  JIRA, and more
- **Code reuse**. Reuse ad hoc runbook scripts as monitoring checks or automated 
  remediations using Sensu Go
- **Audit logging**. The runbook configuration, request, execution, and event 
  processing are all logged by Sensu
- Much, much, more...

[assets]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[tokens]: https://docs.sensu.io/sensu-go/latest/reference/tokens/

## Usage examples

* Run as a standalone program: 

  ```
  $ ./sensu-runbook --help
  Sensu Runbook Automation. Execute commands on Sensu Agent nodes.

  Usage:
    sensu-runbook [flags]
    sensu-runbook [command]

  Available Commands:
    help        Help about any command
    version     Print the version number of this plugin

  Flags:
    -a, --assets string                  Comma-separated list of assets to distribute with the command(s)
    -c, --command string                 The command that should be executed by the Sensu Go agent(s)
    -h, --help                           help for sensu-runbook
    -n, --namespace string               Sensu Namespace to perform the runbook automation (defaults to $SENSU_NAMESPACE) (default "sensu-system")
        --sensu-access-token string      Sensu API Access Token (defaults to $SENSU_ACCESS_TOKEN)
        --sensu-api-url string           Sensu API URL (defaults to $SENSU_API_URL) (default "https://demo.sensu.io:8080")
        --sensu-trusted-ca-file string   Sensu API Trusted Certificate Authority File (defaults to $SENSU_TRUSTED_CA_FILE)
    -s, --subscriptions string           Comma-separated list of subscriptions to execute the command(s) on
    -t, --timeout string                 Command execution timeout, in seconds (default "10")

  Use "sensu-runbook [command] --help" for more information about a command.
  ```

* Use as a [`sensuctl` command plugin][plugin]: 

  ```
  $ sensuctl command install runbook calebhailey/sensu-runbook
  command was installed successfully
  $ sensuctl command exec runbook -- --help 
  INFO[0000] asset includes builds, using builds instead of asset  asset=runbook component=asset-manager entity=sensuctl
  Sensu Runbook Automation. Execute commands on Sensu Agent nodes.

  Usage:
    sensu-runbook [flags]
    sensu-runbook [command]

  Available Commands:
    help        Help about any command
    version     Print the version number of this plugin

  Flags:
    -a, --assets string                  Comma-separated list of assets to distribute with the command(s)
    -c, --command string                 The command that should be executed by the Sensu Go agent(s)
    -h, --help                           help for sensu-runbook
    -n, --namespace string               Sensu Namespace to perform the runbook automation (defaults to $SENSU_NAMESPACE) (default "sensu-system")
        --sensu-access-token string      Sensu API Access Token (defaults to $SENSU_ACCESS_TOKEN)
        --sensu-api-url string           Sensu API URL (defaults to $SENSU_API_URL) (default "https://demo.sensu.io:8080")
        --sensu-trusted-ca-file string   Sensu API Trusted Certificate Authority File (defaults to $SENSU_TRUSTED_CA_FILE)
    -s, --subscriptions string           Comma-separated list of subscriptions to execute the command(s) on
    -t, --timeout string                 Command execution timeout, in seconds (default "10")

  Use "sensu-runbook [command] --help" for more information about a command.
  ```

  [plugin]: https://docs.sensu.io/sensu-go/latest/sensuctl/reference/#extend-sensuctl-with-commands

## Configuration

### Sensuctl command installation

This plugin is intended for use as a [`sensuctl` command plugin][command]. If 
you're using sensuctl 5.13 or later, run the following command: 

```
sensuctl command install calebhailey/sensu-runbook
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-runbook repository:

```
go build
```

## Additional notes

This plugin is in technical preview and should be considered "unstable", but 
feedback is welcome and appreciated! 

### Roadmap 

- Publish asset to Bonsai
- Add support for `--runtime_assets`
- Add support for `--handlers`
- Add support for Assets (`--url` and `--sha512`)
- Write a blog post!

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu-community/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/check-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/check-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[7]: https://github.com/sensu-community/check-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-community/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
