# Product Catalog Service

A production-grade product catalog service implemented using Domain-Driven Design (DDD), Clean Architecture, and the Golden Mutation Pattern with Go, gRPC, and Google Cloud Spanner.

## Architecture Overview

This service demonstrates enterprise-level patterns used in production systems:

- **Domain-Driven Design (DDD)**: Pure domain layer with aggregates, value objects, and domain services
- **Clean Architecture**: Clear separation between domain, application, and infrastructure layers
- **CQRS**: Separate command and query paths for optimal performance
- **Golden Mutation Pattern**: Atomic transactions using CommitPlan pattern
- **Transactional Outbox**: Reliable event publishing within database transactions
- **Change Tracking**: Optimized updates that only modify changed fields

## Project Structure

```
product-catalog-service/
├── cmd/
│   └── server/
│       └── main.go                    # Service entry point
├── internal/
│   ├── app/product/
│   │   ├── domain/                    # Pure business logic (no external deps)
│   │   │   ├── product.go             # Product aggregate
│   │   │   ├── discount.go            # Discount value object
│   │   │   ├── money.go               # Money value object (big.Rat)
│   │   │   ├── change_tracker.go      # Change tracking
│   │   │   ├── domain_events.go       # Domain events
│   │   │   ├── domain_errors.go       # Sentinel errors
│   │   │   └── services/
│   │   │       └── pricing_calculator.go  # Domain service
│   │   ├── usecases/                  # Application layer (orchestration)
│   │   │   ├── create_product/
│   │   │   ├── update_product/
│   │   │   ├── activate_product/
│   │   │   ├── deactivate_product/
│   │   │   ├── apply_discount/
│   │   │   └── remove_discount/
│   │   ├── queries/                   # CQRS read side
│   │   │   ├── get_product/
│   │   │   └── list_products/
│   │   ├── contracts/                 # Repository interfaces
│   │   └── repo/                      # Repository implementations
│   ├── models/                        # Database models
│   │   ├── m_product/
│   │   └── m_outbox/
│   ├── transport/grpc/product/        # gRPC handlers
│   ├── services/                      # DI container
│   ├── commitplan/                    # CommitPlan implementation
│   └── pkg/                           # Shared utilities
├── proto/product/v1/                  # Protocol Buffers
├── migrations/                        # Database migrations
├── tests/e2e/                         # E2E tests
├── docker-compose.yml                 # Spanner emulator
├── Makefile                           # Build automation
└── README.md
```

## Prerequisites

The following tools are installed in this workspace:

- Go 1.21+ (installed in workspace)
- Protocol Buffers compiler (protoc) (installed in workspace)
- Docker (for Spanner emulator)
- Make

## Getting Started

### 1. Start Spanner Emulator

```bash
make docker-up
```

This starts the Google Cloud Spanner emulator on ports 9010 and 9020.

### 2. Setup Database

**Option A: Manual Setup (Recommended for testing)**

The service will attempt to connect to the database specified in the environment variable. For the emulator, you can use:

```bash
export SPANNER_EMULATOR_HOST=localhost:9010
export SPANNER_DATABASE=projects/test-project/instances/test-instance/databases/test-database
```

**Option B: Using spanner-cli**

```bash
# Install spanner-cli
go install github.com/cloudspannerecosystem/spanner-cli@latest

# Create database and run migrations
spanner-cli -p test-project -i test-instance -d test-database < migrations/001_initial_schema.sql
```

### 3. Run Tests

```bash
make test
```

This runs all E2E tests using the Spanner emulator.

### 4. Start the Server

```bash
make run
```

The gRPC server will start on port 50051 (configurable via PORT environment variable).

### 5. Test with grpcurl

```bash
# Create a product
grpcurl -plaintext -d '{
  "name": "Test Product",
  "description": "A test product",
  "category": "Electronics",
  "price_numerator": 1999,
  "price_denominator": 100
}' localhost:50051 product.v1.ProductService/CreateProduct

# Get a product
grpcurl -plaintext -d '{
  "product_id": "YOUR_PRODUCT_ID"
}' localhost:50051 product.v1.ProductService/GetProduct

# List products
grpcurl -plaintext -d '{
  "category": "Electronics",
  "status": "active",
  "limit": 10
}' localhost:50051 product.v1.ProductService/ListProducts
```

## Key Design Decisions

### 1. Domain Purity

The domain layer (`internal/app/product/domain/`) is completely pure:
- ✅ No `context.Context` imports
- ✅ No database/infrastructure imports
- ✅ No external framework dependencies
- ✅ Uses `*big.Rat` for precise monetary calculations
- ✅ Sentinel errors for domain validation
- ✅ Change tracking for optimized updates
- ✅ Events as simple structs (intents only)

### 2. Golden Mutation Pattern

Every write operation follows this exact pattern:

