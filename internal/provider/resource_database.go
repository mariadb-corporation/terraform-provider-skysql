package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mariadb-corporation/skysql-api-go"
)

func resourceDatabase() *schema.Resource {
	s := make(map[string]*schema.Schema)

	s["last_updated"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	}

	// defaults for all fields
	for _, field := range databaseFields() {
		s[reservedNamesAtoT(field)] = &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		}
	}

	// overrides for fields included in create requests
	for _, field := range databaseCreateFields() {
		s[reservedNamesAtoT(field)].Computed = false
		s[reservedNamesAtoT(field)].Required = true
		s[reservedNamesAtoT(field)].ForceNew = true
	}

	// overrides for fields that may be updated in place
	for _, field := range databaseUpdateFields() {
		s[reservedNamesAtoT(field)].ForceNew = false
	}

	return &schema.Resource{
		Description: "MariaDB database service deployed by SkySQL",

		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,

		Schema: s,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*skysql.Client)

	// collect the attributes specified by the user
	attrs := make(map[string]interface{})
	for _, field := range databaseCreateFields() {
		attrs[field] = d.Get(reservedNamesAtoT(field))
	}

	// marshal them into a json byte string to take advantage
	// of the json tags on the generated type in the skysql SDK
	attrsJson, err := json.Marshal(attrs)
	if err != nil {
		diag.FromErr(err)
	}

	// unmarshal the attrs into a valid request body type
	var body skysql.CreateDatabaseJSONRequestBody
	if err = json.Unmarshal(attrsJson, &body); err != nil {
		diag.FromErr(err)
	}

	res, err := client.CreateDatabase(ctx, body)
	if err != nil {
		return diag.FromErr(err)
	}

	database, errDiag := decodeAPIResponseBody(res)
	if errDiag != nil {
		return errDiag
	}

	id := database["id"].(string)
	if id == "" {
		diag.FromErr(fmt.Errorf("bad response from SkySQL: %v", database))
	}
	d.SetId(id)

	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*skysql.Client)
	var diags diag.Diagnostics

	id := d.Id()

	database, err := readDatabase(ctx, client, id)
	if err != nil {
		return err
	}

	for _, field := range databaseFields() {
		d.Set(reservedNamesAtoT(field), database[field])
	}

	return diags
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*skysql.Client)
	id := d.Id()

	// collect the attributes specified by the user
	var updateNeeded bool
	attrs := make(map[string]interface{})
	for _, field := range databaseUpdateFields() {
		attrs[field] = d.Get(reservedNamesAtoT(field))
		updateNeeded = updateNeeded || d.HasChange(field)
	}

	if updateNeeded {
		// marshal them into a json byte string to take advantage
		// of the json tags on the generated type in the skysql SDK
		attrsJson, err := json.Marshal(attrs)
		if err != nil {
			diag.FromErr(err)
		}

		// unmarshal the attrs into a valid request body type
		var body skysql.UpdateDatabaseJSONRequestBody
		if err = json.Unmarshal(attrsJson, &body); err != nil {
			diag.FromErr(err)
		}

		res, err := client.UpdateDatabase(ctx, id, body)
		if err != nil {
			return diag.FromErr(err)
		}

		_, errDiag := decodeAPIResponseBody(res)
		if errDiag != nil {
			return errDiag
		}
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*skysql.Client)
	id := d.Id()

	res, err := client.DeleteDatabase(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = checkAPIStatus(res.StatusCode, res.Body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
