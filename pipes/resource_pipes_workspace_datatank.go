package pipes

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/turbot/pipes-sdk-go"
)

func resourceWorkspaceDatatank() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceDatatankCreate,
		ReadContext:   resourceWorkspaceDatatankRead,
		UpdateContext: resourceWorkspaceDatatankUpdate,
		DeleteContext: resourceWorkspaceDatatankDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"datatank_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state": {
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
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
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
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceWorkspaceDatatankCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var handle, description, workspaceHandle, desiredState string
	var err error

	if value, ok := d.GetOk("handle"); ok {
		handle = value.(string)
	}
	if value, ok := d.GetOk("description"); ok {
		description = value.(string)
	}
	if value, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = value.(string)
	}
	if value, ok := d.GetOk("desired_state"); ok {
		desiredState = value.(string)
	}

	req := pipes.CreateDatatankRequest{
		Handle:      handle,
		Description: &description,
	}
	if desiredState != "" {
		req.DesiredState = (*pipes.DesiredState)(&desiredState)
	}

	client := meta.(*PipesClient)
	var resp pipes.Datatank
	var r *http.Response

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceDatatankCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceDatatanks.Create(ctx, actorHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceDatatanks.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceDatatankCreate. Create datatank api error  %v", decodeResponse(r))
	}

	d.Set("datatank_id", resp.Id)
	d.Set("organization", orgHandle)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("description", resp.Description)
	d.Set("state", resp.State)
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
	d.Set("version_id", resp.VersionId)

	// If datatank is created for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/DatatankHandle" otherwise "WorkspaceHandle/DatatankHandle"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Handle))
	}

	return diags
}

func resourceWorkspaceDatatankRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, datatankHandle string
	var isUser = false

	// If datatank is created for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/DatatankHandle" otherwise "WorkspaceHandle/DatatankHandle"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<datatank-handle>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		datatankHandle = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		datatankHandle = idParts[1]
	}

	if datatankHandle == "" {
		return diag.Errorf("resourceWorkspaceDatatankRead. Datatank handle not present.")
	}

	var resp pipes.Datatank
	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceDatatankRead. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceDatatanks.Get(context.Background(), actorHandle, workspaceHandle, datatankHandle).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceDatatanks.Get(context.Background(), orgHandle, workspaceHandle, datatankHandle).Execute()
	}
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Datatank (%s) not found", datatankHandle),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceWorkspaceDatatankRead. Get datatank error: %v", decodeResponse(r))
	}

	d.Set("datatank_id", resp.Id)
	d.Set("organization", orgHandle)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("description", resp.Description)
	d.Set("state", resp.State)
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
	d.Set("version_id", resp.VersionId)

	// If datatank is created for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/DatatankHandle" otherwise "WorkspaceHandle/DatatankHandle"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Handle))
	}

	return diags
}

func resourceWorkspaceDatatankUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var datatankHandle, workspaceHandle, description, desiredState string
	var r *http.Response
	var resp pipes.Datatank
	var err error

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if value, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = value.(string)
	}
	if value, ok := d.GetOk("handle"); ok {
		datatankHandle = value.(string)
	}
	if value, ok := d.GetOk("description"); ok {
		description = value.(string)
	}
	if value, ok := d.GetOk("desired_state"); ok {
		desiredState = value.(string)
	}

	req := pipes.UpdateDatatankRequest{
		Description: &description,
	}
	if desiredState != "" {
		req.DesiredState = (*pipes.DesiredState)(&desiredState)
	}

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceDatatankUpdate. getUserHandler error:	%v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceDatatanks.Update(context.Background(), actorHandle, workspaceHandle, datatankHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceDatatanks.Update(context.Background(), orgHandle, workspaceHandle, datatankHandle).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceDatatankUpdate. Update datatank error: %v", decodeResponse(r))
	}

	d.Set("datatank_id", resp.Id)
	d.Set("organization", orgHandle)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("description", resp.Description)
	d.Set("state", resp.State)
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
	d.Set("version_id", resp.VersionId)

	// If datatank is created for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/DatatankHandle" otherwise "WorkspaceHandle/DatatankHandle"
	if strings.HasPrefix(resp.IdentityId, "o_") {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Handle))
	}

	return diags
}

func resourceWorkspaceDatatankDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var datatankHandle, workspaceHandle string

	if value, ok := d.GetOk("handle"); ok {
		datatankHandle = value.(string)
	}
	if value, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = value.(string)
	}

	var err error
	var r *http.Response
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceDatatankDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceDatatanks.Delete(ctx, actorHandle, workspaceHandle, datatankHandle).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceDatatanks.Delete(ctx, orgHandle, workspaceHandle, datatankHandle).Execute()
	}

	if err != nil {
		return diag.Errorf("resourceWorkspaceDatatankDelete. Delete datatank error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}
