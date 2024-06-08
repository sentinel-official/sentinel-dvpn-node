package lite

import (
	"github.com/avast/retry-go/v4"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (c *Client) broadcastTx(remote string, txBytes []byte) (*sdk.TxResponse, error) {
	c.log.Debug("Broadcasting the transaction", "remote", remote, "size", len(txBytes))

	client, err := c.GetHttpClient(remote, c.txTimeout)
	if err != nil {
		return nil, err
	}

	ctx := c.ctx.WithClient(client)

	resp, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	switch resp.Code {
	case abcitypes.CodeTypeOK:
		return resp, nil
	case sdkerrors.ErrTxInMempoolCache.ABCICode():
		return resp, nil
	default:
		return nil, errors.New(resp.RawLog)
	}
}

func (c *Client) BroadcastTx(txBytes []byte) (res *sdk.TxResponse, err error) {
	defer func() {
		if err != nil {
			c.log.Error("failed to broadcast the transaction", "error", err)
		}
	}()

	for i := 0; i < len(c.remotes); i++ {
		res, err = c.broadcastTx(c.remotes[i], txBytes)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) calculateGas(remote string, txf tx.Factory, messages ...sdk.Msg) (uint64, error) {
	c.log.Debug("Calculating the gas", "remote", remote, "messages", len(messages))

	client, err := c.GetHttpClient(remote, c.txTimeout)
	if err != nil {
		return 0, err
	}

	ctx := c.ctx.WithClient(client)

	_, gas, err := tx.CalculateGas(ctx, txf, messages...)
	if err != nil {
		return 0, err
	}

	return gas, nil
}

func (c *Client) CalculateGas(txf tx.Factory, messages ...sdk.Msg) (gas uint64, err error) {
	for i := 0; i < len(c.remotes); i++ {
		gas, err = c.calculateGas(c.remotes[i], txf, messages...)
		if err == nil {
			break
		}
	}

	if err != nil {
		return 0, err
	}

	return gas, nil
}

func (c *Client) PrepareTxFactory(messages ...sdk.Msg) (txf tx.Factory, err error) {
	defer func() {
		if err != nil {
			c.log.Error("failed to prepare the transaction", "error", err)
		}
	}()

	acc, err := c.QueryAccount(c.FromAddress())
	if err != nil {
		return txf, err
	}

	txf = c.txf.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence())

	if c.SimulateAndExecute() {
		gas, err := c.CalculateGas(txf, messages...)
		if err != nil {
			return txf, err
		}

		txf = txf.WithGas(gas)
	}

	return txf, nil
}

func (c *Client) tx(messages ...sdk.Msg) (res *sdk.TxResponse, err error) {
	c.log.Info("Preparing the transaction", "messages", len(messages))
	txf, err := c.PrepareTxFactory(messages...)
	if err != nil {
		return nil, err
	}

	c.log.Info("Transaction info", "gas", txf.Gas(), "sequence", txf.Sequence())
	txb, err := tx.BuildUnsignedTx(txf, messages...)
	if err != nil {
		return nil, err
	}

	if err = tx.Sign(txf, c.FromName(), txb, true); err != nil {
		return nil, err
	}

	txBytes, err := c.TxConfig().TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, err
	}

	c.log.Info("Broadcasting the transaction", "size", len(txBytes))
	res, err = c.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) Tx(messages ...sdk.Msg) (res *sdk.TxResponse, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err = retry.Do(
		func() error {
			res, err = c.tx(messages...)
			if err != nil {
				return err
			}

			c.log.Info("Transaction result", "code", res.Code,
				"codespace", res.Codespace, "height", res.Height, "tx_hash", res.TxHash)
			return nil
		},
		retry.Attempts(5),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}
