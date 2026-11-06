---
title: "Steampipe Table: azuredevopssh_test_case - Query Azure DevOps Test Cases using SQL"
description: "Allows users to query test cases in one Azure DevOps test suite."
---

# Table: azuredevopssh_test_case - Query Azure DevOps Test Cases using SQL

A test case is the work item executed from a suite. `project_id`, `plan_id`, and `suite_id` are required. `id` is the test case work item id. Those four columns together fetch one case.

## Examples

### Cases in a suite

```sql+postgres
select
  id,
  name,
  "order"
from
  azuredevopssh_test_case
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and suite_id = 34
order by
  "order";
```

```sql+sqlite
select
  id,
  name,
  "order"
from
  azuredevopssh_test_case
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and suite_id = 34
order by
  "order";
```

### One case

```sql+postgres
select
  id,
  name,
  point_assignments
from
  azuredevopssh_test_case
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and suite_id = 34
  and id = 1001;
```

```sql+sqlite
select
  id,
  name,
  point_assignments
from
  azuredevopssh_test_case
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and suite_id = 34
  and id = 1001;
```
