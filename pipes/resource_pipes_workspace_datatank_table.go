package pipes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	pipes "github.com/turbot/pipes-sdk-go"
)

func resourceWorkspaceDatatankTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceDatatankTableCreate,
		ReadContext:   resourceWorkspaceDatatankTableRead,
		UpdateContext: resourceWorkspaceDatatankTableUpdate,
		DeleteContext: resourceWorkspaceDatatankTableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"datatank_table_id": {
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
			"workspace_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"datatank_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"datatank_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9$]*$`), "Must be a valid postgres table name."),
			},
			"migrating_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"table", "query"}, false),
			},
			"part_per": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"connection"}, false),
			},
			"source_schema": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_table": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_query": {
				Type:     schema.TypeString,
				Optional: true,
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
			"freshness": {
				Type:             schema.TypeString,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: connectionJSONStringsEqual,
			},
			"migrating_freshness": {
				Type:             schema.TypeString,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: connectionJSONStringsEqual,
			},
			"frequency": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
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
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceWorkspaceDatatankTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var workspaceHandle, datatankHandle, name, description, tableType, partPer, sourceSchema, sourceTable, sourceQuery string
	var frequency pipes.PipelineFrequency
	var err error

	if value, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = value.(string)
	}
	if value, ok := d.GetOk("datatank_handle"); ok {
		datatankHandle = value.(string)
	}
	if value, ok := d.GetOk("name"); ok {
		name = value.(string)
	}
	if value, ok := d.GetOk("description"); ok {
		description = value.(string)
	}
	if value, ok := d.GetOk("type"); ok {
		tableType = value.(string)
	}
	if value, ok := d.GetOk("part_per"); ok {
		partPer = value.(string)
	}
	if value, ok := d.GetOk("source_schema"); ok {
		sourceSchema = value.(string)
	}
	if value, ok := d.GetOk("source_table"); ok {
		sourceTable = value.(string)
	}
	if value, ok := d.GetOk("source_query"); ok {
		sourceQuery = value.(string)
	}
	err = json.Unmarshal([]byte(d.Get("frequency").(string)), &frequency)
	if err != nil {
		return diag.Errorf("error parsing frequency for datatank table : %v", d.Get("frequency").(string))
	}

	req := pipes.CreateDatatankTableRequest{
		Name:         name,
		Description:  &description,
		Type:         tableType,
		PartPer:      &partPer,
		SourceSchema: &sourceSchema,
		SourceTable:  &sourceTable,
		SourceQuery:  &sourceQuery,
		Frequency:    &frequency,
	}

	client := meta.(*PipesClient)
	var resp pipes.DatatankTable
	var r *http.Response

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceDatatankTableCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceDatatankTables.Create(ctx, actorHandle, workspaceHandle, datatankHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceDatatankTables.Create(ctx, orgHandle, workspaceHandle, datatankHandle).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceDatatankTableCreate. Create datatank api error  %v", decodeResponse(r))
	}

	d.Set("datatank_table_id", resp.Id)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("datatank_id", resp.DatatankId)
	d.Set("datatank_handle", datatankHandle)
	d.Set("name", resp.Name)
	d.Set("migrating_name", resp.MigratingName)
	d.Set("description", resp.Description)
	d.Set("type", resp.Type)
	d.Set("part_per", resp.PartPer)
	d.Set("source_schema", resp.SourceSchema)
	d.Set("source_table", resp.SourceTable)
	d.Set("source_query", resp.SourceQuery)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("desired_state", resp.DesiredState)
	d.Set("freshness", resp.Freshness)
	d.Set("migrating_freshness", resp.MigratingFreshness)
	d.Set("frequency", resp.Frequency)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	// If datatank table is created for a datatank in a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/DatatankHandle/DatatankTableName" otherwise "WorkspaceHandle/DatatankHandle/DatatankTableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, datatankHandle, resp.Name))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, datatankHandle, resp.Name))
	}

	return diags
}

func resourceWorkspaceDatatankTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, datatankHandle, datatankTableName string
	var isUser = false

	// If datatank table is created for a datatank in a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/DatatankHandle/DatatankTableName" otherwise "WorkspaceHandle/DatatankHandle/DatatankTableName"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 3 && len(idParts) > 4 {
		return diag.Errorf("unexpected format of ID (%q), expected <workspace-handle>/<datatank-handle>/<datatank-table-name>", d.Id())
	}

	if len(idParts) == 4 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		datatankHandle = idParts[2]
		datatankTableName = idParts[3]
	} else if len(idParts) == 3 {
		isUser = true
		workspaceHandle = idParts[0]
		datatankHandle = idParts[1]
		datatankTableName = idParts[2]
	}

	if datatankHandle == "" {
		return diag.Errorf("resourceWorkspaceDatatankTableRead. Datatank handle not present.")
	}

	var resp pipes.DatatankTable
	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceDatatankTableRead. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceDatatankTables.Get(context.Background(), actorHandle, workspaceHandle, datatankHandle, datatankTableName).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceDatatankTables.Get(context.Background(), orgHandle, workspaceHandle, datatankHandle, datatankTableName).Execute()
	}
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Datatank Table (%s) not found", datatankTableName),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceWorkspaceDatatankTableRead. Get datatank error: %v", decodeResponse(r))
	}

	d.Set("datatank_table_id", resp.Id)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("datatank_id", resp.DatatankId)
	d.Set("datatank_handle", datatankHandle)
	d.Set("name", resp.Name)
	d.Set("migrating_name", resp.MigratingName)
	d.Set("description", resp.Description)
	d.Set("type", resp.Type)
	d.Set("part_per", resp.PartPer)
	d.Set("source_schema", resp.SourceSchema)
	d.Set("source_table", resp.SourceTable)
	d.Set("source_query", resp.SourceQuery)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("desired_state", resp.DesiredState)
	d.Set("freshness", resp.Freshness)
	d.Set("migrating_freshness", resp.MigratingFreshness)
	d.Set("frequency", resp.Frequency)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	// If datatank table is created for a datatank in a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/DatatankHandle/DatatankTableName" otherwise "WorkspaceHandle/DatatankHandle/DatatankTableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, datatankHandle, resp.Name))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, datatankHandle, resp.Name))
	}

	return diags
}

func resourceWorkspaceDatatankTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var diags diag.Diagnostics
	var workspaceHandle, datatankHandle, name, description, partPer, sourceSchema, sourceTable, sourceQuery, desiredState string
	var frequency pipes.PipelineFrequency

	if value, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = value.(string)
	}
	if value, ok := d.GetOk("datatank_handle"); ok {
		datatankHandle = value.(string)
	}
	if value, ok := d.GetOk("name"); ok {
		name = value.(string)
	}
	if value, ok := d.GetOk("description"); ok {
		description = value.(string)
	}
	if value, ok := d.GetOk("part_per"); ok {
		partPer = value.(string)
	}
	if value, ok := d.GetOk("source_schema"); ok {
		sourceSchema = value.(string)
	}
	if value, ok := d.GetOk("source_table"); ok {
		sourceTable = value.(string)
	}
	if value, ok := d.GetOk("source_query"); ok {
		sourceQuery = value.(string)
	}
	if value, ok := d.GetOk("desired_state"); ok {
		desiredState = value.(string)
	}
	err := json.Unmarshal([]byte(d.Get("frequency").(string)), &frequency)
	if err != nil {
		return diag.Errorf("error parsing frequency for datatank table : %v", d.Get("frequency").(string))
	}

	req := pipes.UpdateDatatankTableRequest{
		Name:         &name,
		Description:  &description,
		PartPer:      &partPer,
		SourceSchema: &sourceSchema,
		SourceTable:  &sourceTable,
		SourceQuery:  &sourceQuery,
		Frequency:    &frequency,
		DesiredState: &desiredState,
	}

	var r *http.Response
	var resp pipes.DatatankTable

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceDatatankTableUpdate. getUserHandler error:	%v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceDatatankTables.Update(context.Background(), actorHandle, workspaceHandle, datatankHandle, name).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceDatatankTables.Update(context.Background(), orgHandle, workspaceHandle, datatankHandle, name).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceDatatankTableUpdate. Update datatank error: %v", decodeResponse(r))
	}

	d.Set("datatank_table_id", resp.Id)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)
	d.Set("datatank_id", resp.DatatankId)
	d.Set("datatank_handle", datatankHandle)
	d.Set("name", resp.Name)
	d.Set("migrating_name", resp.MigratingName)
	d.Set("description", resp.Description)
	d.Set("type", resp.Type)
	d.Set("part_per", resp.PartPer)
	d.Set("source_schema", resp.SourceSchema)
	d.Set("source_table", resp.SourceTable)
	d.Set("source_query", resp.SourceQuery)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("desired_state", resp.DesiredState)
	d.Set("freshness", resp.Freshness)
	d.Set("migrating_freshness", resp.MigratingFreshness)
	d.Set("frequency", resp.Frequency)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	// If datatank table is created for a datatank in a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/DatatankHandle/DatatankTableName" otherwise "WorkspaceHandle/DatatankHandle/DatatankTableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, datatankHandle, resp.Name))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, datatankHandle, resp.Name))
	}

	return diags
}

func resourceWorkspaceDatatankTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var workspaceHandle, datatankHandle, name string

	if value, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = value.(string)
	}
	if value, ok := d.GetOk("datatank_handle"); ok {
		datatankHandle = value.(string)
	}
	if value, ok := d.GetOk("name"); ok {
		name = value.(string)
	}

	var err error
	var r *http.Response
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceDatatankTableDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceDatatankTables.Delete(ctx, actorHandle, workspaceHandle, datatankHandle, name).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceDatatankTables.Delete(ctx, orgHandle, workspaceHandle, datatankHandle, name).Execute()
	}

	if err != nil {
		return diag.Errorf("resourceWorkspaceDatatankTableDelete. Delete datatank error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}
