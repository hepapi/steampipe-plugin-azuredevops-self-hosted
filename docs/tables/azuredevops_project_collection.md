---
title: "Steampipe Table: azuredevops_project_collection - Query Azure DevOps Project Collections using SQL"
description: "Allows users to query project collections in a self-hosted Azure DevOps Server, providing insights into the collections that group team projects."
---

# Table: azuredevops_project_collection - Query Azure DevOps Project Collections using SQL

Azure DevOps Server (self-hosted) organizes team projects into **project collections**. The `azuredevops_project_collection` table provides insights into these collections, including their name, URL, description, and state.

This table is primarily useful for self-hosted deployments. It enumerates collections at the server root (`server_url`), which is derived from `organization_url` by stripping the collection path segment (or set explicitly via the `server_url` config).

## Table Usage Guide

The `azuredevops_project_collection` table provides insights into project collections within a self-hosted Azure DevOps Server. As an administrator, explore collection-specific details through this table, including the collection name, URL, and state. Utilize it to discover all collections on a server and understand the deployment topology.

## Examples

### Basic info
List all project collections on the server.

```sql+postgres
select
  id,
  name,
  url,
  state
from
  azuredevops_project_collection;
```

```sql+sqlite
select
  id,
  name,
  url,
  state
from
  azuredevops_project_collection;
```

### Get a specific collection
Retrieve details for a collection by its ID.

```sql+postgres
select
  id,
  name,
  description,
  state,
  process_customization_type
from
  azuredevops_project_collection
where
  id = '00000000-0000-0000-0000-000000000000';
```

```sql+sqlite
select
  id,
  name,
  description,
  state,
  process_customization_type
from
  azuredevops_project_collection
where
  id = '00000000-0000-0000-0000-000000000000';
```
