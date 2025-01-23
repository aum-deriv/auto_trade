# Auto Trade

A Go application with a standard project structure.

## Project Structure

```
.
├── cmd/
│   └── app/
│       └── main.go      # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go    # Application configuration
│   └── service/
│       └── service.go   # Business logic
├── go.mod              # Go module definition
└── README.md          # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.21 or higher

### Installation

1. Clone the repository:
```bash
git clone https://github.com/aumbhatt/auto_trade.git
```

2. Navigate to the project directory:
```bash
cd auto_trade
```

3. Run the application:
```bash
go run cmd/app/main.go
```

## Development

The project follows standard Go project layout:

- `/cmd/app`: Contains the main application entry point
- `/internal`: Private application and library code
- `/internal/config`: Application configuration
- `/internal/service`: Core business logic

## License

This project is licensed under the MIT License.
