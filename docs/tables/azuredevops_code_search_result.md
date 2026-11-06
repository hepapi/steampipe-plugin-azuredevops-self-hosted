---
title: "Steampipe Table: azuredevops_code_search_result - Query Azure DevOps Code Search using SQL"
description: "Allows users to search Azure DevOps code and return matching files."
---

# Table: azuredevops_code_search_result - Query Azure DevOps Code Search using SQL

A code search result is a file that matches `search_text`. `search_text` is required. Set `project_id` to search one project. Code Search must be installed on the server.

## Examples

### Search the collection

```sql+postgres
select
  file_name,
  path,
  project_name,
  repository_name
from
  azuredevops_code_search_result
where
  search_text = 'func'
limit
  20;
```

```sql+sqlite
select
  file_name,
  path,
  project_name,
  repository_name
from
  azuredevops_code_search_result
where
  search_text = 'func'
limit
  20;
```
