/*
Package sputnik simplifies creation of satellite/sidecar  processes for servers.

# Modular monolith

sputnik forces modular-monolith design:
  - process created as set of independent asynchronously running blocks

# Satellites blueprint

sputnik supports common for all satellite processes functionality:
  - ordered blocks initialization
  - convenient negotiation between blocks
  - server heartbeat
  - graceful shutdown

# Eats own dog food

sputnik itself consists of 3 blocks:
  - "initiator" - dispatcher of all blocks
  - "finisher"  - listener of external shutdown/exit events
  - "connector" - connects/reconnects with server, provides this
    information to another blocks

# Less You Know, the Better You Sleep

sputnik knows nothing about internals of the process.
It only assumes that server configuration and connection
are required for functioning.
This is the reason to define server connection and
configuration as any.
Use type assertions for "casting" to concrete interface/implementation.
*/
package sputnik
