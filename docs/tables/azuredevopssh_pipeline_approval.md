---
title: "Steampipe Table: azuredevopssh_pipeline_approval - Query Azure DevOps Pipeline Approvals using SQL"
description: "Allows users to query YAML pipeline approvals in an Azure DevOps project."
---

# Table: azuredevopssh_pipeline_approval - Query Azure DevOps Pipeline Approvals using SQL

A pipeline approval is an approval check step created when a YAML pipeline reaches an approval. The `azuredevopssh_pipeline_approval` table reads one approval at a time.

`project_id` and `id` are required. Azure DevOps Server 2022 rejects an approvals query that does not include an approval id. That id is the timeline record id of type `Checkpoint.Approval` on the build that is waiting. Classic release approvals are in `azuredevopssh_release_approval`.

## Table Usage Guide

Use this table to read the status, instructions, and steps of a known YAML pipeline approval. An id that does not exist returns no row.

## Examples

### One approval

```sql+postgres
select
  id,
  status,
  instructions,
  min_required_approvers,
  steps
from
  azuredevopssh_pipeline_approval
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and id = '00000000-0000-0000-0000-000000000000';
```

```sql+sqlite
select
  id,
  status,
  instructions,
  min_required_approvers,
  steps
from
  azuredevopssh_pipeline_approval
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and id = '00000000-0000-0000-0000-000000000000';
```
