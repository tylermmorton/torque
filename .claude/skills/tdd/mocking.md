# When to Mock

Mock at **system boundaries** only:

- External APIs (payment, email, third-party services)
- Time/randomness
- File system (sometimes)
- Database — prefer a real test DB; use mocks for error paths that are impractical to trigger against a real DB (constraint violations, connection failures, etc.)

Don't mock:

- Your own packages
- Internal collaborators
- Anything you control and can test through the real interface

## Designing for Mockability

At system boundaries, design interfaces that are easy to mock:

**1. Use dependency injection**

Pass external dependencies in rather than creating them internally:

```go
// Easy to mock — accepts interface
func processPayment(order Order, client PaymentClient) error {
    return client.Charge(order.Total)
}

// Hard to mock — creates dependency internally
func processPayment(order Order) error {
    client := stripe.NewClient(os.Getenv("STRIPE_KEY"))
    return client.Charge(order.Total)
}
```

**2. Define narrow interfaces at the call site**

Go interfaces are satisfied implicitly. Define the interface where it's consumed, not where it's implemented:

```go
// PaymentClient is defined in the package that uses it
type PaymentClient interface {
    Charge(amount int64) error
}
```

**3. Prefer SDK-style interfaces over generic ones**

Create specific methods for each external operation instead of one generic method with conditional logic:

```go
// GOOD: Each method is independently mockable
type UserAPI interface {
    GetUser(ctx context.Context, id string) (*User, error)
    ListOrders(ctx context.Context, userID string) ([]Order, error)
    CreateOrder(ctx context.Context, data OrderInput) (*Order, error)
}

// BAD: Mock requires conditional logic to distinguish calls
type API interface {
    Do(ctx context.Context, method, endpoint string, body any) ([]byte, error)
}
```
