# Distributed Key-Value Store with Raft Consensus

A distributed key-value store implementation using the Raft consensus algorithm to ensure consistency across multiple server nodes.

## Overview

This project implements a fault-tolerant distributed key-value store where multiple server nodes work together to maintain consistent data. The system uses the Raft consensus algorithm to ensure all servers agree on the state even if some servers fail.

## Features

- Distributed key-value storage with PUT, GET, and APPEND operations
- Fault tolerance using Raft consensus algorithm
- Leader election for continued operation even when nodes fail
- Log replication to ensure all nodes maintain consistent state
- HTTP API for client interactions

## Components

- **Server**: Handles HTTP requests from clients and forwards operations to the Raft cluster
- **Client**: Command-line tool to interact with the key-value store
- **KVStore**: In-memory key-value storage implementation
- **Raft**: Implementation of the Raft consensus algorithm for distributed consistency

## Usage

### Starting a Server Node

```bash
go run main.go -id=1 -port=8080
go run main.go -id=2 -port=8081
go run main.go -id=3 -port=8082
```

### Using the Client

```bash
# Put a key-value pair
go run client_main.go -server=localhost:8080 -op=put -key=mykey -value=myvalue

# Get a value
go run client_main.go -server=localhost:8080 -op=get -key=mykey

# Append to a value
go run client_main.go -server=localhost:8080 -op=append -key=mykey -value=_additional
```

## Architecture

The system follows a client-server architecture where:

1. Clients send requests to any server in the cluster  
2. The server forwards the request to the current Raft leader  
3. The leader replicates the command to follower nodes  
4. Once a majority of nodes confirm, the command is committed  
5. All nodes apply the command to their local key-value store  
6. The server responds to the client

## Dependencies

- Go 1.18 or higher

