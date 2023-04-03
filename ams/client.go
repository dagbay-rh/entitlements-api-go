package ams

import (
	"context"
	"fmt"

	"github.com/RedHatInsights/entitlements-api-go/config"
	sdk "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
)

type AMSInterface interface {
	GetQuotaCost(organizationId string) (*v1.QuotaCost, error)
	GetSubscription(subscriptionId string) (*v1.Subscription, error)
	GetSubscriptions(size, page int) (*v1.SubscriptionList, error)
	DeleteSubscription(subscriptionId string) error
	QuotaAuthorization(accountUsername, quotaVersion string) (*v1.QuotaAuthorizationsPostResponse, error)
}

var _ AMSInterface = &TestClient{}

type TestClient struct{}

func (c *TestClient) GetQuotaCost(organizationId string) (*v1.QuotaCost, error) {
	quotaCost, err := v1.NewQuotaCost().QuotaID("seat|ansible.wisdom").Build()
	if err != nil {
		return nil, err
	}
	return quotaCost, nil
}

func (c *TestClient) GetSubscription(subscriptionId string) (*v1.Subscription, error) {
	if subscriptionId == "" {
		return nil, fmt.Errorf("subscriptionId cannot be an empty string")
	}
	subscription, err := v1.NewSubscription().
		ID(subscriptionId).
		OrganizationID("4384938490324").
		Build()
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func (c *TestClient) DeleteSubscription(subscriptionId string) error {
	return nil
}

// TODO: waiting on updates to the ocm sdk
func (c *TestClient) QuotaAuthorization(accountUsername, quotaVersion string) (*v1.QuotaAuthorizationsPostResponse, error) {
	return nil, nil
}

func (c *TestClient) GetSubscriptions(size, page int) (*v1.SubscriptionList, error) {
	lst, err := v1.NewSubscriptionList().
		Items(
			v1.NewSubscription().
				Creator(v1.NewAccount().Username("testuser")).
				Plan(v1.NewPlan().Type("AnsibleWisdom").Name("AnsibleWisdom")),
		).Build()
	if err != nil {
		return nil, err
	}
	return lst, nil
}

var _ AMSInterface = &Client{}

type Client struct {
	client *sdk.Connection
}

func NewClient() (*Client, error) {

	logger, err := logging.NewGoLoggerBuilder().Debug(false).Build()
	if err != nil {
		return nil, err
	}

	cfg := config.GetConfig()

	clientId := cfg.Options.GetString(config.Keys.ClientID)
	secret := cfg.Options.GetString(config.Keys.ClientSecret)
	tokenUrl := cfg.Options.GetString(config.Keys.TokenURL)
	amsUrl := cfg.Options.GetString(config.Keys.AMSHost)

	client, err := sdk.NewConnectionBuilder().
		Logger(logger).
		Client(clientId, secret).
		TokenURL(tokenUrl).
		URL(amsUrl).
		BuildContext(context.Background())

	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, err
}

func (c *Client) GetQuotaCost(organizationId string) (*v1.QuotaCost, error) {
	resp, err := c.client.AccountsMgmt().V1().Organizations().Organization(organizationId).QuotaCost().List().Search(
		"quota_id LIKE 'seat|ansible.wisdom%'",
	).Send()
	if err != nil {
		return nil, err
	}
	return resp.Items().Get(0), nil
}

func (c *Client) GetSubscription(subscriptionId string) (*v1.Subscription, error) {
	resp, err := c.client.AccountsMgmt().V1().Subscriptions().Subscription(subscriptionId).Get().Send()
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (c *Client) GetSubscriptions(size, page int) (*v1.SubscriptionList, error) {
	req := c.client.AccountsMgmt().V1().Subscriptions().List().
		Search("quota_id LIKE 'seat|ansible.wisdom'%").
		Size(size).
		Page(page)

	resp, err := req.Send()
	if err != nil {
		return nil, err
	}
	return resp.Items(), nil
}

func (c *Client) DeleteSubscription(subscriptionId string) error {
	_, err := c.client.AccountsMgmt().V1().Subscriptions().Subscription(subscriptionId).Delete().Send()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) QuotaAuthorization(accountUsername, quotaVersion string) (*v1.QuotaAuthorizationsPostResponse, error) {

	rr := v1.NewReservedResource().
		ResourceName("ansible.wisdom").
		ResourceType("seat")

	req, err := v1.NewQuotaAuthorizationRequest().
		AccountUsername(accountUsername).
		Reserve(true).
		ProductID("AnsibleWisdom").
		Resources(rr).
		QuotaVersion(quotaVersion).
		Build()

	if err != nil {
		return nil, err
	}
	return c.client.AccountsMgmt().V1().QuotaAuthorizations().Post().Request(req).Send()
}
