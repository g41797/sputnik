// Package sputnik simplifies creation of satellite/sidecar  processes for servers.
//
// # Modular monolith
//
// sputnik forces modular-monolith design:
//   - process created as set of independent asynchronously running blocks
//
// # Satellites blueprint
//
// sputnik supports common for all satellite processes functionality:
//   - ordered blocks initialization
//   - convenient negotiation between blocks
//   - server heartbeat
//   - graceful shutdown
//
// # Eats own dog food
//
// sputnik itself consists of 2 blocks:
//   - "initiator" - dispatcher of all blocks
//   - "finisher" - listener of external shutdown/exit events

//
// # Less You Know, the Better You Sleep
//
// sputnik knows nothing about internals of the process.
// It only assumes that server connection, configuration and logger
// are required for functioning.
// This is the reason to define server connection, logger and
// configuration as any.
// Use type assertions for "casting" to concrete interface/implementation:
// .............
// logger, ok := lp.(*log.Logger)
//
// if ok {
// 	logger.Println(....)
// }
// .............

package sputnik
