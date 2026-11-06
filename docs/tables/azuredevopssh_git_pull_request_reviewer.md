---
title: "Steampipe Table: azuredevopssh_git_pull_request_reviewer - Query Azure DevOps Git Pull Request Reviewers using SQL"
description: "Allows users to query Git Pull Request Reviewers in Azure DevOps, providing details on each reviewer, including their vote, whether their review is required, and the associated pull request."
---

# Table: azuredevopssh_git_pull_request_reviewer - Query Azure DevOps Git Pull Request Reviewers using SQL

Azure DevOps is a service from Microsoft that provides developer services for support teams to plan work, collaborate on code development, and build and deploy applications. Pull Request Reviewers in Azure DevOps are the identified users who have been assigned to review a pull request, and their votes determine whether the change can be merged.

## Table Usage Guide

The `azuredevopssh_git_pull_request_reviewer` table provides insights into Git Pull Request Reviewers within Azure DevOps. As a DevOps engineer, explore reviewer details through this table, including each reviewer's identity, vote, and whether their approval is required. Utilize it to track code review coverage and approval status across your repositories.

## Examples

### Basic info
Get the general information about pull request reviewers, such as their display name, vote, and whether their review is required.

```sql+postgres
select
  reviewer_id,
  pull_request_id,
  repository_id,
  display_name,
  unique_name,
  vote,
  is_required
from
  azuredevopssh_git_pull_request_reviewer;
```

```sql+sqlite
select
  reviewer_id,
  pull_request_id,
  repository_id,
  display_name,
  unique_name,
  vote,
  is_required
from
  azuredevopssh_git_pull_request_reviewer;
```

### List reviewers who approved
Find reviewers who have approved a pull request (vote of 10).

```sql+postgres
select
  pull_request_id,
  display_name,
  unique_name,
  vote
from
  azuredevopssh_git_pull_request_reviewer
where
  vote = 10;
```

```sql+sqlite
select
  pull_request_id,
  display_name,
  unique_name,
  vote
from
  azuredevopssh_git_pull_request_reviewer
where
  vote = 10;
```

### List required reviewers
Identify reviewers whose approval is required before a pull request can be merged.

```sql+postgres
select
  pull_request_id,
  display_name,
  vote
from
  azuredevopssh_git_pull_request_reviewer
where
  is_required;
```

```sql+sqlite
select
  pull_request_id,
  display_name,
  vote
from
  azuredevopssh_git_pull_request_reviewer
where
  is_required;
```

### List reviewers waiting on action
Find reviewers assigned to a pull request who have not yet voted.

```sql+postgres
select
  pull_request_id,
  display_name,
  unique_name
from
  azuredevopssh_git_pull_request_reviewer
where
  vote = 0;
```

```sql+sqlite
select
  pull_request_id,
  display_name,
  unique_name
from
  azuredevopssh_git_pull_request_reviewer
where
  vote = 0;
```