package rc

import (
	"context"
	"fmt"

	"github.com/rclone/rclone/fs/rc"
	"github.com/rclone/rclone/fs/rc/rcserver"
	"k8s.io/klog/v2"
)

// Options wraps the subset of rclone RC options we expose via the CSI driver.
type Options struct {
	Enabled  bool
	Address  string
	NoAuth   bool
	Username string
	Password string
}

// Validate validates the options
func (o *Options) Validate() error {
	if o == nil {
		return fmt.Errorf("options cannot be nil")
	}
	if o.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	if o.NoAuth && (o.Username != "" || o.Password != "") {
		return fmt.Errorf("basic auth credentials must be provided unless rc.noAuth=true")
	}

	if !o.NoAuth && (o.Username == "" || o.Password == "") {
		return fmt.Errorf("rc basic auth credentials must be provided unless rc.noAuth=true")
	}

	return nil
}

// Server interface for RC server lifecycle management
type Server interface {
	Shutdown() error
}

// NewOptions creates a new Options instance
func NewOptions() *Options {
	return &Options{
		Enabled:  false,
		Address:  ":5573",
		NoAuth:   false,
		Username: "",
		Password: "",
	}
}

// Start starts the rclone Remote Control server if enabled
func Start(ctx context.Context, opts *Options) (Server, error) {
	if opts == nil || !opts.Enabled {
		return nil, nil
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	rcOpts := rc.Opt
	rcOpts.Enabled = true
	rcOpts.NoAuth = opts.NoAuth
	if opts.Address != "" {
		rcOpts.HTTP.ListenAddr = []string{opts.Address}
	}
	if !opts.NoAuth {
		rcOpts.Auth.BasicUser = opts.Username
		rcOpts.Auth.BasicPass = opts.Password
	}

	klog.Infof("Starting rclone RC server on %s", opts.Address)
	return rcserver.Start(ctx, &rcOpts)
}
