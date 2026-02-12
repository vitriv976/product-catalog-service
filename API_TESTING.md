# API Testing Guide (Postman / gRPC)

Use these JSON payloads when testing with Postman or grpcurl. **Use `product_id` (snake_case)** in all requests.

## Prerequisites
- Server running: `make run`
- Database created: `make setup-db`
- Use your actual `product_id` from CreateProduct response (e.g. `5eef6a47-6031-4638-84ed-e8758dc8fa5b`)

---

## 1. CreateProduct (works - creates new product)
```json
{
  "name": "Wireless Headphones",
  "description": "Noise cancelling",
  "category": "Electronics",
  "price_numerator": 9999,
  "price_denominator": 100
}
```
Response: `{"product_id": "uuid-here"}` — save this for other calls.

---

## 2. GetProduct
```json
{
  "product_id": "5eef6a47-6031-4638-84ed-e8758dc8fa5b"
}
```
Replace with your actual product_id.

---

## 3. ActivateProduct (required before ApplyDiscount)
```json
{
  "product_id": "5eef6a47-6031-4638-84ed-e8758dc8fa5b"
}
```

---

## 4. UpdateProduct
```json
{
  "product_id": "5eef6a47-6031-4638-84ed-e8758dc8fa5b",
  "name": "Updated Wireless Headphones",
  "description": "Premium noise cancelling",
  "category": "Electronics"
}
```

---

## 5. ApplyDiscount (product must be active)
For timestamps, use Unix seconds. Get current: `date +%s` in terminal.
```json
{
  "product_id": "5eef6a47-6031-4638-84ed-e8758dc8fa5b",
  "discount_percent": 15,
  "start_date": {"seconds": 1739318400, "nanos": 0},
  "end_date": {"seconds": 1739664000, "nanos": 0}
}
```
Use `seconds` values that include today (e.g. start in past, end in future).

---

## 6. RemoveDiscount
```json
{
  "product_id": "5eef6a47-6031-4638-84ed-e8758dc8fa5b"
}
```

---

## 7. ListProducts
List all products (including draft):
```json
{}
```

List active products only:
```json
{
  "status": "active",
  "limit": 10,
  "offset": 0
}
```

List by category:
```json
{
  "category": "Electronics",
  "status": "draft",
  "limit": 10
}
```

---

## 8. DeactivateProduct
```json
{
  "product_id": "5eef6a47-6031-4638-84ed-e8758dc8fa5b"
}
```

---

## Testing Order
1. CreateProduct → get product_id
2. ActivateProduct
3. GetProduct → verify
4. UpdateProduct
5. ApplyDiscount
6. GetProduct → verify effective_price
7. ListProducts
8. RemoveDiscount / DeactivateProduct

---

## If You See Errors
- **"product_id is required"** — Use `product_id` (snake_case), not `productId`
- **"product not found"** — Wrong product_id or product doesn't exist
- **"internal server error: ..."** — Check server logs; the message shows the real error
- **ApplyDiscount fails** — Ensure product is active first; use valid start_date/end_date
