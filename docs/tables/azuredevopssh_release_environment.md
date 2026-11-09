---
title: "Steampipe Table: azuredevopssh_release_environment - Query Azure DevOps Release Environments using SQL"
description: "Allows users to query the environments of Azure DevOps classic releases, including deployment status."
---

# Table: azuredevopssh_release_environment - Query Azure DevOps Release Environments using SQL

A release environment is one stage of a classic release. The `azuredevopssh_release_environment` table returns one row per environment on each release, including status, rank, and time to deploy.

An unscoped query lists releases in every project. Pass `project_id`, and `release_id` when you already know the release.

## Table Usage Guide

Use this table to see which stages succeeded, failed, or are still running. `definition_environment_id` is the stage id on the release definition. `id` is the id of that stage on this release.

## Examples

### Recent release stages

```sql
select
  project_id,
  release_id,
  id,
  name,
  status,
  rank,
  time_to_deploy
from
  azuredevopssh_release_environment
limit 20;
```

### Stages of one release

```sql
select
  name,
  status,
  rank,
  created_on,
  modified_on
from
  azuredevopssh_release_environment
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and release_id = 42
order by
  rank;
```

