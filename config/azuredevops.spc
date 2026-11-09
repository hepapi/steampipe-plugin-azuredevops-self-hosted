connection "azuredevops" {
  plugin = "ghcr.io/hepapi/azuredevopssh"

  # `organization_url`: collection URL on Azure DevOps Server, or an organization URL on Azure DevOps Services. (Required)
  # Server example: https://azuredevops.example.com/DefaultCollection
  # Services example: https://dev.azure.com/test
  # See https://learn.microsoft.com/en-us/azure/devops/extend/develop/work-with-urls?view=azure-devops&tabs=http.
  # Can also be set with the AZDO_ORG_SERVICE_URL environment variable.
  # organization_url = "https://azuredevops.example.com/DefaultCollection"

  # `server_url`: Azure DevOps Server root URL (self-hosted only, optional).
  # Used only by the `azuredevopssh_project_collection` table to enumerate collections.
  # When unset, it is derived from `organization_url` by stripping the last path segment.
  # For self-hosted, e.g. https://azuredevops.example.com/
  # server_url = "https://azuredevops.example.com/"

  # `personal_access_token`: Azure DevOps Personal Access Token. (Required)
  # For more information on the Personal Access Token, please see https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows.
  # Can also be set with the AZDO_PERSONAL_ACCESS_TOKEN environment variable.
  # personal_access_token = "......................"
}
