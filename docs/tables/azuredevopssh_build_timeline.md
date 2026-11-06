---
title: "Steampipe Table: azuredevopssh_build_timeline - Query Azure DevOps Build Timeline Records using SQL"
description: "Allows users to query the timeline records of a single Azure DevOps build, including stage, job, and task state."
---

# Table: azuredevopssh_build_timeline - Query Azure DevOps Build Timeline Records using SQL

A build timeline is the tree of stages, jobs, and tasks that ran for one build. The `azuredevopssh_build_timeline` table returns one row per timeline record.

`project_id` and `build_id` are required. The table calls `GetBuildTimeline` for that build and does not scan the build history.

## Table Usage Guide

Use this table to see which steps ran, how long they took, and which ones failed. Join `id` to `azuredevopssh_build_log` through the `log` reference when you need the log id.

## Examples

### Timeline for one build
List the records of a build.

```sql+postgres
select
  id,
  name,
  type,
  state,
  result,
  start_time,
  finish_time,
  error_count
from
  azuredevopssh_build_timeline
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123;
```

```sql+sqlite
select
  id,
  name,
  type,
  state,
  result,
  start_time,
  finish_time,
  error_count
from
  azuredevopssh_build_timeline
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123;
```

### Failed tasks
Find completed records that did not succeed.

```sql+postgres
select
  name,
  type,
  result,
  error_count,
  warning_count
from
  azuredevopssh_build_timeline
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123
  and result is not null
  and result <> 'succeeded';
```

```sql+sqlite
select
  name,
  type,
  result,
  error_count,
  warning_count
from
  azuredevopssh_build_timeline
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123
  and result is not null
  and result <> 'succeeded';
```
