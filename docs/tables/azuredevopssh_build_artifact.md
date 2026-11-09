---
title: "Steampipe Table: azuredevopssh_build_artifact - Query Azure DevOps Build Artifacts using SQL"
description: "Allows users to query artifacts published by a single Azure DevOps build."
---

# Table: azuredevopssh_build_artifact - Query Azure DevOps Build Artifacts using SQL

A build artifact is a file or directory a build publishes for later jobs or releases. The `azuredevopssh_build_artifact` table lists the artifacts of one build, including the download resource.

`project_id` and `build_id` are required.

## Table Usage Guide

Use this table to see what a build published and where the artifact can be downloaded. The `resource` column holds the artifact type and download URL.

## Examples

### Artifacts for one build

```sql
select
  id,
  name,
  source,
  resource
from
  azuredevopssh_build_artifact
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123;
```

### One artifact by name

```sql
select
  id,
  name,
  resource
from
  azuredevopssh_build_artifact
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and build_id = 123
  and name = 'drop';
```

