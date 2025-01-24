# Auto Trade API Documentation

## Trading Endpoints

### REST API

The trading API provides endpoints for managing individual trades, including opening new positions and closing existing ones, with real-time updates via WebSocket.

#### Create Trade (Buy)
> Opens a new long position for the specified symbol at the given price
```http
POST /api/trades/buy
```

Request Body:
```json
{
    "symbol": "AAPL",
    "entry_price": 150.25
}
```

Success Response (200 OK):
```json
{
    "trade_id": "trade-abc123",
    "symbol": "AAPL",
    "entry_price": 150.25,
    "entry_time": "2025-01-23T14:23:38Z"
}
```

Error Response (400 Bad Request):
```json
{
    "code": "INVALID_SYMBOL",
    "message": "Invalid trading symbol: XYZ"
}
```

#### Close Trade (Sell)
> Closes an existing position identified by trade_id and records the exit price
```http
POST /api/trades/sell
```

Request Body:
```json
{
    "trade_id": "trade-abc123"
}
```

Success Response (200 OK):
```json
{
    "trade_id": "trade-abc123",
    "symbol": "AAPL",
    "entry_price": 150.25,
    "exit_price": 151.50,
    "entry_time": "2025-01-23T14:23:38Z",
    "exit_time": "2025-01-23T14:30:00Z"
}
```

Error Response (404 Not Found):
```json
{
    "code": "TRADE_NOT_FOUND",
    "message": "Trade not found: trade-abc123"
}
```

### WebSocket Events

Connect to WebSocket endpoint: `ws://localhost:8080/ws`

#### Subscribe to Open Positions
> Provides real-time updates of all currently open trading positions
```json
// Client -> Server
{
    "type": "subscribe",
    "payload": {
        "type": "open_positions"
    }
}

// Server -> Client (Success)
{
    "type": "open_positions",
    "subscribe_id": "sub-123",
    "payload": [
        {
            "trade_id": "trade-abc123",
            "symbol": "AAPL",
            "entry_price": 150.25,
            "entry_time": "2025-01-23T14:23:38Z"
        }
    ]
}
```

#### Subscribe to Trade History
> Delivers updates about completed trades, including entry/exit prices and timestamps
```json
// Client -> Server
{
    "type": "subscribe",
    "payload": {
        "type": "trade_history"
    }
}

// Server -> Client (Success)
{
    "type": "trade_history",
    "subscribe_id": "sub-456",
    "payload": [
        {
            "trade_id": "trade-xyz789",
            "symbol": "GOOGL",
            "entry_price": 140.50,
            "exit_price": 142.75,
            "entry_time": "2025-01-23T13:00:00Z",
            "exit_time": "2025-01-23T14:00:00Z"
        }
    ]
}
```

## Strategy Endpoints

### REST API

The strategy API provides endpoints for managing automated trading strategies, including listing available strategies, starting new instances with custom parameters, and controlling their execution.

#### List Available Strategies
> Returns a list of all available trading strategies with their parameters and execution flow
```http
GET /api/strategies/default
```

Success Response (200 OK):
```json
{
    "strategies": [
        {
            "name": "martingale",
            "parameters": [
                {
                    "name": "symbol",
                    "type": "string",
                    "required": true,
                    "description": "Trading symbol (e.g. AAPL)"
                },
                {
                    "name": "base_position",
                    "type": "number",
                    "required": true,
                    "description": "Initial position size in dollars"
                },
                {
                    "name": "take_profit",
                    "type": "number",
                    "required": true,
                    "description": "Price increase percentage for taking profit"
                },
                {
                    "name": "max_positions",
                    "type": "number",
                    "required": true,
                    "description": "Maximum number of increasing positions"
                }
            ],
            "strategy_flow": [
                "1. Start with base_position size",
                "2. Enter long position at market price",
                "3. Set take profit target at entry_price * (1 + take_profit/100)",
                "4. If target hit: Take profit and reset position size",
                "5. If price drops: Exit at loss",
                "6. If under max_positions: Double position size and enter new position",
                "7. If at max_positions: Reset position size to base_position",
                "8. Repeat from step 1"
            ]
        }
    ]
}
```

#### Start Strategy
> Initiates a new instance of the specified strategy with the given parameters
```http
POST /api/strategies/start
```

Request Body:
```json
{
    "name": "martingale",
    "parameters": {
        "symbol": "AAPL",
        "base_position": 100.0,
        "take_profit": 1.0,
        "max_positions": 3
    }
}
```

Success Response (200 OK):
```json
{
    "strategy_id": "strat-abc123",
    "name": "martingale",
    "status": "running",
    "start_time": "2025-01-23T14:23:38Z"
}
```

Error Response (400 Bad Request):
```json
{
    "code": "INVALID_PARAMETERS",
    "message": "Invalid base_position: must be greater than 0"
}
```

#### Stop Strategy
> Gracefully stops a running strategy instance and records its completion time
```http
POST /api/strategies/stop
```

