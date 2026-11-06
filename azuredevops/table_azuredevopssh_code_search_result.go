package azuredevops

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const codeSearchPageSize = 25

// codeSearchRow keeps the required search text on the row so Steampipe keeps
// the results of that search.
type codeSearchRow struct {
	search.CodeResult
	SearchText string
}

func tableAzureDevOpsCodeSearchResult(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_code_search_result",
		Description: "Search code in Azure DevOps repositories.",
		List: &plugin.ListConfig{
			Hydrate: listCodeSearchResults,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "search_text", Require: plugin.Required},
				{Name: "project_id", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "search_text",
				Description: "Text that was searched.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "file_name",
				Description: "Name of the matching file.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "path",
				Description: "Path of the matching file.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "content_id",
				Description: "Content identifier of the matching file.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "Project that contains the file. Set this to search one project.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Project.Id"),
			},
			{
				Name:        "project_name",
				Description: "Name of the project that contains the file.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Project.Name"),
			},
			{
				Name:        "repository_id",
				Description: "Repository that contains the file.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Repository.Id"),
			},
			{
				Name:        "repository_name",
				Description: "Name of the repository that contains the file.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Repository.Name"),
			},
			{
				Name:        "matches",
				Description: "Hit offsets in the file.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "versions",
				Description: "Versions of the matching file.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("FileName"),
			},
		}),
	}
}

func listCodeSearchResults(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	searchText := qualString(d, "search_text")
	if searchText == "" {
		return nil, fmt.Errorf("azuredevopssh_code_search_result requires search_text")
	}
	projectID := qualString(d, "project_id")

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_code_search_result.listCodeSearchResults", "connection_error", err)
		return nil, err
	}
	client, err := search.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_code_search_result.listCodeSearchResults", "client_error", err)
		return nil, err
	}

	skip := 0
	for {
		remaining := d.RowsRemaining(ctx)
		if remaining == 0 {
			return nil, nil
		}
		top := codeSearchPageSize
		if remaining < int64(top) {
			top = int(remaining)
		}
		request := search.CodeSearchRequest{
			SearchText:     &searchText,
			Skip:           &skip,
			Top:            &top,
			IncludeFacets:  boolPtr(false),
			IncludeSnippet: boolPtr(false),
		}
		response, err := client.FetchCodeSearchResults(ctx, search.FetchCodeSearchResultsArgs{
			Request: &request,
			Project: optionalProject(projectID),
		})
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_code_search_result.listCodeSearchResults", "api_error", err)
			return nil, err
		}
		if response == nil || response.Results == nil || len(*response.Results) == 0 {
			return nil, nil
		}
		for _, item := range *response.Results {
			d.StreamListItem(ctx, codeSearchRow{CodeResult: item, SearchText: searchText})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		skip += len(*response.Results)
		if len(*response.Results) < top {
			return nil, nil
		}
		if response.Count != nil && skip >= *response.Count {
			return nil, nil
		}
	}
}
