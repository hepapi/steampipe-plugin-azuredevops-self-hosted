package azuredevops

import (
	"context"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// azureDevOpsToTime converts a timestamp column that the Microsoft Azure DevOps
// Go SDK exposes as its custom azuredevops.Time wrapper (rather than a native
// time.Time) into a value the Steampipe SDK can coerce for a
// ColumnType_TIMESTAMP column.
//
// The SDK represents every date/time field as azuredevops.Time{Time time.Time},
// and its JSON unmarshaler returns that struct (or a *azuredevops.Time). Without
// a transform, Steampipe's interfaceToColumnValue fails with "cannot convert ...
// to a time.Time", which aborts the whole query. This transform unwraps the
// wrapper so the value becomes a native *time.Time.
func azureDevOpsToTime(_ context.Context, d *transform.TransformData) (interface{}, error) {
	switch v := d.Value.(type) {
	case azuredevops.Time:
		return v.Time, nil
	case *azuredevops.Time:
		if v == nil {
			return nil, nil
		}
		return v.Time, nil
	case time.Time:
		return v, nil
	case *time.Time:
		if v == nil {
			return nil, nil
		}
		return *v, nil
	default:
		// Pass through anything the SDK can already coerce (e.g. a string), or
		// nil for null values.
		if d.Value == nil {
			return nil, nil
		}
		return d.Value, nil
	}
}