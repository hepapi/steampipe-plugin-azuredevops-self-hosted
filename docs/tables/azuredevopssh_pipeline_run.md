---
title: "Steampipe Table: azuredevopssh_pipeline_run - Query Azure DevOps Pipeline Runs using SQL"
description: "Allows users to query runs of a single Azure DevOps YAML pipeline."
---

# Table: azuredevopssh_pipeline_run - Query Azure DevOps Pipeline Runs using SQL

A pipeline run is one execution of a YAML pipeline. The `azuredevopssh_pipeline_run` table lists runs for one pipeline, including state, result, and timestamps.

`project_id` and `pipeline_id` are required. `ListRuns` can return a large history for that pipeline, so add a SQL `limit` when you only need the latest rows. `id`, `pipeline_id`, and `project_id` together fetch one run.

## Table Usage Guide

Use this table for YAML pipeline history. Classic build history stays in `azuredevopssh_build`. This table does not select the expanded final YAML.

## Examples

### Recent runs of one pipeline

```sql+postgres
select
  id,
  name,
  state,
  result,
  created_date,
  finished_date
from
  azuredevopssh_pipeline_run
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and pipeline_id = 12
limit 20;
```

```sql+sqlite
select
  id,
  name,
  state,
  result,
  created_date,
  finished_date
from
  azuredevopssh_pipeline_run
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and pipeline_id = 12
limit 20;
```

### One run

```sql+postgres
select
  id,
  name,
  state,
  result,
  url
from
  azuredevopssh_pipeline_run
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and pipeline_id = 12
  and id = 340;
```

```sql+sqlite
select
  id,
  name,
  state,
  result,
  url
from
  azuredevopssh_pipeline_run
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and pipeline_id = 12
  and id = 340;
```
