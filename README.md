# Auto Trade

A Go-based trading automation project with REST API and WebSocket support.

## Getting Started

To run the project:

```bash
go run main.go
```

The server will start on port 8080.

## API Endpoints

### Health Check
- **URL**: `/health`
- **Method**: `GET`
- **Response**: 
```json
{
    "status": "ok"
}
```

### Chat WebSocket
- **URL**: `ws://localhost:8080/ws`
- **Protocol**: `WebSocket`
- **Description**: Establishes a WebSocket connection for real-time chat communication

### Market Data Ticks WebSocket
- **URL**: `ws://localhost:8080/ws/ticks`
- **Protocol**: `WebSocket`
- **Description**: Streams real-time market data ticks
- **Data Format**:
```json
{
    "symbol": "BTC/USD",
    "price": 40000.00,
    "volume": 50.00,
    "timestamp": "2025-01-22T11:34:00Z"
}
```

## Test Pages
1. Chat Test Page (http://localhost:8080)
   - Connect to the WebSocket server
   - Send messages to all connected clients
   - Receive messages from other clients in real-time
   - Auto-reconnect if connection is lost

2. Market Data Page (http://localhost:8080/ticks.html)
   - Real-time market data updates
   - Price changes highlighted in green (up) or red (down)
   - Automatic updates for BTC/USD, ETH/USD, and SOL/USD
   - Auto-reconnect if connection is lost

## Requirements

- Go 1.x or higher
- github.com/gorilla/websocket
