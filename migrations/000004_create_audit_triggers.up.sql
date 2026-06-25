-- Automatic audit trail triggers
-- These ensure every INSERT/UPDATE/DELETE on tracked tables is logged
-- without any application code changes.

CREATE OR REPLACE FUNCTION log_audit()
RETURNS TRIGGER AS $$
DECLARE
    next_code VARCHAR(20);
    next_val INT;
    changed_by_id UUID;
BEGIN
    -- Get the next audit code
    UPDATE entity_sequences SET next_value = next_value + 1 WHERE entity_name = 'audit_trail' RETURNING next_value - 1 INTO next_val;
    next_code := 'AUD' || LPAD(next_val::TEXT, 6, '0');

    -- Try to get changed_by from session variable (set by application)
    BEGIN
        changed_by_id := current_setting('app.changed_by')::UUID;
    EXCEPTION WHEN OTHERS THEN
        changed_by_id := NULL;
    END;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
        VALUES (next_code, TG_TABLE_NAME, NEW.id, 'CREATE', '', '', row_to_json(NEW)::TEXT, changed_by_id);
        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
        VALUES (next_code, TG_TABLE_NAME, OLD.id, 'DELETE', '', row_to_json(OLD)::TEXT, '', changed_by_id);
        RETURN OLD;

    ELSIF TG_OP = 'UPDATE' THEN
        -- Log each changed field
        IF NEW.code IS DISTINCT FROM OLD.code THEN
            INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
            VALUES (next_code, TG_TABLE_NAME, NEW.id, 'UPDATE', 'code', OLD.code, NEW.code, changed_by_id);
        END IF;
        IF NEW.name IS DISTINCT FROM OLD.name THEN
            INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
            VALUES (next_code, TG_TABLE_NAME, NEW.id, 'UPDATE', 'name', OLD.name, NEW.name, changed_by_id);
        END IF;
        IF NEW.status IS DISTINCT FROM OLD.status THEN
            INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
            VALUES (next_code, TG_TABLE_NAME, NEW.id, 'UPDATE', 'status', OLD.status, NEW.status, changed_by_id);
        END IF;
        IF NEW.risk_level IS DISTINCT FROM OLD.risk_level THEN
            INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
            VALUES (next_code, TG_TABLE_NAME, NEW.id, 'UPDATE', 'risk_level', OLD.risk_level, NEW.risk_level, changed_by_id);
        END IF;
        IF NEW.assigned_dept_manager_id IS DISTINCT FROM OLD.assigned_dept_manager_id THEN
            INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
            VALUES (next_code, TG_TABLE_NAME, NEW.id, 'UPDATE', 'assigned_dept_manager_id', OLD.assigned_dept_manager_id::TEXT, NEW.assigned_dept_manager_id::TEXT, changed_by_id);
        END IF;
        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Vendor triggers
CREATE TRIGGER trg_vendors_audit_insert
    AFTER INSERT ON vendors
    FOR EACH ROW EXECUTE FUNCTION log_audit();
CREATE TRIGGER trg_vendors_audit_update
    AFTER UPDATE ON vendors
    FOR EACH ROW EXECUTE FUNCTION log_audit();
CREATE TRIGGER trg_vendors_audit_delete
    BEFORE DELETE ON vendors
    FOR EACH ROW EXECUTE FUNCTION log_audit();

-- Risk assessment triggers
CREATE TRIGGER trg_risk_assessments_audit_insert
    AFTER INSERT ON risk_assessments
    FOR EACH ROW EXECUTE FUNCTION log_audit();
CREATE TRIGGER trg_risk_assessments_audit_update
    AFTER UPDATE ON risk_assessments
    FOR EACH ROW EXECUTE FUNCTION log_audit();
CREATE TRIGGER trg_risk_assessments_audit_delete
    BEFORE DELETE ON risk_assessments
    FOR EACH ROW EXECUTE FUNCTION log_audit();

-- Compliance record triggers
CREATE TRIGGER trg_compliance_records_audit_insert
    AFTER INSERT ON compliance_records
    FOR EACH ROW EXECUTE FUNCTION log_audit();
CREATE TRIGGER trg_compliance_records_audit_update
    AFTER UPDATE ON compliance_records
    FOR EACH ROW EXECUTE FUNCTION log_audit();
CREATE TRIGGER trg_compliance_records_audit_delete
    BEFORE DELETE ON compliance_records
    FOR EACH ROW EXECUTE FUNCTION log_audit();

-- Contract triggers
CREATE TRIGGER trg_contracts_audit_insert
    AFTER INSERT ON contracts
    FOR EACH ROW EXECUTE FUNCTION log_audit();
CREATE TRIGGER trg_contracts_audit_update
    AFTER UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION log_audit();
CREATE TRIGGER trg_contracts_audit_delete
    BEFORE DELETE ON contracts
    FOR EACH ROW EXECUTE FUNCTION log_audit();