```go
func (it *Interactor) Execute(ctx context.Context, req Request) error {
    // 1. Load or create aggregate
    product := ...
    
    // 2. Call domain method (validation happens here)
    product.DoSomething(...)
    
    // 3. Build commit plan
    plan := committer.NewPlan()
    
    // 4. Get mutations from repository (repo returns, doesn't apply)
    if mut := it.repo.UpdateMut(product); mut != nil {
        plan.Add(spanner.NewSpannerMutation(mut))
    }
    
    // 5. Add outbox events
    for _, event := range product.DomainEvents() {
        outboxMut := it.outboxRepo.InsertMut(enrichEvent(event))
        plan.Add(spanner.NewSpannerMutation(outboxMut))
    }
    
    // 6. Apply plan atomically (usecase applies, NOT handler!)
    return it.committer.Apply(ctx, plan)
}
```

### 3. CQRS Separation

- **Commands**: Go through domain aggregate → CommitPlan → Apply
- **Queries**: Direct database access → DTOs (optimized reads, no domain logic)

### 4. Repository Pattern

Repositories **return mutations, never apply them**:
- Maps domain ↔ database models
- Uses change tracker to build targeted updates
- Only includes changed fields in updates

### 5. Money Handling

All monetary values use `*big.Rat` for precise decimal arithmetic:

```go
// Stored in database as numerator/denominator (INT64 columns)
basePrice := domain.NewMoney(1999, 100)  // $19.99

// Discount calculation
discountPercent := big.NewRat(20, 100)   // 20%
discountAmount := basePrice.Multiply(discountPercent)
effectivePrice := basePrice.Subtract(discountAmount)  // $15.99
```

### 6. Transactional Outbox

- Domain captures event intents (simple structs)
- Usecases enrich events with metadata
- Events stored in same transaction as aggregate
- Background processor would publish to message broker (not implemented)

## Testing

### E2E Tests

The service includes comprehensive E2E tests that:
- Use real Spanner emulator (no mocking)
- Test complete flows through use cases
- Verify business rules and validation
- Check side effects (outbox events)
- Test discount calculations
- Verify state transitions

Run tests:
```bash
make test
```

### Test Scenarios Covered

1. **Product Creation Flow** - Create → verify in DB → check outbox
2. **Product Update Flow** - Update details → verify changes
3. **Discount Application** - Apply discount → verify price calculation
4. **Activation/Deactivation** - State transitions
5. **Business Rule Validation** - Domain error handling
6. **List Products** - Filtering and pagination

## API Reference

### gRPC Service: ProductService

#### Commands (Write Operations)

- `CreateProduct` - Creates a new product
- `UpdateProduct` - Updates product details
- `ActivateProduct` - Activates a product
- `DeactivateProduct` - Deactivates a product
- `ApplyDiscount` - Applies a percentage discount
- `RemoveDiscount` - Removes current discount

#### Queries (Read Operations)

- `GetProduct` - Retrieves a product with effective price
- `ListProducts` - Lists products with filtering and pagination

See `proto/product/v1/product_service.proto` for full API definition.

## Business Rules

1. Only **active** products can have discounts
2. Only **one active discount** per product at a time
3. Discounts have **start and end dates**
4. Discount percentage must be **0-100**
5. Product status flow: **Draft → Active → Inactive → Archived**
6. Archived products **cannot be modified**

## Environment Variables

- `PORT` - Server port (default: 50051)
- `SPANNER_DATABASE` - Spanner database path
- `SPANNER_EMULATOR_HOST` - Emulator host (default: localhost:9010)

## Makefile Commands

```bash
make help           # Show all available commands
make proto          # Generate protobuf code
make docker-up      # Start Spanner emulator
make docker-down    # Stop Spanner emulator
make test           # Run all tests
make run            # Start the server
make build          # Build server binary
make clean          # Clean generated files
make fmt            # Format code
make vet            # Run go vet
```

## Design Patterns Used

1. **Aggregate Pattern** - Product is the aggregate root
2. **Value Objects** - Money, Discount are immutable value objects
3. **Repository Pattern** - Abstract data access
4. **Factory Pattern** - Domain object construction
5. **Domain Services** - PricingCalculator for cross-entity logic
6. **Command Pattern** - Use cases as commands
7. **Query Object Pattern** - CQRS queries
8. **Transactional Outbox** - Reliable event publishing
9. **Change Tracking** - Optimistic updates
10. **Dependency Injection** - Container pattern

## Production Considerations

### Not Implemented (Out of Scope)

- Background outbox processor
- Actual message broker publishing
- Authentication/Authorization
- Rate limiting
- API gateway
- REST endpoints
- Metrics/monitoring beyond logging
- Distributed tracing
- Circuit breakers

### What Would Be Added for Production

1. **Observability**
   - OpenTelemetry tracing
   - Prometheus metrics
   - Structured logging (zerolog)

2. **Resilience**
   - Retry policies
   - Circuit breakers
   - Graceful degradation

3. **Security**
   - JWT authentication
   - Authorization middleware
   - TLS for gRPC

4. **Operations**
   - Health checks
   - Readiness probes
   - Configuration management
   - Secret management

## License

This is a test task implementation for demonstration purposes.

## Notes

This implementation uses a local `commitplan` package since the referenced `github.com/Vektor-AI/commitplan` package was not publicly accessible. The local implementation provides the same API surface and functionality for demonstration purposes.
