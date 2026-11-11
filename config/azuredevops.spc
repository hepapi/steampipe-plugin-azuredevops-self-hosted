connection "azuredevops" {
  plugin = "ghcr.io/hepapi/azuredevopssh"

  # `organization_url`: collection URL, or the server root when one connection should cover more than one collection. (Required)
  # Single collection: https://azuredevops.example.com/DefaultCollection
  # All collections:   https://azuredevops.example.com
  # Services example:  https://dev.azure.com/test
  # See https://learn.microsoft.com/en-us/azure/devops/extend/develop/work-with-urls?view=azure-devops&tabs=http.
  # Can also be set with the AZDO_ORG_SERVICE_URL environment variable.
  # organization_url = "https://azuredevops.example.com/DefaultCollection"

  # `collections`: optional allow-list used when organization_url is the server root,
  # or when you set "*" to discover every collection under a virtual directory.
  # Omit it to query every collection the token can list.
  # collections = ["DefaultCollection", "AnotherCollection"]

  # `excluded_collections`: optional deny-list. Discovered collections are queried
  # except these names. Use this for TFVC-only collections whose Git location is not registered.
  # A name in both lists is excluded. Matching is case-insensitive.
  # excluded_collections = ["LegacyTfs"]

  # `server_url`: Azure DevOps Server root URL (self-hosted only, optional).
  # Used to enumerate project collections and to read server-scoped tables
  # (azuredevopssh_project_collection, azuredevopssh_agent_pool, azuredevopssh_agent)
  # when organization_url does not name one collection.
  # When unset, it is derived from `organization_url` by stripping the last path segment.
  # Set it explicitly when the server root has a path, for example https://azuredevops.example.com/tfs.
  # server_url = "https://azuredevops.example.com/"

  # `personal_access_token`: Azure DevOps Personal Access Token. (Required)
  # For more information on the Personal Access Token, please see https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows.
  # Can also be set with the AZDO_PERSONAL_ACCESS_TOKEN environment variable.
  # personal_access_token = "......................"
}
