package lite

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	hubtypes "github.com/sentinel-official/hub/types"
	nodetypes "github.com/sentinel-official/hub/x/node/types"
	sessiontypes "github.com/sentinel-official/hub/x/session/types"
	subscriptiontypes "github.com/sentinel-official/hub/x/subscription/types"
	vpntypes "github.com/sentinel-official/hub/x/vpn/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (c *Client) queryAccount(remote string, accAddr sdk.AccAddress) (authtypes.AccountI, error) {
	c.log.Debug("Querying the account", "remote", remote, "address", accAddr)

	client, err := c.GetHttpClient(remote, c.queryTimeout)
	if err != nil {
		return nil, err
	}

	var (
		ctx = c.ctx.WithClient(client)
		qc  = authtypes.NewQueryClient(ctx)
	)

	resp, err := qc.Account(
		context.TODO(),
		&authtypes.QueryAccountRequest{
			Address: accAddr.String(),
		},
	)
	if err != nil {
		return nil, types.QueryError(err)
	}

	var result authtypes.AccountI
	if err = c.ctx.InterfaceRegistry.UnpackAny(resp.Account, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) QueryAccount(accAddr sdk.AccAddress) (result authtypes.AccountI, err error) {
	c.log.Info("Querying the account", "address", accAddr)
	for i := 0; i < len(c.remotes); i++ {
		result, err = c.queryAccount(c.remotes[i], accAddr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) queryNode(remote string, nodeAddr hubtypes.NodeAddress) (*nodetypes.Node, error) {
	c.log.Debug("Querying the node", "remote", remote, "address", nodeAddr)

	client, err := c.GetHttpClient(remote, c.queryTimeout)
	if err != nil {
		return nil, err
	}

	var (
		ctx = c.ctx.WithClient(client)
		qc  = nodetypes.NewQueryServiceClient(ctx)
	)

	res, err := qc.QueryNode(
		context.TODO(),
		nodetypes.NewQueryNodeRequest(nodeAddr),
	)
	if err != nil {
		return nil, types.QueryError(err)
	}

	return &res.Node, nil
}

func (c *Client) QueryNode(nodeAddr hubtypes.NodeAddress) (result *nodetypes.Node, err error) {
	c.log.Info("Querying the node", "address", nodeAddr)
	for i := 0; i < len(c.remotes); i++ {
		result, err = c.queryNode(c.remotes[i], nodeAddr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) querySubscription(remote string, id uint64) (subscriptiontypes.Subscription, error) {
	c.log.Debug("Querying the subscription", "remote", remote, "id", id)

	client, err := c.GetHttpClient(remote, c.queryTimeout)
	if err != nil {
		return nil, err
	}

	var (
		ctx = c.ctx.WithClient(client)
		qc  = subscriptiontypes.NewQueryServiceClient(ctx)
	)

	res, err := qc.QuerySubscription(
		context.TODO(),
		subscriptiontypes.NewQuerySubscriptionRequest(id),
	)
	if err != nil {
		return nil, types.QueryError(err)
	}

	var result subscriptiontypes.Subscription
	if err = c.ctx.InterfaceRegistry.UnpackAny(res.Subscription, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) QuerySubscription(id uint64) (result subscriptiontypes.Subscription, err error) {
	c.log.Info("Querying the subscription", "id", id)
	for i := 0; i < len(c.remotes); i++ {
		result, err = c.querySubscription(c.remotes[i], id)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) queryAllocation(remote string, id uint64, accAddr sdk.AccAddress) (*subscriptiontypes.Allocation, error) {
	c.log.Debug("Querying the allocation", "remote", remote, "id", id, "address", accAddr)

	client, err := c.GetHttpClient(remote, c.queryTimeout)
	if err != nil {
		return nil, err
	}

	var (
		ctx = c.ctx.WithClient(client)
		qc  = subscriptiontypes.NewQueryServiceClient(ctx)
	)

	res, err := qc.QueryAllocation(
		context.TODO(),
		subscriptiontypes.NewQueryAllocationRequest(id, accAddr),
	)
	if err != nil {
		return nil, types.QueryError(err)
	}

	return &res.Allocation, nil
}

func (c *Client) QueryAllocation(id uint64, accAddr sdk.AccAddress) (result *subscriptiontypes.Allocation, err error) {
	c.log.Info("Querying the allocation", "id", id, "address", accAddr)
	for i := 0; i < len(c.remotes); i++ {
		result, err = c.queryAllocation(c.remotes[i], id, accAddr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) querySession(remote string, id uint64) (*sessiontypes.Session, error) {
	c.log.Debug("Querying the session", "remote", remote, "id", id)

	client, err := c.GetHttpClient(remote, c.queryTimeout)
	if err != nil {
		return nil, err
	}

	var (
		ctx = c.ctx.WithClient(client)
		qc  = sessiontypes.NewQueryServiceClient(ctx)
	)

	res, err := qc.QuerySession(
		context.TODO(),
		sessiontypes.NewQuerySessionRequest(id),
	)
	if err != nil {
		return nil, types.QueryError(err)
	}

	return &res.Session, nil
}

func (c *Client) QuerySession(id uint64) (result *sessiontypes.Session, err error) {
	c.log.Info("Querying the session", "id", id)
	for i := 0; i < len(c.remotes); i++ {
		result, err = c.querySession(c.remotes[i], id)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) hasNodeForPlan(remote string, id uint64, nodeAddr hubtypes.NodeAddress) (bool, error) {
	client, err := c.GetHttpClient(remote, c.queryTimeout)
	if err != nil {
		return false, err
	}

	ctx := c.ctx.WithClient(client)

	value, _, err := ctx.QueryStore(
		append(
			[]byte(nodetypes.ModuleName+"/"),
			nodetypes.NodeForPlanKey(id, nodeAddr)...,
		),
		vpntypes.ModuleName,
	)
	if err != nil {
		return false, err
	}
	if value == nil {
		return false, nil
	}

	return true, nil
}

func (c *Client) HasNodeForPlan(id uint64, nodeAddr hubtypes.NodeAddress) (result bool, err error) {
	for i := 0; i < len(c.remotes); i++ {
		result, err = c.hasNodeForPlan(c.remotes[i], id, nodeAddr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return false, err
	}

	return result, nil
}
