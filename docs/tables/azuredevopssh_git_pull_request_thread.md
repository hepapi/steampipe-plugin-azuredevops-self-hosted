---
title: "Steampipe Table: azuredevopssh_git_pull_request_thread - Query Azure DevOps Git Pull Request Threads using SQL"
description: "Allows users to query Git Pull Request Threads in Azure DevOps, providing details on each code review thread, including its comments, status, and associated pull request."
---

# Table: azuredevopssh_git_pull_request_thread - Query Azure DevOps Git Pull Request Threads using SQL

Pull Request Threads in Azure DevOps represent code review conversations attached to a pull request, where reviewers can leave comments and discuss specific lines of code.

## Table Usage Guide

The `azuredevopssh_git_pull_request_thread` table provides insights into Git Pull Request Threads within Azure DevOps. As a DevOps engineer, explore thread details through this table, including the thread id, its status, whether it is active, and the associated pull request and repository. Utilize it to review and audit code review discussions across your repositories.

## Examples

### Basic info
Get the general information about pull request threads, such as their status and associated pull requests.

```sql
select
  thread_id,
  pull_request_id,
  repository_id,
  status,
  is_deleted
from
  azuredevopssh_git_pull_request_thread;
```

### List active pull request threads
Find code review threads that are still active within your pull requests.

```sql
select
  thread_id,
  pull_request_id,
  repository_id,
  status
from
  azuredevopssh_git_pull_request_thread
where
  status = 'active';
```

### List threads for a specific pull request
Filter threads by pull request to focus on the code review discussion for a single change.

```sql
select
  thread_id,
  status,
  thread_context
from
  azuredevopssh_git_pull_request_thread
where
  pull_request_id = 42;
```
