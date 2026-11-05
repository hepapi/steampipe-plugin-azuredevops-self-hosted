---
title: "Steampipe Table: azuredevops_release_approval - Query Azure DevOps Release Approvals using SQL"
description: "Allows users to query pre-deploy and post-deploy approvals on Azure DevOps classic releases."
---

# Table: azuredevops_release_approval - Query Azure DevOps Release Approvals using SQL

A release approval is a manual or automated gate on a classic release environment. The `azuredevops_release_approval` table lists those approvals, including who should approve and who did.

When `status` is omitted, the table requests pending, approved, rejected, canceled, skipped, and reassigned approvals. The Azure DevOps API returns only pending approvals if the status filter is left unset, so the table sets the filter itself. Pass `status` to read one state. Pass `project_id` and `release_id` to avoid scanning every project.

## Table Usage Guide

Use this table to find approvals that are still waiting, and to see the approver recorded on a release. `approval_type` is `preDeploy` or `postDeploy`.

## Examples

### Pending approvals

```sql+postgres
select
  project_id,
  id,
  release_id,
  release_environment_id,
  approval_type,
  approver
from
  azuredevops_release_approval
where
  status = 'pending'
limit 20;
```

```sql+sqlite
select
  project_id,
  id,
  release_id,
  release_environment_id,
  approval_type,
  approver
from
  azuredevops_release_approval
where
  status = 'pending'
limit 20;
```

### Approvals for one release

```sql+postgres
select
  id,
  status,
  approval_type,
  comments,
  modified_on
from
  azuredevops_release_approval
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and release_id = 42;
```

```sql+sqlite
select
  id,
  status,
  approval_type,
  comments,
  modified_on
from
  azuredevops_release_approval
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and release_id = 42;
```
