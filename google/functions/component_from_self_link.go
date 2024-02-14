package functions

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = EchoFunction{}

func NewProjectFromSelfLinkFunction() function.Function {
	return &ProjectFromSelfLinkFunction{}
}

type ProjectFromSelfLinkFunction struct{}

func (f ProjectFromSelfLinkFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "project_id_from_self_link"
}

func (f ProjectFromSelfLinkFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Returns the project name within the resource self link or id provided as an argument.",
		Description: "Takes a single string argument, which should be a self link or id of a resource. This function will either return the project name from the input string or raise an error due to no project being present in the string. The function uses the presence of \"project/{{project}}\" in the input string to identify the project name, e.g. when the function is passed the self link \"https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-c/instances/my-instance\" as an argument it will return \"my-project\".",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "self_link",
				Description: "A self link of a resouce, or an id. For example, both \"https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-c/instances/my-instance\" and \"projects/my-project/zones/us-central1-c/instances/my-instance\" are valid inputs",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f ProjectFromSelfLinkFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {

	// Load arguments from function call
	var arg0 string
	resp.Diagnostics.Append(req.Arguments.GetArgument(ctx, 0, &arg0)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pattern := "projects/{project}/"                              // Human-readable pattern used in errors and warnings
	regex := regexp.MustCompile("projects/(?P<ProjectId>[^/]+)/") // Should match the pattern above

	submatches := regex.FindAllStringSubmatchIndex(arg0, -1)

	// Zero matches means unusable input; error returned
	if len(submatches) == 0 {
		resp.Diagnostics.AddArgumentError(
			0,
			"No project id is present in the input string",
			fmt.Sprintf("The input string \"%s\" doesn't contain the expected pattern \"%s\".", arg0, pattern),
		)
		resp.Diagnostics.Append(resp.Result.Set(ctx, "")...)
		return
	}

	// >1 matches means input usable but not ideal; issue warning
	if len(submatches) > 1 {
		resp.Diagnostics.AddArgumentWarning(
			0,
			"Ambiguous input string could contain more than one project id",
			fmt.Sprintf("The input string \"%s\" contains more than one match for the pattern \"%s\". Terraform will use the first found match.", arg0, pattern),
		)
	}

	// Return found project id
	submatch := submatches[0] // Take the only / left-most submatch
	template := "$ProjectId"
	result := []byte{}
	result = regex.ExpandString(result, template, arg0, submatch)
	projectId := string(result)
	resp.Diagnostics.Append(resp.Result.Set(ctx, projectId)...)
}
