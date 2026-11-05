---
title: "Steampipe Table: azuredevops_git_pull_request_comment - Query Azure DevOps Git Pull Request Comments using SQL"
description: "Allows users to query Git Pull Request Comments in Azure DevOps, providing details on each code review comment, including its content, author, and associated thread and pull request."
---

# Table: azuredevops_git_pull_request_comment - Query Azure DevOps Git Pull Request Comments using SQL

Azure DevOps is a service from Microsoft that provides developer services for support teams to plan work, collaborate on code development, and build and deploy applications. Pull Request Comments in Azure DevOps are individual messages left within a code review thread, where reviewers and authors discuss specific aspects of a proposed change.

## Table Usage Guide

The `azuredevops_git_pull_request_comment` table provides insights into Git Pull Request Comments within Azure DevOps. As a DevOps engineer, explore comment details through this table, including the comment content, author, comment type, and the thread and pull request it belongs to. Utilize it to review and audit code review feedback across your repositories.

## Examples

### Basic info
Get the general information about pull request comments, such as their content and authors.

```sql+postgres
select
  comment_id,
  thread_id,
  pull_request_id,
  repository_id,
  content,
  comment_type,
  author
from
  azuredevops_git_pull_request_comment;
```

```sql+sqlite
select
  comment_id,
  thread_id,
  pull_request_id,
  repository_id,
  content,
  comment_type,
  author
from
  azuredevops_git_pull_request_comment;
```

### List text comments
Find all comments that are of the text type, which are the primary form of feedback on a pull request.

```sql+postgres
select
  comment_id,
  thread_id,
  pull_request_id,
  content
from
  azuredevops_git_pull_request_comment
where
  comment_type = 'text';
```

```sql+sqlite
select
  comment_id,
  thread_id,
  pull_request_id,
  content
from
  azuredevops_git_pull_request_comment
where
  comment_type = 'text';
```

### List recently published comments
Find comments that were published within the last 7 days.

```sql+postgres
select
  comment_id,
  pull_request_id,
  content,
  published_date
from
  azuredevops_git_pull_request_comment
where
  published_date >= now() - interval '7 days';
```

```sql+sqlite
select
  comment_id,
  pull_request_id,
  content,
  published_date
from
  azuredevops_git_pull_request_comment
where
  published_date >= datetime('now', '-7 days');
```

### List comments on a specific pull request
Filter comments by pull request to review all feedback for a single change.

```sql+postgres
select
  comment_id,
  thread_id,
  content,
  author
from
  azuredevops_git_pull_request_comment
where
  pull_request_id = 42;
```

```sql+sqlite
select
  comment_id,
  thread_id,
  content,
  author
from
  azuredevops_git_pull_request_comment
where
  pull_request_id = 42;
```