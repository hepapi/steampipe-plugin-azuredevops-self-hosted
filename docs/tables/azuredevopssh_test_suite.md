---
title: "Steampipe Table: azuredevopssh_test_suite - Query Azure DevOps Test Suites using SQL"
description: "Allows users to query suites in one Azure DevOps test plan."
---

# Table: azuredevopssh_test_suite - Query Azure DevOps Test Suites using SQL

A test suite is a folder of test cases inside a test plan. `project_id` and `plan_id` are required. Suites are returned as a flat list. `id`, `plan_id`, and `project_id` together fetch one suite.

## Examples

### Suites in a plan

```sql
select
  id,
  name,
  suite_type,
  parent_suite_id,
  has_children
from
  azuredevopssh_test_suite
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
order by
  name;
```

### One suite

```sql
select
  id,
  name,
  suite_type,
  query_string,
  requirement_id
from
  azuredevopssh_test_suite
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and id = 34;
```

