package pipes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	pipes "github.com/turbot/pipes-sdk-go"
)

// test suites
func TestAccUserWorkspacePipeline_Basic(t *testing.T) {
	resourceName := "pipes_workspace_pipeline.pipeline_1"
	processDataSourceName := "data.pipes_process.process_run"
	workspaceHandle := "workspace" + randomString(3)
	title := "Daily CIS Job"
	pipeline := "pipeline.snapshot_dashboard"
	mod := "github.com/turbot/steampipe-mod-aws-compliance"
	frequency := `
		{
			"type": "interval",
			"schedule": "daily"
		}
	`
	args := `
		{
			"resource": "aws_compliance.benchmark.cis_v140",
			"identity_type": "user",
			"identity_handle": "testuser",
			"workspace_handle": "dev",
			"inputs": {},
			"tags": {
				"series": "daily_cis"
			}
		}
	`
	tags := `
		{
			"name": "pipeline_1",
			"foo": "bar"
		}
	`
	updatedFrequency := `
		{
			"type": "interval",
			"schedule": "hourly"
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspacePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspacePipelineConfig(workspaceHandle, title, pipeline, frequency, args, tags, mod),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspacePipelineExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "pipeline", pipeline),
					resource.TestCheckResourceAttr(resourceName, "last_process_id", ""),
					TestJSONFieldEqual(t, resourceName, "frequency", frequency),
					TestJSONFieldEqual(t, resourceName, "args", args),
					TestJSONFieldEqual(t, resourceName, "tags", tags),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "args", "frequency", "tags"},
			},
			{
				Config: testAccUserWorkspacePipelineUpdateConfig(workspaceHandle, title, pipeline, updatedFrequency, args, tags, mod),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspacePipelineExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "pipeline", pipeline),
					resource.TestCheckResourceAttr(resourceName, "last_process_id", ""),
					TestJSONFieldEqual(t, resourceName, "frequency", updatedFrequency),
					TestJSONFieldEqual(t, resourceName, "args", args),
					TestJSONFieldEqual(t, resourceName, "tags", tags),
					runPipeline(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "args", "frequency", "tags"},
			},
			{
				Config: testAccUserWorkspacePipelineProcessConfig(workspaceHandle, title, pipeline, updatedFrequency, args, tags, mod),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspacePipelineExists(workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "pipeline", pipeline),
					resource.TestMatchResourceAttr(resourceName, "last_process_id", regexp.MustCompile(`^p_[0-9a-v]{20}`)),
					TestJSONFieldEqual(t, resourceName, "frequency", updatedFrequency),
					TestJSONFieldEqual(t, resourceName, "args", args),
					TestJSONFieldEqual(t, resourceName, "tags", tags),
					resource.TestMatchResourceAttr(processDataSourceName, "process_id", regexp.MustCompile(`^p_[0-9a-v]{20}`)),
					resource.TestCheckResourceAttr(processDataSourceName, "type", "pipeline.command.run"),
				),
			},
		},
	})
}

func testAccUserWorkspacePipelineConfig(workspaceHandle, title, pipeline, frequency, args, tags, mod string) string {
	return fmt.Sprintf(`
	

	resource "pipes_workspace" "test_workspace" {
		handle = "%s"
	}

	resource "pipes_workspace_mod" "aws_compliance" {
		workspace_handle = pipes_workspace.test_workspace.handle
		path = "%s"
	}
	
	resource "pipes_workspace_pipeline" "pipeline_1" {
		workspace = pipes_workspace.test_workspace.handle
		title            = "%s"
		pipeline         = "%s"
		frequency        = jsonencode(%s)
		args             = jsonencode(%s)
		tags             = jsonencode(%s)

		depends_on = [pipes_workspace_mod.aws_compliance]
	}`, workspaceHandle, mod, title, pipeline, frequency, args, tags)
}

func testAccUserWorkspacePipelineUpdateConfig(workspaceHandle, title, pipeline, frequency, args, tags, mod string) string {
	return fmt.Sprintf(`
	resource "pipes_workspace" "test_workspace" {
		handle = "%s"
	}

	resource "pipes_workspace_mod" "aws_compliance" {
		workspace_handle = pipes_workspace.test_workspace.handle
		path = "%s"
	}
	
	resource "pipes_workspace_pipeline" "pipeline_1" {
		workspace = pipes_workspace.test_workspace.handle
		title            = "%s"
		pipeline         = "%s"
		frequency        = jsonencode(%s)
		args             = jsonencode(%s)
		tags             = jsonencode(%s)

		depends_on = [pipes_workspace_mod.aws_compliance]
	}`, workspaceHandle, mod, title, pipeline, frequency, args, tags)
}

func testAccUserWorkspacePipelineProcessConfig(workspaceHandle, title, pipeline, frequency, args, tags, mod string) string {
	return fmt.Sprintf(`
	resource "pipes_workspace" "test_workspace" {
		handle = "%s"
	}

	resource "pipes_workspace_mod" "aws_compliance" {
		workspace_handle = pipes_workspace.test_workspace.handle
		path = "%s"
	}
	
	resource "pipes_workspace_pipeline" "pipeline_1" {
		workspace = pipes_workspace.test_workspace.handle
		title            = "%s"
		pipeline         = "%s"
		frequency        = jsonencode(%s)
		args             = jsonencode(%s)
		tags             = jsonencode(%s)

		depends_on = [pipes_workspace_mod.aws_compliance]
	}
	
	data "pipes_process" "process_run" {
		workspace  = pipes_workspace.test_workspace.handle
		process_id = pipes_workspace_pipeline.pipeline_1.last_process_id
	}
	`, workspaceHandle, mod, title, pipeline, frequency, args, tags)
}

func testAccCheckWorkspacePipelineExists(workspaceHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*PipesClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pipes_workspace_pipeline" {
				continue
			}

			pipelineId := rs.Primary.Attributes["workspace_pipeline_id"]
			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""

			var err error
			if isUser {
				var userHandle string
				userHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %s", err)
				}
				_, _, err = client.APIClient.UserWorkspacePipelines.Get(ctx, userHandle, workspaceHandle, pipelineId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching pipeline %s in user workspace with handle %s. %s", pipelineId, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspacePipelines.Get(ctx, org, workspaceHandle, pipelineId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching pipeline %s in org workspace with handle %s. %s", pipelineId, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

func runPipeline() resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*PipesClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pipes_workspace_pipeline" {
				continue
			}

			var resp pipes.PipelineCommandResponse
			pipelineId := rs.Primary.Attributes["workspace_pipeline_id"]
			workspaceId := rs.Primary.Attributes["workspace_id"]
			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""

			// Create request
			req := pipes.PipelineCommandRequest{Command: "run"}

			var err error
			if isUser {
				var userHandle string
				userHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %s", err)
				}
				resp, _, err = client.APIClient.UserWorkspacePipelines.Command(ctx, userHandle, workspaceId, pipelineId).Request(req).Execute()
				if err != nil {
					return fmt.Errorf("error fetching pipeline %s in user workspace with handle %s. %s", pipelineId, workspaceId, err)
				}
			} else {
				resp, _, err = client.APIClient.OrgWorkspacePipelines.Command(ctx, org, workspaceId, pipelineId).Request(req).Execute()
				if err != nil {
					return fmt.Errorf("error fetching pipeline %s in org workspace with handle %s. %s", pipelineId, workspaceId, err)
				}
			}
			log.Printf("\n[DEBUG] Pipeline Run Response: %v", resp)
		}
		return nil
	}
}

// testAccCheckWorkspacePipelineDestroy verifies the pipeline has been deleted from the workspace
func testAccCheckWorkspacePipelineDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_workspace_pipeline" {
			continue
		}

		// Retrieve workspace handle and pipeline id by referencing it's state handle for API lookup
		workspaceHandle := rs.Primary.Attributes["workspace"]
		pipelineId := rs.Primary.Attributes["workspace_pipeline_id"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var userHandle string
			userHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspacePipelines.Get(ctx, userHandle, workspaceHandle, pipelineId).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspacePipelines.Get(ctx, org, workspaceHandle, pipelineId).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Pipeline %s/%s still exists", workspaceHandle, pipelineId)
		}

		if isUser {
			if r.StatusCode != 404 {
				log.Printf("[INFO] testAccCheckWorkspacePipelineDestroy testAccCheckUserWorkspacePipelineDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		} else {
			if r.StatusCode != 403 {
				log.Printf("[INFO] testAccCheckWorkspacePipelineDestroy testAccCheckUserWorkspacePipelineDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		}

	}

	return nil
}
