package session

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
)

type RequestAddSession struct {
	Data      []byte
	Signature []byte

	URI struct {
		ID uint64 `uri:"id" binding:"gt=0"`
	}
	Body struct {
		Data      string `json:"data"`
		Signature string `json:"signature"`
	}
}

func (r *RequestAddSession) Msg() []byte {
	return append(types.Uint64ToBigEndian(r.URI.ID), r.Data...)
}

func NewRequestAddSession(c *gin.Context) (req *RequestAddSession, err error) {
	req = &RequestAddSession{}
	if err = c.ShouldBindUri(&req.URI); err != nil {
		return nil, err
	}
	if err = c.ShouldBindJSON(&req.Body); err != nil {
		return nil, err
	}

	if req.Body.Data == "" {
		return nil, errors.New("data cannot be empty")
	}
	if req.Body.Signature == "" {
		return nil, errors.New("signature cannot be empty")
	}

	req.Data, err = base64.StdEncoding.DecodeString(req.Body.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %s", err)
	}

	req.Signature, err = base64.StdEncoding.DecodeString(req.Body.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %s", err)
	}

	return req, nil
}
