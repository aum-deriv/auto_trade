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

### Place Buy Trade
- **URL**: `/api/trade/buy`
- **Method**: `POST`
- **Request Body**:
```json
{
    "symbol": "BTC/USD"
}
```
- **Response**:
```json
{
    "trade_id": "TRADE-1",
    "symbol": "BTC/USD",
    "price": 40000.00,
    "status": "OPEN"
}
```
- **Supported Symbols**: BTC/USD, ETH/USD, SOL/USD

### Close (Sell) Trade
- **URL**: `/api/trade/sell`
- **Method**: `POST`
- **Request Body**:
```json
{
    "trade_id": "TRADE-1"
}
```
- **Response**:
```json
{
    "trade_id": "TRADE-1",
    "symbol": "BTC/USD",
    "price": 40000.00,
    "status": "CLOSED"
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

### Open Trades WebSocket
- **URL**: `ws://localhost:8080/ws/trades`
- **Protocol**: `WebSocket`
- **Description**: Streams real-time updates of all open trades
- **Data Format**:
```json
[
    {
        "id": "TRADE-1",
        "symbol": "BTC/USD",
        "type": "BUY",
        "price": 40000.00,
        "status": "OPEN",
        "timestamp": "2025-01-22T11:34:00Z"
    }
]
```

### Single Trade WebSocket
- **URL**: `ws://localhost:8080/ws/trade/{trade_id}`
- **Protocol**: `WebSocket`
- **Description**: Streams real-time updates for a specific trade
- **Data Format**:
```json
{
    "id": "TRADE-1",
    "symbol": "BTC/USD",
    "type": "BUY",
    "price": 40000.00,
    "status": "OPEN",
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

3. Trade Monitor Page (http://localhost:8080/trades.html)
   - Real-time monitoring of all open trades
   - Individual trade monitoring by ID
   - Automatic updates of trade status
   - Split view for all trades and single trade monitoring

## Requirements

- Go 1.x or higher
- github.com/gorilla/websocket
