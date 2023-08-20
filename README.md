# GRPCChatter - Real-time Chat Application with gRPC

GRPCChatter is a real-time chat application implemented in Go, leveraging gRPC to facilitate seamless communication between clients. This application empowers users to connect and exchange messages in a responsive and uninterrupted manner.

## Technologies

- Go 1.21
- gRPC

## Architecture Overview

GRPCChatter employs a robust client-server architecture where communication is facilitated through the gRPC framework. Clients establish connections with the server via gRPC bi-directional channels using `GRPCChatter Client`, allowing for real-time, bi-directional message exchange.

![Architecture](./docs/architecture.png)

## Features

- **Real-Time Communication**: GRPCChatter ensures instantaneous message delivery, establishing true real-time communication among clients.

- **Efficient Protocol**: By leveraging the gRPC framework, the application benefits from the efficiency and high-performance of the protobuf communication protocol.

- **User-Friendly**: The provided client code offers an intuitive and user-friendly interface, simplifying the process of initiating conversations and exchanging messages.

## Key Dependencies

## Requirements

Make sure you have Go 1.21 or a compatible version installed.

## Getting Started

### Installation and Running the `GRPCChatter Server`

To start using GRPCChatter, follow these steps to run the server:

1. Clone the repository:

    ```
    git clone https://github.com/MSSkowron/GRPCChatter.git
    ```

2. Navigate to the project directory:

    ```
    cd GRPCChatter
    ```

3. Build & Run the application's server:

    ```
    go run ./cmd/grpcchatter/main.go
    ```

### Using the `GRPCChatter Client` for chatting

Utilize the client code located in `/pkg/client` to initiate conversations and start chatting.
An example client's code is provided [**here**](./examples/client/main.go).

## License

The project is available as open source under the terms of the MIT License.
