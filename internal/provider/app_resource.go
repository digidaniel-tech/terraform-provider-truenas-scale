// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/digidaniel/terraform-provider-truenas-scale/internal/websockethelper"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AppResource{}
var _ resource.ResourceWithImportState = &AppResource{}

func truenas_app() resource.Resource {
	return &AppResource{}
}

type AppResource struct {
	client *http.Client
}

type AppResourceModel struct {
    Name                types.String    `tfsdk:"name"`
    Id                  types.String    `tfsdk:"id"`
    State               types.String    `tfsdk:"state"`
    UpgradeAvailable    types.Bool      `tfsdk:"upgrade_available"`
    HumanVersion        types.String    `tfsdk:"human_version"`
    Version             types.String    `tfsdk:"version"`
    Metadata            types.Object    `tfsdk:"metadata"`
    ActiveWorkloads     types.Object    `tfsdk:"active_workloads"`
}

type WebSocketMessage struct {
    ID               string            `json:"id"`
    Name             string            `json:"name"`
    State            string            `json:"state"`
    UpgradeAvailable bool              `json:"upgrade_available"`
    HumanVersion     string            `json:"human_version"`
    Version          string            `json:"version"`
    Metadata         map[string]string `json:"metadata"`
    ActiveWorkloads  map[string]string `json:"active_workloads"`
}

type JobResponse struct {
    Msg    string               `json:"msg"`
	ID     string               `json:"id"`
	Result []JobResponseResult  `json:"result"`
}

type JobResponseResult struct {
	ID       int                    `json:"id"`
	State    string                 `json:"state"`
	Progress JobResponseProgress    `json:"progress"`
	Error  interface{}              `json:"error"`
	Result interface{}              `json:"result"`
}

type JobResponseProgress struct {
    Percent     float64 `json:"percent"`
	Description string  `json:"description"`
}

func (r *AppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *AppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "App resource",

		Attributes: map[string]schema.Attribute{
            "custom_app": schema.BoolAttribute{
                MarkdownDescription:    "Catalog or custom app",
                Optional:               true,
                Default:                booldefault.StaticBool(false),
            },
            "values": schema.ObjectAttribute{
                MarkdownDescription:    "Application settings, ex. volumes, environment variables.",
                Optional:               true,
            },
            "custom_compose_config": schema.ObjectAttribute{
                MarkdownDescription:    "Custom app configuration as an object",
                Optional:               true,
            },
            "custom_compose_config_string": schema.StringAttribute{
                MarkdownDescription:    "Custom app configuration as yaml",
                Optional:               true,
            },
            "catalog_app": schema.StringAttribute{
                MarkdownDescription:    "Catalog app to use when installing application",
                Optional:               true,
            },
            "app_name": schema.StringAttribute{
                MarkdownDescription:    "Application name to be set for installed application",
                Required:               true,
            },
            "train": schema.StringAttribute{
                MarkdownDescription:    "Train to use when download application, ex. stable, test, community.",
                Optional:               true,
            },
            "version": schema.StringAttribute{
                MarkdownDescription:    "Version of application to use",
                Optional:               true,
            },
		},
	}
}

func (r *AppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *AppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

    url := "wss://nas.home.wollbro.se/api/v1/websocket"

    client, err := websockethelper.NewWebSocketClient(url)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create app, got error: %s", err))
        return
    }
    defer client.Close()

    requestID, err := client.Send("catalog.sync_all", nil)
	if err != nil {
		resp.Diagnostics.AddError("WebSocket Error", fmt.Sprintf("Failed to send message: %s", err))
		return
	}

    response, err := client.Receive()
	if err != nil {
		resp.Diagnostics.AddError("WebSocket Error", fmt.Sprintf("Failed to read response: %s", err))
		return
	}

    jobID, ok := response["result"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Response Error", "Invalid job_id received")
		return
	}

    state, err := client.PollJobStatus(int(jobID))
	if err != nil {
		resp.Diagnostics.AddError("Job Status Error", fmt.Sprintf("Job failed: %s", err))
		return
	}

    // Logga resultatet
	tflog.Debug(ctx, fmt.Sprintf("Job completed with state: %s", state))

	// Uppdatera Terraform-staten
	data.Id = jobID       // Exempel på att sätta jobID som resursens ID
	data.Status = state        // Sätt job-status till state
	data.LastUpdated = time.Now().Format(time.RFC3339) // Sätt tidsstämpel

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AppResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *AppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
