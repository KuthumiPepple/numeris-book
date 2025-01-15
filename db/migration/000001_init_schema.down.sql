ALTER TABLE "line_items" DROP CONSTRAINT "line_items_invoice_number_fkey";

DROP INDEX IF EXISTS "invoices_status_idx";
DROP INDEX IF EXISTS "line_items_invoice_number_idx";

DROP TABLE IF EXISTS "line_items";
DROP TABLE IF EXISTS "invoices";