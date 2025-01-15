package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"cosmossdk.io/math"
	cosmossdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/sentinel-official/hub/v12/types/v1"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/v2ray"
	"github.com/sentinel-official/sentinel-go-sdk/wireguard"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/database/models"
	"github.com/sentinel-official/dvpn-node/database/operations"
)

func HandlerAddSession(c *context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: validate current peer count

		req, err := NewRequestAddSession(ctx)
		if err != nil {
			err = fmt.Errorf("invalid request format: %w", err)
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(2, err))
			return
		}

		query := map[string]interface{}{
			"id": req.URI.ID,
		}

		record, err := operations.SessionFindOne(c.Database(), query)
		if err != nil {
			err = fmt.Errorf("failed to retrieve session from database for id %d: %w", req.URI.ID, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(3, err))
			return
		}
		if record != nil {
			err = fmt.Errorf("session already exists for id %d", req.URI.ID)
			ctx.JSON(http.StatusConflict, types.NewResponseError(3, err))
			return
		}

		var peerKey string
		if c.Service().Type() == types.ServiceTypeV2Ray {
			var r v2ray.AddPeerRequest
			if err := json.Unmarshal(req.Data, &r); err != nil {
				err = fmt.Errorf("failed to decode v2ray add peer request: %w", err)
				ctx.JSON(http.StatusBadRequest, types.NewResponseError(4, err))
				return
			}

			peerKey = r.Key()
		}
		if c.Service().Type() == types.ServiceTypeWireGuard {
			var r wireguard.AddPeerRequest
			if err := json.Unmarshal(req.Data, &r); err != nil {
				err = fmt.Errorf("failed to decode wireguard add peer request: %w", err)
				ctx.JSON(http.StatusBadRequest, types.NewResponseError(4, err))
				return
			}

			peerKey = r.Key()
		}

		query = map[string]interface{}{
			"peer_key": peerKey,
		}

		record, err = operations.SessionFindOne(c.Database(), query)
		if err != nil {
			err = fmt.Errorf("failed to retrieve session from database for peer_key %s: %w", peerKey, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(5, err))
			return
		}
		if record != nil {
			err = fmt.Errorf("session already exists for peer_key %s", peerKey)
			ctx.JSON(http.StatusConflict, types.NewResponseError(5, err))
			return
		}

		session, err := c.Client().Session(ctx, req.URI.ID)
		if err != nil {
			err = fmt.Errorf("failed to query session %d from blockchain: %w", req.URI.ID, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(6, err))
			return
		}
		if session == nil {
			err = fmt.Errorf("session %d does not exist", req.URI.ID)
			ctx.JSON(http.StatusNotFound, types.NewResponseError(6, err))
			return
		}
		if !session.GetStatus().Equal(v1.StatusActive) {
			err = fmt.Errorf("invalid session status; got %s, expected %s", session.GetStatus(), v1.StatusActive)
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(6, err))
			return
		}
		if session.GetNodeAddress() != c.NodeAddr().String() {
			err = fmt.Errorf("node address mismatch: got %s, expected %s", session.GetNodeAddress(), c.NodeAddr())
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(6, err))
			return
		}

		accAddr, err := cosmossdk.AccAddressFromBech32(session.GetAccAddress())
		if err != nil {
			err = fmt.Errorf("invalid account address format: %w", err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(7, err))
			return
		}

		account, err := c.Client().Account(ctx, accAddr)
		if err != nil {
			err = fmt.Errorf("failed to query account %s: %w", accAddr, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(7, err))
			return
		}
		if account == nil {
			err = fmt.Errorf("account %s does not exist", accAddr)
			ctx.JSON(http.StatusNotFound, types.NewResponseError(7, err))
			return
		}
		if account.GetPubKey() == nil {
			err = fmt.Errorf("public key for account %s does not exist", accAddr)
			ctx.JSON(http.StatusNotFound, types.NewResponseError(7, err))
			return
		}

		if ok := account.GetPubKey().VerifySignature(req.Msg(), req.Signature); !ok {
			err = errors.New("signature verification failed")
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(7, err))
			return
		}

		var data interface{}
		if c.Service().Type() == types.ServiceTypeV2Ray {
			var r v2ray.AddPeerRequest
			if err := json.Unmarshal(req.Data, &r); err != nil {
				err = fmt.Errorf("failed to decode v2ray add peer request: %w", err)
				ctx.JSON(http.StatusBadRequest, types.NewResponseError(8, err))
				return
			}

			data, err = c.Service().AddPeer(ctx, &r)
			if err != nil {
				err = fmt.Errorf("failed to add v2ray peer: %w", err)
				ctx.JSON(http.StatusInternalServerError, types.NewResponseError(8, err))
				return
			}
		}
		if c.Service().Type() == types.ServiceTypeWireGuard {
			var r wireguard.AddPeerRequest
			if err := json.Unmarshal(req.Data, &r); err != nil {
				err = fmt.Errorf("failed to decode wireguard add peer request: %w", err)
				ctx.JSON(http.StatusBadRequest, types.NewResponseError(8, err))
				return
			}

			data, err = c.Service().AddPeer(ctx, &r)
			if err != nil {
				err = fmt.Errorf("failed to add wireguard peer: %w", err)
				ctx.JSON(http.StatusInternalServerError, types.NewResponseError(8, err))
				return
			}
		}

		item := models.NewSession().
			WithID(session.GetID()).
			WithAccAddr(accAddr).
			WithNodeAddr(c.NodeAddr()).
			WithDownloadBytes(math.ZeroInt()).
			WithUploadBytes(math.ZeroInt()).
			WithMaxBytes(session.GetMaxBytes()).
			WithDuration(0).
			WithMaxDuration(session.GetMaxDuration()).
			WithSignature(nil).
			WithPeerKey(peerKey).
			WithServiceType(c.Service().Type())

		if err := operations.SessionInsertOne(c.Database(), item); err != nil {
			err = fmt.Errorf("failed to insert session into database: %w", err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(9, err))
			return
		}

		res := &ResultAddSession{
			Addrs: c.RemoteAddrs(),
			Data:  data,
		}

		ctx.JSON(http.StatusOK, types.NewResponseResult(res))
	}
}
