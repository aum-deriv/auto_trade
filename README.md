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

### Start Trading Strategy
- **URL**: `/api/strategy/start`
- **Method**: `POST`
- **Request Body**:
```json
{
    "type": "MARTINGALE",
    "symbol": "BTC/USD",
    "base_size": 100.0,
    "max_losses": 3
}
```
- **Response**:
```json
{
    "strategy_id": "MART-BTC/USD-100.00-3",
    "status": "started",
    "message": "Strategy started successfully"
}
```
- **Supported Strategy Types**: MARTINGALE

### Stop Trading Strategy
- **URL**: `/api/strategy/stop`
- **Method**: `POST`
- **Request Body**:
```json
{
    "strategy_id": "MART-BTC/USD-100.00-3"
}
```
- **Response**:
```json
{
    "strategy_id": "MART-BTC/USD-100.00-3",
    "status": "stopped",
    "message": "Strategy stopped successfully"
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

### Active Strategies WebSocket
- **URL**: `ws://localhost:8080/ws/strategies`
- **Protocol**: `WebSocket`
- **Description**: Streams real-time updates of all active trading strategies
- **Data Format**:
```json
[
    {
        "strategy_id": "MART-BTC/USD-100.00-3",
        "type": "MARTINGALE",
        "symbol": "BTC/USD",
        "base_size": 100.00,
        "max_losses": 3,
        "current_size": 200.00,
        "consecutive_losses": 1
    }
]
```

### Single Strategy WebSocket
- **URL**: `ws://localhost:8080/ws/strategy/{strategy_id}`
- **Protocol**: `WebSocket`
- **Description**: Streams real-time updates for a specific strategy
- **Data Format**:
```json
{
    "strategy_id": "MART-BTC/USD-100.00-3",
    "type": "MARTINGALE",
    "symbol": "BTC/USD",
    "base_size": 100.00,
    "max_losses": 3,
    "current_size": 200.00,
    "consecutive_losses": 1
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

4. Strategy Management Page (http://localhost:8080/strategy.html)
   - Real-time market data display with price change indicators
   - Start/stop Martingale strategies with configurable parameters
   - Live strategy status monitoring (position size, loss counter)
   - Automatic WebSocket reconnection on disconnection
   - Clean UI with success/error status messages

## Trading Strategies

### Martingale Strategy
The project includes a Martingale trading strategy implementation that automatically adjusts position sizes based on trading outcomes.

#### Strategy Parameters
```go
strategy := strategies.NewMartingaleStrategy(
    baseSize: 100.0,    // Initial position size (e.g., $100)
    symbol: "BTC/USD",  // Trading pair
    maxLosses: 3,       // Maximum consecutive losses allowed
)

// Strategy ID is automatically generated in format: MART-{symbol}-{baseSize}-{maxLosses}
// Example: MART-BTC/USD-100.00-3
```

The strategy ID format encodes the key parameters, making it easy to:
- Track performance of specific strategy configurations
- Share strategy configurations with other users
- Match trades with their corresponding strategy instance

#### How it Works
1. Starts with the base position size
2. After a loss:
   - Doubles the position size for the next trade
   - Continues doubling until either:
     - A winning trade occurs (resets to base size)
     - Max losses limit is reached (stops trading)
3. After a win:
   - Resets position size back to base amount
   - Resets consecutive loss counter

#### Usage Example
```go
// Create new strategy instance
strategy := strategies.NewMartingaleStrategy(100.0, "BTC/USD", 3)

// Process trade result and get next position size
nextSize, shouldTrade, err := strategy.ProcessTrade(trade, closePrice)
if err != nil {
    // Handle error (e.g., max losses reached)
    return
}
if !shouldTrade {
    // Strategy suggests stopping (e.g., max losses reached)
    return
}

// Use nextSize for your next trade
// ...
```

## Requirements

- Go 1.x or higher
- github.com/gorilla/websocket
