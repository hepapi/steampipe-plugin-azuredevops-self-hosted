---
title: "Steampipe Table: azuredevops_wiki_page - Query Azure DevOps Wiki Pages using SQL"
description: "Allows users to query pages in an Azure DevOps wiki."
---

# Table: azuredevops_wiki_page - Query Azure DevOps Wiki Pages using SQL

A wiki page is one document in a wiki. `project_id` and `wiki_id` are required. `id`, `project_id`, and `wiki_id` together fetch one page, including its content.

## Examples

### Pages in a wiki

```sql+postgres
select
  id,
  path
from
  azuredevops_wiki_page
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and wiki_id = '00000000-0000-0000-0000-000000000000'
limit
  20;
```

```sql+sqlite
select
  id,
  path
from
  azuredevops_wiki_page
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and wiki_id = '00000000-0000-0000-0000-000000000000'
limit
  20;
```
