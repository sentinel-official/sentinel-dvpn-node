package node

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/sentinel-official/hub/v12/x/node/types/v3"
	"github.com/sentinel-official/sentinel-go-sdk/libs/cron"
	"github.com/sentinel-official/sentinel-go-sdk/libs/log"

	nodecontext "github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/utils"
)

// Node represents the application node, holding its context, router, and scheduler.
type Node struct {
	router     *gin.Engine     // HTTP router for handling API requests.
	scheduler  *cron.Scheduler // Scheduler for managing periodic tasks.
	listenAddr string          // Address the Node listens on for incoming requests.
}

// New creates a new Node with the provided context.
func New() *Node {
	return &Node{}
}

// WithListenAddr sets the listen address for the Node and returns the updated Node.
func (n *Node) WithListenAddr(v string) *Node {
	n.listenAddr = v
	return n
}

// WithRouter sets the router for the Node and returns the updated Node.
func (n *Node) WithRouter(v *gin.Engine) *Node {
	n.router = v
	return n
}

// WithScheduler sets the scheduler for the Node and returns the updated Node.
func (n *Node) WithScheduler(v *cron.Scheduler) *Node {
	n.scheduler = v
	return n
}

// ListenAddr returns the listen address of the Node.
func (n *Node) ListenAddr() string {
	return n.listenAddr
}

// Router returns the router configured for the Node.
func (n *Node) Router() *gin.Engine {
	return n.router
}

// Scheduler returns the scheduler configured for the Node.
func (n *Node) Scheduler() *cron.Scheduler {
	return n.scheduler
}

// Register registers the node on the network if not already registered.
func (n *Node) Register(c *nodecontext.Context) error {
	node, err := c.Client().Node(context.TODO(), c.AccAddr().Bytes())
	if err != nil {
		return fmt.Errorf("failed to query node: %w", err)
	}
	if node != nil {
		return nil
	}

	log.Info("Registering node...")

	// Prepare a message to register the node.
	msg := v3.NewMsgRegisterNodeRequest(
		c.AccAddr().Bytes(),
		c.GigabytePrices(),
		c.HourlyPrices(),
		c.RemoteAddrs()[0],
	)

	// Broadcast the registration transaction.
	res, err := c.BroadcastTx(context.TODO(), msg)
	if err != nil {
		return fmt.Errorf("failed to broadcast register node tx: %w", err)
	}
	if !res.TxResult.IsOK() {
		err := errors.New(res.TxResult.Log)
		return fmt.Errorf("register node tx failed with code %d: %w", res.TxResult.Code, err)
	}

	return nil
}

// UpdateDetails updates the node's pricing and address details on the network.
func (n *Node) UpdateDetails(c *nodecontext.Context) error {
	log.Info("Updating node details...")

	// Prepare a message to update the node's details.
	msg := v3.NewMsgUpdateNodeDetailsRequest(
		c.AccAddr().Bytes(),
		c.GigabytePrices(),
		c.HourlyPrices(),
		c.RemoteAddrs()[0],
	)

	// Broadcast the update transaction.
	res, err := c.BroadcastTx(context.TODO(), msg)
	if err != nil {
		return fmt.Errorf("failed to broadcast update node details tx: %w", err)
	}
	if !res.TxResult.IsOK() {
		err := errors.New(res.TxResult.Log)
		return fmt.Errorf("update node deatils tx failed with code %d: %w", res.TxResult.Code, err)
	}

	return nil
}

// Start initializes the Node's services, scheduler, and HTTPS server.
func (n *Node) Start(c *nodecontext.Context, errChan chan error) error {
	log.Info("Starting node...")

	go func() {
		// Bring up the service by running pre-defined tasks.
		if err := c.Service().Up(context.TODO()); err != nil {
			errChan <- fmt.Errorf("failed to run service up task: %w", err)
			return
		}
		if err := c.Service().PostUp(); err != nil {
			errChan <- fmt.Errorf("failed to run service post-up task: %w", err)
			return
		}
	}()

	go func() {
		// Register the node and update its details.
		if err := n.Register(c); err != nil {
			errChan <- fmt.Errorf("failed to register node: %w", err)
			return
		}
		if err := n.UpdateDetails(c); err != nil {
			errChan <- fmt.Errorf("failed to update node details: %w", err)
			return
		}

		// Start the cron scheduler to execute periodic tasks.
		if err := n.Scheduler().Start(); err != nil {
			errChan <- fmt.Errorf("failed to start scheduler: %w", err)
			return
		}

		// Define paths for TLS certificate and key files.
		certPath := filepath.Join(c.HomeDir(), "tls.crt")
		keyPath := filepath.Join(c.HomeDir(), "tls.key")

		// Start the HTTPS server using the configured TLS certificates and router.
		if err := utils.ListenAndServeTLS(n.ListenAddr(), certPath, keyPath, n.Router()); err != nil {
			errChan <- fmt.Errorf("failed to listen and serve tls: %w", err)
			return
		}
	}()

	return nil
}

// Stop gracefully stops the Node's operations.
func (n *Node) Stop(c *nodecontext.Context) error {
	if err := c.Service().PreDown(); err != nil {
		return fmt.Errorf("failed to run service pre-down task: %w", err)
	}
	if err := c.Service().Down(context.TODO()); err != nil {
		return fmt.Errorf("failed to run service down task: %w", err)
	}
	if err := c.Service().PostDown(); err != nil {
		return fmt.Errorf("failed to run service post-up task: %w", err)
	}

	return nil
}
