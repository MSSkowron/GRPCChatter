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

## Requirements

Make sure you have Go 1.21 or a compatible version installed.

## Getting Started

### Installation and Running the GRPCChatter Server

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

    You can use the following flags to configure the server:
    - **--address**: The address at which the server listens.
    - **--port**: The port on which the server listens.
    - **--queue_size**: The maximum size of the message queue that is used to store messages to be sent to clients.

    Example of flags usage:

    ```
    go run ./cmd/grpcchatter/main.go --address "127.0.0.1" --port 5001 --queue_size 500
    ```

### GRPCChatter Server

The Server is the core component of the GRPCChatter application, responsible for handling various gRPC methods. Below are the methods supported by the server, along with their descriptions:

- **CreateChatRoom**: Allows clients to create a new chat room with custom name and password. Upon successful creation, it returns a short access code that users can employ to join the room later using JoinChatRoom method.

- **JoinChatRoom**: Allow clients to join existing chat rooms by providing the room's short access code and the associated password. After successful authentication, it returns a JSON Web Token (JWT) that contains two essential values in its payload: `userName` and `shortCode`. This JWT is crucial for facilitating communication within the room.

- **Chat**: Establishes a bidirectional streaming connection, enabling real-time chat between clients and the server. Clients can send messages to the server, and the server responds with incoming messages. To use this method, clients must attach a gRPC header with the key `token`, containing a valid JSON Web Token acquired from the JoinChatRoom method. This mechanism ensures continuous and secure communication within the chat room.

### GRPCChatter Client

The GRPCChatter Client is responsible for managing the client-side logic of the GRPCChatter application. It provides methods for creating chat rooms, joining chat rooms, sending messages, and receiving messages from the server. Client package is located in `/pkg/client`. Below are the methods supported by the client, along with their descriptions:

- **CreateChatRoom**: Creates a new chat room with the provided name and password. Upon successful creation, it returns the shortcode of the newly created chat room.

- **JoinChatRoom**: Connects the client to a specific chat room, enabling message reception and transmission. If the client is not connected, it establishes a connection, joins the chat room, and sets up a bidirectional stream for communication.

- **Send**: Sends a message to the server. It blocks until the message is sent or returns immediately if an error occurred. The JoinChatRoom method must be called before the first usage.

- **Receive**: Receives a message from the server. It blocks until a message arrives or returns immediately if an error occurred. The JoinChatRoom method must be called before the first usage.

- **Disconnect**: Disconnects the client from the server, closing the connection with the server.

## Example

An example client's code is provided [**here**](./examples/client_cli/main.go).
This example demonstrates the usage of the client package.

![Example](./docs/examples_client_cli.png)

## License

The project is available as open source under the terms of the MIT License.
