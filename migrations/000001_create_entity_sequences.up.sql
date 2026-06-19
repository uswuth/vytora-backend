CREATE TABLE
    IF NOT EXISTS entity_sequences (
        entity_name VARCHAR(50) PRIMARY KEY,
        next_value INTEGER NOT NULL DEFAULT 1
    );

INSERT INTO
    entity_sequences (entity_name, next_value)
VALUES
    ('user', 1),
    ('vendor', 1),
    ('vendor_contact', 1),
    ('risk_assessment', 1),
    ('compliance_record', 1),
    ('contract', 1),
    ('audit_trail', 1),
    ('category', 1);
