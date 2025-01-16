 CREATE TABLE "invoices" (
 "invoice_number" bigserial PRIMARY KEY,
  "customer_name" varchar NOT NULL,
  "customer_email" varchar NOT NULL,
  "customer_phone" varchar NOT NULL,
  "customer_address" varchar NOT NULL,
  "sender_name" varchar NOT NULL,
  "sender_email" varchar NOT NULL,
  "sender_phone" varchar NOT NULL,
  "sender_address" varchar NOT NULL,
  "issue_date" varchar NOT NULL,
  "due_date" varchar NOT NULL,
  "status" varchar NOT NULL,
  "subtotal" bigint NOT NULL,
  "discount_rate_in_percent" varchar NOT NULL,
  "discount" bigint NOT NULL,
  "total_amount" bigint NOT NULL,
  "billing_currency" varchar NOT NULL DEFAULT 'USD',
  "payment_info" varchar NOT NULL,
  "note" varchar NOT NULL DEFAULT 'Thank you for your patronage',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "line_items" (
  "id" bigserial PRIMARY KEY,
  "invoice_number" bigint NOT NULL,
  "description" varchar NOT NULL,
  "quantity" bigint NOT NULL,
  "unit_price" bigint NOT NULL,
  "total_price" bigint NOT NULL
);

CREATE INDEX ON "invoices" ("status");

CREATE INDEX ON "line_items" ("invoice_number");

ALTER TABLE "line_items" ADD FOREIGN KEY ("invoice_number") REFERENCES "invoices" ("invoice_number");