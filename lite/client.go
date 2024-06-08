package lite

import (
	"fmt"

	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	jsonrpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

type Client struct {
	ctx          client.Context
	log          tmlog.Logger
	mutex        *sync.Mutex
	queryTimeout uint
	remotes      []string
	txf          tx.Factory
	txTimeout    uint
}

func NewClient() *Client {
	return &Client{
		mutex: &sync.Mutex{},
	}
}

func NewDefaultClient() *Client {
	var (
		cfg = DefaultEncodingConfig()
		ctx = client.Context{}.
			WithBroadcastMode(flags.BroadcastBlock).
			WithCodec(cfg.Codec).
			WithInterfaceRegistry(cfg.InterfaceRegistry).
			WithLegacyAmino(cfg.Amino).
			WithOutputFormat("text").
			WithOutput(io.Discard).
			WithSkipConfirmation(true)
	)

	return NewClient().
		WithContext(ctx).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithTxConfig(cfg.TxConfig)
}

func (c *Client) WithContext(v client.Context) *Client {
	c.ctx = v
	return c
}

func (c *Client) WithLogger(v tmlog.Logger) *Client {
	c.log = v
	return c
}

func (c *Client) WithQueryTimeout(v uint) *Client {
	c.queryTimeout = v
	return c
}

func (c *Client) WithRemotes(v []string) *Client {
	c.remotes = v
	return c
}

func (c *Client) WithTxTimeout(v uint) *Client {
	c.txTimeout = v
	return c
}

func (c *Client) WithAccountRetriever(v client.AccountRetriever) *Client {
	c.ctx = c.ctx.WithAccountRetriever(v)
	c.txf = c.txf.WithAccountRetriever(v)
	return c
}

func (c *Client) WithChainID(v string) *Client {
	c.ctx = c.ctx.WithChainID(v)
	c.txf = c.txf.WithChainID(v)
	return c
}

func (c *Client) WithFeeGranterAddress(v sdk.AccAddress) *Client {
	c.ctx = c.ctx.WithFeeGranterAddress(v)
	return c
}

func (c *Client) WithFromAddress(v sdk.AccAddress) *Client {
	c.ctx = c.ctx.WithFromAddress(v)
	return c
}

func (c *Client) WithFromName(v string) *Client {
	c.ctx = c.ctx.WithFromName(v)
	return c
}

func (c *Client) WithGas(v uint64) *Client {
	c.txf = c.txf.WithGas(v)
	return c
}

func (c *Client) WithGasAdjustment(v float64) *Client {
	c.txf = c.txf.WithGasAdjustment(v)
	return c
}

func (c *Client) WithGasPrices(v string) *Client {
	c.txf = c.txf.WithGasPrices(v)
	return c
}

func (c *Client) WithKeyring(v keyring.Keyring) *Client {
	c.ctx = c.ctx.WithKeyring(v)
	c.txf = c.txf.WithKeybase(v)
	return c
}

func (c *Client) WithSignModeStr(v string) *Client {
	m := signing.SignMode_SIGN_MODE_UNSPECIFIED
	switch v {
	case flags.SignModeDirect:
		m = signing.SignMode_SIGN_MODE_DIRECT
	case flags.SignModeLegacyAminoJSON:
		m = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	}

	c.ctx = c.ctx.WithSignModeStr(v)
	c.txf = c.txf.WithSignMode(m)
	return c
}

func (c *Client) WithSimulateAndExecute(v bool) *Client {
	c.txf = c.txf.WithSimulateAndExecute(v)
	return c
}

func (c *Client) WithTxConfig(v client.TxConfig) *Client {
	c.ctx = c.ctx.WithTxConfig(v)
	c.txf = c.txf.WithTxConfig(v)
	return c
}

func (c *Client) FromAddress() sdk.AccAddress { return c.ctx.FromAddress }
func (c *Client) FromName() string            { return c.ctx.FromName }
func (c *Client) SimulateAndExecute() bool    { return c.txf.SimulateAndExecute() }
func (c *Client) TxConfig() client.TxConfig   { return c.ctx.TxConfig }

type AddHeaderTransport struct {
	T http.RoundTripper
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	verInfo := version.NewInfo()

	req.Header.Add("User-Agent", fmt.Sprintf("sentinel-dvpn-node/v%s", verInfo.Version))
	return adt.T.RoundTrip(req)
}

func (c *Client) GetHttpClient(remote string, timeout uint) (*rpchttp.HTTP, error) {
	httpClient, err := jsonrpcclient.DefaultHTTPClient(remote)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = time.Duration(timeout) * time.Second
	httpClient.Transport = &AddHeaderTransport{T: httpClient.Transport}

	return rpchttp.NewWithClient(remote, "/websocket", httpClient)
}
