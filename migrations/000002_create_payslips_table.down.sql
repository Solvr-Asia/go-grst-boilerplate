-- 000002_create_payslips_table.down.sql
-- Drop payslips table

DROP TRIGGER IF EXISTS update_payslips_updated_at ON payslips;
DROP INDEX IF EXISTS idx_payslips_deleted_at;
DROP INDEX IF EXISTS idx_payslips_status;
DROP INDEX IF EXISTS idx_payslips_year_month;
DROP INDEX IF EXISTS idx_payslips_employee_id;
DROP TABLE IF EXISTS payslips;
