---
title: "Steampipe Table: azuredevops_pipeline_check - Query Azure DevOps Pipeline Checks using SQL"
description: "Allows users to query check configurations on one Azure DevOps pipeline resource."
---

# Table: azuredevops_pipeline_check - Query Azure DevOps Pipeline Checks using SQL

A pipeline check is a gate configured on a resource such as a service connection, environment, queue, variable group, secure file, or repository. The `azuredevops_pipeline_check` table lists the checks on one resource.

`project_id`, `resource_type`, and `resource_id` are required. The checks API does not offer a collection-wide inventory. `resource_type` is the value Azure DevOps expects, for example `endpoint`, `environment`, `queue`, `variablegroup`, `securefile`, or `repository`.

The typed client returns the check id, type, resource, timeout, and audit fields. Check-specific settings are not on that model.

## Table Usage Guide

Use this table to see which checks protect a service connection or environment. Resolve `resource_id` from the resource table first, for example `azuredevops_serviceendpoint.id` with `resource_type = 'endpoint'`.

## Examples

### Checks on a service connection

```sql+postgres
select
  id,
  type_name,
  resource_name,
  timeout,
  created_on
from
  azuredevops_pipeline_check
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and resource_type = 'endpoint'
  and resource_id = '00000000-0000-0000-0000-000000000000';
```

```sql+sqlite
select
  id,
  type_name,
  resource_name,
  timeout,
  created_on
from
  azuredevops_pipeline_check
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and resource_type = 'endpoint'
  and resource_id = '00000000-0000-0000-0000-000000000000';
```
