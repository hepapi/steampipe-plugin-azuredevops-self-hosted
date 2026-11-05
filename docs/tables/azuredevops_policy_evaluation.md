---
title: "Steampipe Table: azuredevops_policy_evaluation - Query Azure DevOps Policy Evaluations using SQL"
description: "Allows users to query Azure DevOps policy evaluation records for a pull request or other policy artifact, including status, policy type, and whether the policy is blocking."
---

# Table: azuredevops_policy_evaluation - Query Azure DevOps Policy Evaluations using SQL

Azure DevOps evaluates each applicable branch policy against a pull request and records the result. The `azuredevops_policy_evaluation` table returns those records: one row per policy running against one artifact. Use it to see which policies approved, rejected, or failed a pull request.

`project_id` and `artifact_id` are required. The artifact id for a pull request is `vstfs:///CodeReview/CodeReviewId/{projectId}/{pullRequestId}`, and `azuredevops_git_pull_request.artifact_id` already has that value. The list call includes policies that do not apply (`status = 'notApplicable'`) so a status filter can see them.

## Table Usage Guide

The `azuredevops_policy_evaluation` table provides the current evaluation state of policies on a specific artifact. As a DevOps engineer, use it to find rejected or broken checks on a pull request and to see which policy configuration produced each result.

## Examples

### Basic info
List every policy evaluation for one pull request artifact.

```sql+postgres
select
  id,
  status,
  policy_type_display_name,
  is_enabled,
  is_blocking,
  started_date,
  completed_date
from
  azuredevops_policy_evaluation
where
  project_id = 'bdcdf70b-7757-4253-b36c-33c08ca07dbb'
  and artifact_id = 'vstfs:///CodeReview/CodeReviewId/bdcdf70b-7757-4253-b36c-33c08ca07dbb/42';
```

```sql+sqlite
select
  id,
  status,
  policy_type_display_name,
  is_enabled,
  is_blocking,
  started_date,
  completed_date
from
  azuredevops_policy_evaluation
where
  project_id = 'bdcdf70b-7757-4253-b36c-33c08ca07dbb'
  and artifact_id = 'vstfs:///CodeReview/CodeReviewId/bdcdf70b-7757-4253-b36c-33c08ca07dbb/42';
```

### List rejected or broken evaluations
Find evaluations that are blocking the artifact.

```sql+postgres
select
  id,
  status,
  policy_type_display_name,
  configuration_id
from
  azuredevops_policy_evaluation
where
  project_id = 'bdcdf70b-7757-4253-b36c-33c08ca07dbb'
  and artifact_id = 'vstfs:///CodeReview/CodeReviewId/bdcdf70b-7757-4253-b36c-33c08ca07dbb/42'
  and status in ('rejected', 'broken');
```

```sql+sqlite
select
  id,
  status,
  policy_type_display_name,
  configuration_id
from
  azuredevops_policy_evaluation
where
  project_id = 'bdcdf70b-7757-4253-b36c-33c08ca07dbb'
  and artifact_id = 'vstfs:///CodeReview/CodeReviewId/bdcdf70b-7757-4253-b36c-33c08ca07dbb/42'
  and status in ('rejected', 'broken');
```

### Join evaluations to a pull request
Use the pull request's `artifact_id` instead of building the artifact string yourself.

```sql+postgres
select
  pr.pull_request_id,
  pr.title as pull_request_title,
  e.status,
  e.policy_type_display_name,
  e.is_blocking
from
  azuredevops_git_pull_request as pr
  join azuredevops_policy_evaluation as e
    on e.project_id = pr.project_id
    and e.artifact_id = pr.artifact_id
where
  pr.pull_request_id = 42
  and pr.repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21';
```

```sql+sqlite
select
  pr.pull_request_id,
  pr.title as pull_request_title,
  e.status,
  e.policy_type_display_name,
  e.is_blocking
from
  azuredevops_git_pull_request as pr
  join azuredevops_policy_evaluation as e
    on e.project_id = pr.project_id
    and e.artifact_id = pr.artifact_id
where
  pr.pull_request_id = 42
  and pr.repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21';
```
