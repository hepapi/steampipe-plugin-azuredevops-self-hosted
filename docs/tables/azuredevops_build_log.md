---
title: "Steampipe Table: azuredevops_build_log - Query Azure DevOps Build Log Metadata using SQL"
description: "Allows users to query metadata for the logs of a single Azure DevOps build, including the log URL and line count."
---

# Table: azuredevops_build_log - Query Azure DevOps Build Log Metadata using SQL

Each build keeps one or more logs. The `azuredevops_build_log` table lists that metadata: id, type, URL, line count, and timestamps.

`project_id` and `build_id` are required. The table does not download log lines or the log zip.

## Table Usage Guide

Use this table to find the log URL for a build, or to see how large each log is, before opening the log in Azure DevOps.

## Examples

### Logs for one build

```sql+postgres
select
  id,
  type,
  line_count,
  created_on,
  url
from
  azuredevops_build_log
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123;
```

```sql+sqlite
select
  id,
  type,
  line_count,
  created_on,
  url
from
  azuredevops_build_log
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123;
```

### One log by id

```sql+postgres
select
  id,
  line_count,
  last_changed_on,
  url
from
  azuredevops_build_log
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123
  and id = 1;
```

```sql+sqlite
select
  id,
  line_count,
  last_changed_on,
  url
from
  azuredevops_build_log
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123
  and id = 1;
```
