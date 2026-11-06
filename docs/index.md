---
organization: Turbot
category: ["saas"]
icon_url: "/images/plugins/turbot/azuredevops.svg"
brand_color: "#0090F1"
display_name: "Azure DevOps"
short_name: "azuredevops"
description: "Steampipe plugin to query projects, groups, builds and more from Azure DevOps."
og_description: "Query Azure DevOps with SQL! Open source CLI. No DB required."
og_image: "/images/plugins/turbot/azuredevops-social-graphic.png"
engines: ["steampipe", "sqlite", "postgres", "export"]
---

# Azure DevOps + Steampipe

[Steampipe](https://steampipe.io) is an open-source zero-ETL engine to instantly query cloud APIs using SQL.

[Azure DevOps](https://dev.azure.com) is a software as a service (SaaS) platform that provides DevOps practices and tools for the end-to-end software life cycle.

List your Azure DevOps projects:

```sql
select
  id,
  name,
  state,
  visibility
from
  azuredevopssh_project;
```

```
+--------------------------------------+-----------------+------------+------------+
| id                                   | name            | state      | visibility |
+--------------------------------------+-----------------+------------+------------+
| bdcdf70b-7757-4253-b36c-33c08ca07dbb | test-1          | wellFormed | private    |
| 9a1f26ce-c715-4ef4-b557-503fdb6be55a | private_project | wellFormed | private    |
+--------------------------------------+-----------------+------------+------------+
```

## Documentation

- **[Table definitions & examples →](/plugins/turbot/azuredevops/tables)**

## Quick start

### Install

Download and install the latest Azure DevOps plugin:

```sh
steampipe plugin install azuredevops
```

### Credentials

| Item        | Description                                                                                                                                                                                                                                                                                                                                            |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Credentials | Azure DevOps requires an [Organization URL](https://learn.microsoft.com/en-us/azure/devops/extend/develop/work-with-urls?view=azure-devops&tabs=http) (SaaS) or a Collection URL (self-hosted) and a [Personal Access Token](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows) for all requests. |
| Permissions | Personal Access Tokens have the same permissions as the user who creates them, and if the user permissions change, the Personal Access Token permissions also change.                                                                                                                                                                                  |
| Radius      | Each connection represents a single Azure DevOps organization (SaaS) or collection (self-hosted).                                                                                                                                                                                                                                                       |
| Resolution  | 1. Credentials explicitly set in a steampipe config file (`~/.steampipe/config/azuredevops.spc`)<br />2. Credentials specified in environment variables, e.g., `AZDO_ORG_SERVICE_URL` and `AZDO_PERSONAL_ACCESS_TOKEN`.                                                                                                                                |

### Configuration

Installing the latest azuredevops plugin will create a config file (`~/.steampipe/config/azuredevops.spc`) with a single connection named `azuredevops`:

Configure your account details in `~/.steampipe/config/azuredevops.spc`:

```hcl
connection "azuredevops" {
  plugin = "azuredevops"

  # `organization_url`: Azure DevOps Organization URL (SaaS) or Collection URL (self-hosted). (Required)
  # For SaaS, this is the organization URL, e.g. https://dev.azure.com/test
  # For self-hosted Azure DevOps Server, this is a specific collection URL, e.g. https://code.digiturk.net/SCMPOCCollection
  # Can also be set with the AZDO_ORG_SERVICE_URL environment variable.
  # organization_url = "https://dev.azure.com/test"

  # `server_url`: Azure DevOps Server root URL (self-hosted only, optional).
  # Used only by the `azuredevopssh_project_collection` table to enumerate collections.
  # When unset, it is derived from `organization_url` by stripping the last path segment.
  # server_url = "https://code.digiturk.net/"

  # `personal_access_token`: Azure DevOps Personal Access Token. (Required)
  # For more information on the Personal Access Token, please see https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows.
  # Can also be set with the AZDO_PERSONAL_ACCESS_TOKEN environment variable.
  # personal_access_token = "wf3hahidy7i7fkzmeqr3e6fbjwuspabpo766grp7hl4o65v2"
}
```

Alternatively, you can also use the standard Azure DevOps environment variables to obtain credentials **only if other arguments (`organization_url` and `personal_access_token`) are not specified** in the connection:

```sh
export AZDO_ORG_SERVICE_URL=https://dev.azure.com/test
export AZDO_PERSONAL_ACCESS_TOKEN=wf3hahidy7i7fkzmeqr3e6fbjwuspabpo766grp7hl4o65v2
```

### Self-hosted Azure DevOps Server

For self-hosted Azure DevOps Server (e.g. 2022.1), the `organization_url` points at a specific **collection** rather than an organization:

```sh
export AZDO_ORG_SERVICE_URL=https://code.digiturk.net/SCMPOCCollection
export AZDO_PERSONAL_ACCESS_TOKEN=...
```

To enumerate all collections on the server, use the `azuredevopssh_project_collection` table. It uses the `server_url` (or derives it from `organization_url` by stripping the collection path segment):

```sql
select
  id,
  name,
  url
from
  azuredevopssh_project_collection;
```
