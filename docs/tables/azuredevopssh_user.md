---
title: "Steampipe Table: azuredevopssh_user - Query Azure DevOps Users using SQL"
description: "Allows users to query Azure DevOps Users, specifically user details, providing insights into user roles, access, and permissions."
---

# Table: azuredevopssh_user - Query Azure DevOps Users using SQL

The `azuredevopssh_user` table lists users in the collection. It reads the Graph API when that location is registered. Azure DevOps Server 2022.1 often does not register the v7 Graph users location. The table then calls `/_apis/graph/users` directly. When that route is also missing, it lists users by walking the Valid Users group through the Identity service. On that path, `membership_state` is the identity active flag and `memberships` are the user's direct groups.

## Table Usage Guide

The `azuredevopssh_user` table provides insights into users within Azure DevOps. As a DevOps engineer or system administrator, explore user-specific details through this table, including roles, access levels, and associated metadata. Utilize it to uncover information about users, such as their access permissions, the projects they are associated with, and their activity patterns.

## Examples

### Basic info
Explore which users are active in your Azure DevOps environment. This can help in managing user access and understanding the distribution of users across different domains and origins.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin
from
  azuredevopssh_user;
```

### List inactive users
Explore which users in your Azure DevOps environment are inactive. This can help maintain system security by identifying potential unused or unnecessary accounts.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin
from
  azuredevopssh_user
where
  not membership_state;
```

### List users from aad
Explore which Azure DevOps users have their origin set to 'aad', enabling you to understand their source and manage user access effectively. This is particularly useful in large organizations where user management can be complex.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin
from
  azuredevopssh_user
where
  origin = 'aad';
```

### List the users that have been deleted in the identity provider
Explore the users who have been removed from the identity provider. This query is useful for auditing and maintaining security compliance by tracking changes in user access.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin
from
  azuredevopssh_user
where
  is_deleted_in_origin;
```

### List users who are part of `Project Collection Administrators` group
Explore which users are part of a specific administrative group within a project collection. This is useful for auditing purposes, allowing you to ensure only authorized individuals have administrative access.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin
from
  azuredevopssh_user,
  jsonb_array_elements(memberships) as m
where
  m ->> 'containerDescriptor' in
  (
    select
      descriptor
    from
      azuredevopssh_group
    where
      display_name = 'Project Collection Administrators'
  );
```
