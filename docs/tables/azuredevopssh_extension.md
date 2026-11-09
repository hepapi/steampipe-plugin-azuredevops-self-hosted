---
title: "Steampipe Table: azuredevopssh_extension - Query Azure DevOps Extensions using SQL"
description: "Allows users to query extensions installed in an Azure DevOps collection."
---

# Table: azuredevopssh_extension - Query Azure DevOps Extensions using SQL

An extension is a gallery extension installed in the collection. `publisher_id` and `extension_id` together fetch one installation. Those values are the publisher and extension identifiers, not the display names.

## Examples

### Installed extensions

```sql
select
  publisher_id,
  extension_id,
  extension_name,
  version
from
  azuredevopssh_extension
limit
  20;
```

