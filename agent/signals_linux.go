// Copyright © 2017 Joyent, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build linux

package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

func (a *Agent) setupSignals() {
	// Handle shutdown via a.shutdownCtx
	a.signalCh = make(chan os.Signal, 10)
	signal.Notify(a.signalCh, os.Interrupt, unix.SIGTERM, unix.SIGHUP, unix.SIGPIPE)

	a.shutdownCtx, a.shutdown = context.WithCancel(context.Background())
	a.pgConnCtx, a.pgConnShutdown = context.WithCancel(a.shutdownCtx)
}

// handleSignals runs the signal handler thread
func (a *Agent) handleSignals() {
	for {
		select {
		case <-a.shutdownCtx.Done():
			log.Debug().Msg("Shutting down")
			return
		case sig := <-a.signalCh:
			log.Info().Str("signal", sig.String()).Msg("Received signal")
			switch sig {
			case os.Interrupt, unix.SIGTERM:
				a.shutdown()
			case unix.SIGPIPE, unix.SIGHUP:
				// Noop
			default:
				panic(fmt.Sprintf("unsupported signal: %v", sig))
			}
		}
	}
}
