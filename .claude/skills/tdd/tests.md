# Good and Bad Tests

## Good Tests

**Integration-style**: Test through real interfaces, not mocks of internal parts.

```go
// GOOD: Tests observable behavior
func TestOrder_Checkout(t *testing.T) {
    cart := newCart()
    cart.Add(product)
    result, err := checkout(ctx, cart, paymentMethod)
    require.NoError(t, err)
    require.Equal(t, "confirmed", result.Status)
}
```

Characteristics:

- Tests behavior callers care about
- Uses public API only
- Survives internal refactors
- Describes WHAT, not HOW
- One logical assertion per test (or per `t.Run` sub-case)

## Bad Tests

**Implementation-detail tests**: Coupled to internal structure.

```go
// BAD: Asserts on call counts — tests HOW, not WHAT
func TestOrder_Checkout_CallsPayment(t *testing.T) {
    mock := &MockPaymentClient{}
    _ = checkout(ctx, cart, mock)
    require.Equal(t, 1, mock.ChargeCalls)
}
```

Red flags:

- Asserting on call counts or call order
- Mocking internal collaborators
- Testing unexported functions
- Test breaks on refactor without behavior change
- Test name describes HOW not WHAT

```go
// BAD: Bypasses interface to verify — queries DB directly
func TestUser_Create_SavesToDB(t *testing.T) {
    err := createUser(ctx, db, User{Name: "Alice"})
    require.NoError(t, err)

    var row User
    err = db.QueryRowContext(ctx, "SELECT * FROM users WHERE name = $1", "Alice").Scan(&row)
    require.NoError(t, err)
}

// GOOD: Verifies through the interface
func TestUser_Create(t *testing.T) {
    user, err := createUser(ctx, db, User{Name: "Alice"})
    require.NoError(t, err)

    retrieved, err := getUser(ctx, db, user.ID)
    require.NoError(t, err)
    require.Equal(t, "Alice", retrieved.Name)
}
```

## Sub-behaviors with t.Run

Use `t.Run` for error cases and variants within the same function under test:

```go
func TestUser_GetByID(t *testing.T) {
    user, err := createUser(ctx, db, User{Name: "Alice"})
    require.NoError(t, err)

    retrieved, err := getUserByID(ctx, db, user.ID)
    require.NoError(t, err)
    require.Equal(t, "Alice", retrieved.Name)

    t.Run("not_found", func(t *testing.T) {
        _, err := getUserByID(ctx, db, unknownID)
        require.Error(t, err)
        require.ErrorContains(t, err, "no rows in result set")
    })
}
```