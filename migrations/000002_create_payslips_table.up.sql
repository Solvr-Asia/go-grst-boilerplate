-- 000002_create_payslips_table.up.sql
-- Create payslips table

CREATE TABLE IF NOT EXISTS payslips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL CHECK (month >= 1 AND month <= 12),
    basic_salary DECIMAL(15, 2) NOT NULL DEFAULT 0,
    allowances DECIMAL(15, 2) NOT NULL DEFAULT 0,
    deductions DECIMAL(15, 2) NOT NULL DEFAULT 0,
    gross_salary DECIMAL(15, 2) NOT NULL DEFAULT 0,
    net_salary DECIMAL(15, 2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    paid_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(employee_id, year, month)
);

-- Create indexes
CREATE INDEX idx_payslips_employee_id ON payslips(employee_id);
CREATE INDEX idx_payslips_year_month ON payslips(year, month);
CREATE INDEX idx_payslips_status ON payslips(status);
CREATE INDEX idx_payslips_deleted_at ON payslips(deleted_at);

-- Apply updated_at trigger
CREATE TRIGGER update_payslips_updated_at
    BEFORE UPDATE ON payslips
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
