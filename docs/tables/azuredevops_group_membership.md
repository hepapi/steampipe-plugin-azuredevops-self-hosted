---
title: "Steampipe Table: azuredevops_group_membership - Query Azure DevOps Group Membership using SQL"
description: "Allows users to query direct Azure DevOps group membership for a subject."
---

# Table: azuredevops_group_membership - Query Azure DevOps Group Membership using SQL

A membership row is one direct edge between a group and a member. `subject_descriptor` is required. `direction` defaults to `down` (members of the subject). `up` lists groups the subject belongs to.

The table uses the Graph membership API when that location is registered. Azure DevOps Server 2022.1 does not register it, so the table reads direct membership from the Identity service instead. Those descriptors are legacy identity descriptors. Pass either a graph subject descriptor or a legacy identity descriptor.

## Examples

### Members of a group

```sql+postgres
select
  container_descriptor,
  member_descriptor,
  direction
from
  azuredevops_group_membership
where
  subject_descriptor = 'vssgp.example'
limit
  20;
```

```sql+sqlite
select
  container_descriptor,
  member_descriptor,
  direction
from
  azuredevops_group_membership
where
  subject_descriptor = 'vssgp.example'
limit
  20;
```

### Groups a subject belongs to

```sql+postgres
select
  container_descriptor,
  member_descriptor
from
  azuredevops_group_membership
where
  subject_descriptor = 'vssgp.example'
  and direction = 'up';
```

```sql+sqlite
select
  container_descriptor,
  member_descriptor
from
  azuredevops_group_membership
where
  subject_descriptor = 'vssgp.example'
  and direction = 'up';
```
