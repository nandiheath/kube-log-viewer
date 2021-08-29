# kube-logs-viewer
A lightweight tool that expose your pod's logs configured with annotations

## Design

### Pure HTTP/2 Implementation

- HTTP/2 is not designed for server push payload as API calls (we can but the browser does not fully support)
- Raw implementation on browser side may needed

### GRPC Implementation

- .proto overheads (writing/building)
- No official Web base GRPC library support streams - workarounds may introduce more work

### Websocket

- easy implementation on client side
- easy infrastructure setup

