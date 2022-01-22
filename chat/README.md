# TCP Chat Server and Client
Simple TCP chat server and client.  I started this to explore Go's concurrency in depth.

- single TCP server handles multiple clients connected concurrently
- one-to-many message broadcast 

# Running
Run a single server and multiple clients in separate terminals.
```zsh
# Server (single)
> go run server/chat_server.go

# Client (one or more)
> go run client/chat_client.go
```

# TODO
- concurrent error handling
- colorized terminal output
- common client/server message serialization
- message compression
- message encryption
- utilize channels as means of concurrently write stdout and record file-based chat log
- parameterize config info (e.g. port, etc)

# Future Ideas
- concept of chat "rooms" with join/leave capability
- add persistence and stateless servers with high availability

#Acknowledgements
I seeded this implementation from the basic implementation here
 https://golangforall.com/en/post/golang-tcp-server-chat.html
