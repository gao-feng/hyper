package system

import (
	"github.com/docker/engine-api/types"
	"github.com/hyperhq/hyper/engine"
)

// Backend is the methods that need to be implemented to provide
// system specific functionality.
type Backend interface {
	CmdSystemInfo() (*engine.Env, error)
	CmdSystemVersion() *engine.Env
	CmdAuthenticateToRegistry(authConfig *types.AuthConfig) (*engine.Env, error)
}
