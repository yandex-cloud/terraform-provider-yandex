package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/billing/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func resourceYandexBillingServiceInstanceBindingCreate(serviceInstanceType string, idFieldName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)

		billingAccountId := d.Get("billing_account_id").(string)
		serviceInstanceId := d.Get(idFieldName).(string)

		billableObject := billing.BillableObject{
			Type: serviceInstanceType,
			Id:   serviceInstanceId,
		}
		req := billing.BindBillableObjectRequest{
			BillingAccountId: billingAccountId,
			BillableObject:   &billableObject,
		}

		op, err := config.sdk.Billing().BillingAccount().BindBillableObject(
			config.Context(),
			&req,
		)

		if opErr := op.GetError(); opErr != nil {
			log.Printf("[WARN] Operation ended with error: %s", protojson.Format(opErr))

			return diag.FromErr(status.Error(codes.Code(opErr.Code), opErr.Message))
		}

		if err != nil {
			return diag.FromErr(fmt.Errorf("Error while requesting API binding cloud to billing account: %w", err))
		}

		id := bindServiceInstanceId{
			billingAccountId:    billingAccountId,
			serviceInstanceType: serviceInstanceType,
			serviceInstanceId:   serviceInstanceId,
		}

		d.SetId(id.compute())

		return nil
	}
}

func resourceYandexBillingServiceInstanceBindingRead(serviceInstanceType string, idFieldName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)

		id, err := parseBindServiceInstanceId(d.Id())

		if err != nil {
			return diag.FromErr(err)
		}

		return tryFindInstanceBindingById(serviceInstanceType, idFieldName)(id, config, d)
	}
}

func resourceYandexBillingServiceInstanceBindingUpdate(serviceInstanceType string, idFieldName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceYandexBillingServiceInstanceBindingCreate(serviceInstanceType, idFieldName)
}

func resourceYandexBillingServiceInstanceBindingDelete(serviceInstanceType string, idFieldName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		log.Printf("[INFO] The resource of binding to billign account is deleted however the binding itself still exists. " +
			"This is an excepted behaviour. " +
			"See documentation for details.")

		d.SetId("")

		return nil
	}
}

func dataSourceYandexBillingServiceInstanceBindingRead(serviceInstanceType string, idFieldName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)

		billingAccountId := d.Get("billing_account_id").(string)
		serviceInstanceId := d.Get(idFieldName).(string)

		id := bindServiceInstanceId{
			billingAccountId:    billingAccountId,
			serviceInstanceType: serviceInstanceType,
			serviceInstanceId:   serviceInstanceId,
		}

		return tryFindInstanceBindingById(serviceInstanceType, idFieldName)(&id, config, d)
	}
}

func tryFindInstanceBindingById(serviceInstanceType string, idFieldName string) func(*bindServiceInstanceId, *Config, *schema.ResourceData) diag.Diagnostics {
	return func(id *bindServiceInstanceId, config *Config, d *schema.ResourceData) diag.Diagnostics {
		req := billing.ListBillableObjectBindingsRequest{
			BillingAccountId: id.billingAccountId,
		}

		it := config.sdk.Billing().BillingAccount().BillingAccountBillableObjectBindingsIterator(
			config.Context(),
			&req,
		)

		for it.Next() {
			billableObject := it.Value().BillableObject

			if serviceInstanceType == billableObject.Type && id.serviceInstanceId == billableObject.Id {
				err := d.Set("billing_account_id", id.billingAccountId)
				if err != nil {
					return diag.FromErr(fmt.Errorf("Unable to set billing_account_id=%s", id.billingAccountId))
				}

				err = d.Set(idFieldName, id.serviceInstanceId)
				if err != nil {
					return diag.FromErr(fmt.Errorf("Unable to set %s=%s", idFieldName, id.serviceInstanceId))
				}

				d.SetId(id.compute())

				return nil
			}
		}

		d.SetId("")

		return diag.FromErr(fmt.Errorf("Bound %s to billing account not found", serviceInstanceType))
	}
}

type bindServiceInstanceId struct {
	billingAccountId    string
	serviceInstanceType string
	serviceInstanceId   string
}

func (id bindServiceInstanceId) compute() string {
	return id.billingAccountId + "/" + id.serviceInstanceType + "/" + id.serviceInstanceId
}

func parseBindServiceInstanceId(s string) (*bindServiceInstanceId, error) {
	splitId := strings.Split(s, "/")

	if len(splitId) != 3 {
		return nil, fmt.Errorf("unexcepted Id format occurred while parsing bindServiceInstanceId")
	}

	id := bindServiceInstanceId{
		billingAccountId:    splitId[0],
		serviceInstanceType: splitId[1],
		serviceInstanceId:   splitId[2],
	}

	return &id, nil
}
