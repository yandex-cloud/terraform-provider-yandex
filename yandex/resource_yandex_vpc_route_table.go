package yandex

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

const yandexVPCRouteTableDefaultTimeout = 3 * time.Minute

func resourceYandexVPCRouteTable() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a route table within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts).\n\n* How-to Guides\n  * [Cloud Networking](https://yandex.cloud/docs/vpc/)\n",

		Create: resourceYandexVPCRouteTableCreate,
		Read:   resourceYandexVPCRouteTableRead,
		Update: resourceYandexVPCRouteTableUpdate,
		Delete: resourceYandexVPCRouteTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCRouteTableDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCRouteTableDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCRouteTableDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:        schema.TypeString,
				Description: "ID of the network this route table belongs to.",
				Required:    true,
				ForceNew:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Default:     "",
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"static_route": {
				Type:        schema.TypeSet,
				Description: "A list of static route records for the route table.\n\n~> Only one of `next_hop_address` or `gateway_id` should be specified.\n",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_prefix": {
							Type:        schema.TypeString,
							Description: "Route prefix in CIDR notation.",
							Optional:    true,
						},
						"next_hop_address": {
							Type:        schema.TypeString,
							Description: "Address of the next hop.",
							Optional:    true,
						},
						"gateway_id": {
							Type:        schema.TypeString,
							Description: "ID of the gateway used ad next hop.",
							Optional:    true,
						},
					},
				},
				Set: resourceYandexVPCRouteTableHash,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}

}

func resourceYandexVPCRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating route table: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating route table: %s", err)
	}

	staticRoutes, err := expandStaticRoutes(d.Get("static_route"))
	if err != nil {
		return fmt.Errorf("Error expanding static routes while creating route table: %s", err)
	}

	req := vpc.CreateRouteTableRequest{
		FolderId:     folderID,
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		NetworkId:    d.Get("network_id").(string),
		StaticRoutes: staticRoutes,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().RouteTable().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create route table: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get route table create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*vpc.CreateRouteTableMetadata)
	if !ok {
		return fmt.Errorf("could not get Route Table ID from create operation metadata")
	}

	d.SetId(md.RouteTableId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create route table: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("route table creation failed: %s", err)
	}

	return resourceYandexVPCRouteTableRead(d, meta)
}

func resourceYandexVPCRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	routeTable, err := config.sdk.VPC().RouteTable().Get(config.Context(), &vpc.GetRouteTableRequest{
		RouteTableId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Route table %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(routeTable.CreatedAt))
	d.Set("name", routeTable.Name)
	d.Set("folder_id", routeTable.FolderId)
	d.Set("description", routeTable.Description)
	d.Set("network_id", routeTable.NetworkId)

	if err := d.Set("labels", routeTable.Labels); err != nil {
		return err
	}

	return d.Set("static_route", flattenStaticRoutes(routeTable))
}

func resourceYandexVPCRouteTableUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	d.Partial(true)

	req := &vpc.UpdateRouteTableRequest{
		RouteTableId: d.Id(),
		UpdateMask:   &field_mask.FieldMask{},
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if d.HasChange("static_route") {
		_, new := d.GetChange("static_route")
		nrs := new.(*schema.Set).List()

		var newRoutes []*vpc.StaticRoute
		for _, route := range nrs {
			sr, err := routeDescriptionToStaticRoute(route)
			if err != nil {
				return err
			}
			newRoutes = append(newRoutes, sr)
		}

		req.StaticRoutes = newRoutes
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "static_routes")
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().RouteTable().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Route table %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Route table %q: %s", d.Id(), err)
	}

	d.Partial(false)

	return resourceYandexVPCRouteTableRead(d, meta)
}

func resourceYandexVPCRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Route table %q", d.Id())

	req := &vpc.DeleteRouteTableRequest{
		RouteTableId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().RouteTable().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Route table %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Route table %q", d.Id())
	return nil
}

func resourceYandexVPCRouteTableHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}

	if v, ok := m["next_hop_address"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["destination_prefix"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}
