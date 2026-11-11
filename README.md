# Azure DevOps Server plugin for Steampipe

Query a self-hosted Azure DevOps Server with SQL. This is Hepapi's unofficial plugin for Azure DevOps Server 2022.1. It started from [Turbot's azuredevops plugin](https://github.com/turbot/steampipe-plugin-azuredevops) and adds collection, Git, policy, build, release, work item, test, identity, artifact, wiki, search, audit, extension, and agent tables.

SQL tables use the `azuredevopssh_` prefix, so this plugin can be installed beside Turbot's `azuredevops` plugin. The config file and connection name remain `azuredevops`.

- Documentation: [docs/index.md](docs/index.md)
- Issues: [hepapi/steampipe-plugin-azuredevops-self-hosted](https://github.com/hepapi/steampipe-plugin-azuredevops-self-hosted/issues)

## Quick start

Install the plugin:

```bash
steampipe plugin install ghcr.io/hepapi/azuredevopssh
```

The first install creates `~/.steampipe/config/azuredevops.spc` with a connection named `azuredevops`. Set a personal access token and either one collection URL or the server root:

```hcl
connection "azuredevops" {
  plugin = "ghcr.io/hepapi/azuredevopssh"

  # One collection. Collection-scoped tables read only this URL.
  # organization_url = "https://azuredevops.example.com/DefaultCollection"

  # Server root. Collection-scoped tables run once per collection the token can list.
  # organization_url = "https://azuredevops.example.com"
  # Skip TFVC-only collections. New collections are included until you name them here.
  # excluded_collections = ["LegacyTfs"]
  # Or query only these collections.
  # collections = ["DefaultCollection", "AnotherCollection"]

  # Server root. Optional. Used to list collections and to read server-scoped tables
  # (project collections, agent pools, and agents) when organization_url is not one collection.
  # When unset, it is derived by stripping the last path segment of organization_url.
  # server_url = "https://azuredevops.example.com/"

  # personal_access_token = "..."
}
```

The same two values can come from the environment when the connection leaves them unset:

```sh
export AZDO_ORG_SERVICE_URL=https://azuredevops.example.com/DefaultCollection
export AZDO_PERSONAL_ACCESS_TOKEN=...
```

An Azure DevOps Services organization URL (`https://dev.azure.com/{organization}`) stays one organization because it has a path. `server_url` is the server root used to list project collections. On a server-root connection it is also the URL for server-scoped tables: `azuredevopssh_project_collection`, `azuredevopssh_agent_pool`, and `azuredevopssh_agent`. Those tables run once and do not take a collection qualifier. `where organization = 'SomeCollection'` narrows collection-scoped tables only.

Run a query:

```sql
select
  id,
  name,
  state,
  visibility
from
  azuredevopssh_project;
```

Table definitions, server caveats, and more examples are in [docs/index.md](docs/index.md).

## Engines

| Engine | Description |
| --- | --- |
| [Steampipe](https://steampipe.io/docs) | Install the plugin into the Steampipe CLI and query it with SQL. |

## Developing

Prerequisites: [Steampipe](https://steampipe.io/downloads) and [Go](https://go.dev/doc/install).

```sh
git clone https://github.com/hepapi/steampipe-plugin-azuredevops-self-hosted.git
cd steampipe-plugin-azuredevops-self-hosted
make
cp config/* ~/.steampipe/config
```

`make` builds the plugin into `~/.steampipe/plugins`. Then:

```
steampipe query
> .inspect azuredevops
```

Further reading: [Writing plugins](https://steampipe.io/docs/develop/writing-plugins).

## License

Source code is [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0). Documentation is [CC BY-NC-ND 4.0](docs/LICENSE).
