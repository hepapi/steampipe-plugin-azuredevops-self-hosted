---
title: "Steampipe Table: azuredevopssh_variable_group - Query Azure DevOps Variable Groups using SQL"
description: "Allows users to query Azure DevOps variable groups, including their variables."
---

# Table: azuredevopssh_variable_group - Query Azure DevOps Variable Groups using SQL

A variable group stores pipeline variables for a project. Set `project_id` to read one project. A query without `project_id` visits every project. `id` and `project_id` together fetch one group.

The `variables` column is the map returned by Azure DevOps. It includes secret values when the connection token is allowed to read them.

## Examples

### Variable groups in a project

```sql
select
  id,
  name,
  type,
  is_shared
from
  azuredevopssh_variable_group
where
  project_id = '00000000-0000-0000-0000-000000000000';
```