Request Body:
```json
{
    "strategy_id": "strat-abc123"
}
```

Success Response (200 OK):
```json
{
    "strategy_id": "strat-abc123",
    "name": "martingale",
    "status": "stopped",
    "start_time": "2025-01-23T14:23:38Z",
    "stop_time": "2025-01-23T14:30:00Z"
}
```

Error Response (404 Not Found):
```json
{
    "code": "STRATEGY_NOT_FOUND",
    "message": "Strategy not found: strat-abc123"
}
```

### WebSocket Events

#### Subscribe to Active Strategies
> Provides real-time updates about currently running strategies and their status
```json
// Client -> Server
{
    "type": "subscribe",
    "payload": {
        "type": "active_strategies"
    }
}

// Server -> Client (Success)
{
    "type": "active_strategies",
    "subscribe_id": "sub-789",
    "payload": [
        {
            "strategy_id": "strat-abc123",
            "name": "martingale",
            "status": "running",
            "parameters": {
                "symbol": "AAPL",
                "base_position": 100.0,
                "take_profit": 1.0,
                "max_positions": 3
            },
            "start_time": "2025-01-23T14:23:38Z"
        }
    ]
}
```

## Understanding Strategy Metadata

The `/api/strategies/default` endpoint returns metadata that describes available trading strategies. This information is crucial for:

1. Understanding strategy behavior
2. Validating parameters
3. Building dynamic user interfaces

### Metadata Structure

Each strategy includes:

1. **Name**: Unique identifier for the strategy
2. **Parameters**: List of required and optional configuration values
3. **Strategy Flow**: Step-by-step description of execution logic

### Parameter Types

1. **string**
   - Used for: Symbols, identifiers
   - Validation:
     * Non-empty if required
     * Case-sensitive
     * No special validation unless specified

2. **number**
   - Used for: Prices, quantities, percentages
   - Validation:
     * Must be positive unless specified
     * Decimal precision based on context
     * Range validation if specified

3. **boolean**
   - Used for: Flags, toggles
   - Values: true/false
   - No special validation

### UI Generation Guidelines

1. **Text Fields (string)**
   - Components:
     * Input field for text entry
     * Label showing parameter name
     * Required field indicator (if applicable)
     * Description/placeholder text
   - Behavior:
     * Trim whitespace on input
     * Validate non-empty if required
     * Show validation feedback

2. **Number Fields (number)**
   - Components:
     * Numeric input field
     * Step controls (if applicable)
     * Unit indicator (if applicable)
     * Valid range hints
   - Behavior:
     * Accept decimal input
     * Validate numeric range
     * Format according to locale
     * Show validation feedback

3. **Boolean Fields (boolean)**
   - Components:
     * Toggle switch or checkbox
     * Label with description
   - Behavior:
     * Binary state (true/false)
     * Clear visual indication of state

### Parameter Validation

1. **Required Fields**
   - Check if value is present
   - Check if value is non-empty
   - Return appropriate error message
   - Example error format:
     ```json
     {
         "field": "symbol",
         "error": "Field is required"
     }
     ```

2. **Type-Specific Validation**

   a. String Parameters:
   - Validation Rules:
     * Non-empty check
     * Length limits (if specified)
     * Pattern matching (if applicable)
   - Example:
     ```
     symbol: "AAPL" ✓
     symbol: ""     ✗ (empty)
     symbol: "123"  ✗ (invalid pattern)
     ```

   b. Number Parameters:
   - Validation Rules:
     * Numeric format
     * Range checks
     * Precision limits
   - Example:
     ```
     base_position: 100.00  ✓
     base_position: -50.00  ✗ (negative)
     base_position: "abc"   ✗ (non-numeric)
     ```

   c. Boolean Parameters:
   - Validation Rules:
     * Must be true/false
     * No null values
   - Example:
     ```
     enabled: true   ✓
     enabled: false  ✓
     enabled: null   ✗ (invalid)
     ```

### Error Handling

1. **Client-Side Validation**
   - Process:
     1. Collect all parameter values
     2. Validate each parameter
     3. Aggregate validation errors
     4. Return error collection
   - Error Format:
     ```json
     {
         "errors": [
             {
                 "field": "symbol",
                 "error": "Invalid trading symbol"
             },
             {
                 "field": "base_position",
                 "error": "Must be greater than 0"
             }
         ]
     }
     ```

2. **Server Response Handling**
   - HTTP Status Codes:
     * 200: Success
     * 400: Invalid parameters
     * 404: Resource not found
     * 500: Server error
   - Error Response Format:
     ```json
     {
         "code": "INVALID_PARAMETERS",
         "message": "Invalid strategy parameters",
         "details": {
             "symbol": "Invalid trading symbol: XYZ",
             "base_position": "Must be greater than 0"
         }
     }
     ```
   - Success Response Format:
     ```json
     {
         "strategy_id": "strat-abc123",
         "status": "running",
         "message": "Strategy started successfully"
     }
     ```
