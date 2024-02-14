package functions_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
)

func TestAccProviderFunction_project_id_from_self_link(t *testing.T) {
	t.Parallel()
	acctest.SkipIfVcr(t) // Need to determine if compatible with VCR, as functions are implemented in PF provider

	projectId := "my-project"
	projectIdRegex := regexp.MustCompile(fmt.Sprintf("^%s$", projectId))
	selfLink := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/us-central1-c/instances/my-instance", projectId)

	badInput := "zones/us-central1-c/instances/my-instance"

	context := map[string]interface{}{
		"output_name": "project_id_from_selflink",
		"self_link":   "",
	}

	acctest.VcrTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testProviderFunction_project_id_from_self_link(context, selfLink),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput(context["output_name"].(string), projectIdRegex),
				),
			},
			{
				Config: testProviderFunction_project_id_from_self_link(context, badInput),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput(context["output_name"].(string), projectIdRegex),
				),
				ExpectError: regexp.MustCompile("Error in function call"), // ExpectError doesn't inspect the specific error messages
			},
		},
	})
}

func testProviderFunction_project_id_from_self_link(context map[string]interface{}, selfLink string) string {
	context["self_link"] = selfLink

	return acctest.Nprintf(`
	# terraform block required for provider function to be found
	terraform {
		required_providers {
			google = {
				source = "hashicorp/google"
			}
		}
	}

	output "%{output_name}" {
		value = provider::google::project_id_from_self_link("%{self_link}")
	}
`, context)
}
