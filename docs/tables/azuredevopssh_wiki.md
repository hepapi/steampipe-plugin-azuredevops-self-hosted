---
title: "Steampipe Table: azuredevopssh_wiki - Query Azure DevOps Wikis using SQL"
description: "Allows users to query Azure DevOps project and code wikis."
---

# Table: azuredevopssh_wiki - Query Azure DevOps Wikis using SQL

A wiki is either a project wiki or a code wiki published from a git folder. The `azuredevopssh_wiki` table lists wikis in the collection. Set `project_id` to list wikis in one project. `id` fetches one wiki.

## Examples

### Wikis in the collection

```sql
select
  id,
  name,
  type,
  project_id,
  mapped_path
from
  azuredevopssh_wiki
limit
  20;
```

