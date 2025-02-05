package pipes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipes-sdk-go"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state_reason": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"desired_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"db1.shared", "db1.small"}, false),
			},
			"db_volume_size_bytes": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"hive": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"host": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp pipes.Workspace
	var handle, instanceType, desiredState string
	var dbVolumeSizeBytes int64

	if value, ok := d.GetOk("handle"); ok {
		handle = value.(string)
	}
	if value, ok := d.GetOk("instance_type"); ok {
		instanceType = value.(string)
	}
	if value, ok := d.GetOk("db_volume_size_bytes"); ok {
		dbVolumeSizeBytes = int64(value.(int))
	}
	// Default instance type to `db1.shared`
	if instanceType == "" {
		instanceType = "db1.shared"
	}
	if value, ok := d.GetOk("desired_state"); ok {
		desiredState = value.(string)
	}

	// Create request
	req := pipes.CreateWorkspaceRequest{Handle: handle, InstanceType: (*pipes.WorkspaceInstanceType)(&instanceType)}
	if desiredState != "" {
		req.DesiredState = (*pipes.DesiredState)(&desiredState)
	}
	if dbVolumeSizeBytes != 0 {
		req.DbVolumeSizeBytes = &dbVolumeSizeBytes
	}

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandler string
		userHandler, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaces.Create(ctx, userHandler).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaces.Create(ctx, orgHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Workspace created: %s", resp.Handle)

	// Set property values
	d.Set("handle", resp.Handle)
	d.Set("organization", orgHandle)
	d.Set("workspace_id", resp.Id)
	d.Set("workspace_state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("desired_state", resp.DesiredState)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("instance_type", resp.InstanceType)
	d.Set("db_volume_size_bytes", resp.DbVolumeSizeBytes)
	d.Set("database_name", resp.DatabaseName)
	d.Set("hive", resp.Hive)
	d.Set("host", resp.Host)
	d.Set("identity_id", resp.IdentityId)
	d.Set("version_id", resp.VersionId)

	// If workspace is created inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle" otherwise "WorkspaceHandle"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s", orgHandle, resp.Handle))
	} else {
		d.SetId(resp.Handle)
	}

	return diags
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var orgHandle, workspaceHandle string
	var isUser = false

	var diags diag.Diagnostics
	var resp pipes.Workspace
	var err error
	var r *http.Response

	id := d.Id()

	// For backward-compatibility, we see whether the id contains : or /
	separator := "/"
	if strings.Contains(id, ":") {
		separator = ":"
	}
	// If workspace exists inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle" otherwise "WorkspaceHandle"
	ids := strings.Split(id, separator)
	if len(ids) == 2 {
		orgHandle = ids[0]
		workspaceHandle = ids[1]
	} else if len(ids) == 1 {
		isUser = true
		workspaceHandle = ids[0]
	}

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaces.Get(ctx, actorHandle, workspaceHandle).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaces.Get(ctx, orgHandle, workspaceHandle).Execute()
	}

	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Workspace (%s) not found", workspaceHandle),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("error reading %s: %v ", workspaceHandle, decodeResponse(r))
	}

	// assign results back into ResourceData
	d.Set("workspace_id", resp.Id)
	d.Set("handle", resp.Handle)
	d.Set("organization", orgHandle)
	d.Set("workspace_state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("desired_state", resp.DesiredState)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("instance_type", resp.InstanceType)
	d.Set("db_volume_size_bytes", resp.DbVolumeSizeBytes)
	d.Set("database_name", resp.DatabaseName)
	d.Set("hive", resp.Hive)
	d.Set("host", resp.Host)
	d.Set("identity_id", resp.IdentityId)
	d.Set("version_id", resp.VersionId)
	if separator == ":" {
		d.SetId(strings.ReplaceAll(id, ":", "/"))
	}

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var desiredState string
	var dbVolumeSizeBytes int64

	client := meta.(*PipesClient)

	oldHandle, newHandle := d.GetChange("handle")
	if value, ok := d.GetOk("desired_state"); ok {
		desiredState = value.(string)
	}
	if value, ok := d.GetOk("db_volume_size_bytes"); ok {
		dbVolumeSizeBytes = int64(value.(int))
	}

	// Create request
	req := pipes.UpdateWorkspaceRequest{
		Handle: types.String(newHandle.(string)),
	}
	if desiredState != "" {
		req.DesiredState = (*pipes.DesiredState)(&desiredState)
	}
	if dbVolumeSizeBytes != 0 {
		req.DbVolumeSizeBytes = &dbVolumeSizeBytes
	}

	log.Printf("\n[DEBUG] Updating Workspace: %s", *req.Handle)

	var resp pipes.Workspace
	var userHandler string
	var err error
	var r *http.Response

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		userHandler, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionUpdate. getUserHandler error:	%v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaces.Update(ctx, userHandler, oldHandle.(string)).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaces.Update(ctx, orgHandle, oldHandle.(string)).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error updating workspace: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Workspace updated: %s", resp.Handle)

	// Update state file
	d.SetId(resp.Handle)
	d.Set("handle", resp.Handle)
	d.Set("organization", orgHandle)
	d.Set("workspace_id", resp.Id)
	d.Set("workspace_state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("desired_state", resp.DesiredState)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("instance_type", resp.InstanceType)
	d.Set("db_volume_size_bytes", resp.DbVolumeSizeBytes)
	d.Set("database_name", resp.DatabaseName)
	d.Set("hive", resp.Hive)
	d.Set("host", resp.Host)
	d.Set("identity_id", resp.IdentityId)
	d.Set("version_id", resp.VersionId)

	// If workspace is created inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle" otherwise "WorkspaceHandle"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s", orgHandle, resp.Handle))
	} else {
		d.SetId(resp.Handle)
	}

	return diags
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var workspaceHandle string
	if value, ok := d.GetOk("handle"); ok {
		workspaceHandle = value.(string)
	}

	var err error
	var r *http.Response

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaces.Delete(ctx, actorHandle, workspaceHandle).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaces.Delete(ctx, orgHandle, workspaceHandle).Execute()
	}

	if err != nil {
		return diag.Errorf("error deleting workspace: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
