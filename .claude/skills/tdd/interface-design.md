# Interface Design for Testability

Good interfaces make testing natural:

**1. Accept dependencies, don't create them**

```go
// Testable — dependency injected
func processOrder(ctx context.Context, order Order, gateway PaymentGateway) error {
    return gateway.Charge(ctx, order.Total)
}

// Hard to test — dependency created internally
func processOrder(ctx context.Context, order Order) error {
    gateway := stripe.NewGateway(os.Getenv("STRIPE_KEY"))
    return gateway.Charge(ctx, order.Total)
}
```

**2. Return results, don't mutate through side effects**

```go
// Testable — result is observable
func calculateDiscount(cart Cart) (Discount, error) {
    // ...
}

// Hard to test — mutation is the only observable effect
func applyDiscount(cart *Cart) {
    cart.Total -= computeDiscount(cart)
}
```

**3. Small surface area**

- Fewer methods = fewer tests needed
- Fewer parameters = simpler test setup
- Define interfaces where they're consumed, not where they're implemented

```go
// Narrow interface at the call site — only what this function needs
type UserLookup interface {
    GetByID(ctx context.Context, id string) (*User, error)
}

func greetUser(ctx context.Context, id string, users UserLookup) (string, error) {
    user, err := users.GetByID(ctx, id)
    if err != nil {
        return "", err
    }
    return "Hello, " + user.Name, nil
}
```