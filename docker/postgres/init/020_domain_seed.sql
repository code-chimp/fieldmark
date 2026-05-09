-- FieldMark domain schema: reference data seed.
--
-- Infrastructure-owned per ADR-014. Runs after 010_domain_tables.sql on first
-- container startup. Populates the three reference tables that must exist
-- before any application logic can run.
--
-- UUIDs are hardcoded (not gen_random_uuid()) so they are stable across
-- `docker compose down -v && docker compose up -d` cycles. Application code
-- and fixture loaders in all three stacks may reference these UUIDs directly.
--
-- Do NOT use INSERT OR IGNORE / ON CONFLICT DO NOTHING here — if this script
-- runs twice the container was not recreated cleanly and that is a misconfiguration.

-- ---------------------------------------------------------------------------
-- TradeType
-- (§3.3 — Construction trade or subsystem; admin-managed reference data)
-- ---------------------------------------------------------------------------

INSERT INTO domain.trade_type (id, code, name, description, active) VALUES
    ('a1b2c3d4-0001-0001-0001-000000000001', 'ELEC',   'Electrical',  'Electrical systems, wiring, panels, and fixtures',               TRUE),
    ('a1b2c3d4-0001-0001-0001-000000000002', 'PLUMB',  'Plumbing',    'Plumbing, drainage, water supply, and gas lines',               TRUE),
    ('a1b2c3d4-0001-0001-0001-000000000003', 'HVAC',   'HVAC',        'Heating, ventilation, and air conditioning systems',             TRUE),
    ('a1b2c3d4-0001-0001-0001-000000000004', 'STRUCT', 'Structural',  'Structural elements including framing, foundations, and loads',  TRUE);

-- ---------------------------------------------------------------------------
-- ViolationCategory
-- (§3.6 — Admin-managed catalog of violation types)
-- Covers each severity level so compliance-scoring demos produce varied output.
-- ---------------------------------------------------------------------------

INSERT INTO domain.violation_category (id, code, name, trade_type_id, default_severity, description, active) VALUES
    -- Electrical
    ('b2c3d4e5-0002-0002-0002-000000000001',
     'ELEC_NO_GFCI',
     'Missing GFCI Protection',
     'a1b2c3d4-0001-0001-0001-000000000001',
     'High',
     'Required GFCI outlet or breaker protection is absent in a wet or outdoor location',
     TRUE),

    ('b2c3d4e5-0002-0002-0002-000000000002',
     'ELEC_OPEN_JBOX',
     'Open Junction Box',
     'a1b2c3d4-0001-0001-0001-000000000001',
     'Medium',
     'Electrical junction box left uncovered, exposing live connections',
     TRUE),

    ('b2c3d4e5-0002-0002-0002-000000000003',
     'ELEC_OVERLOAD',
     'Circuit Overload',
     'a1b2c3d4-0001-0001-0001-000000000001',
     'Critical',
     'Circuit is loaded beyond rated ampacity, presenting a fire hazard',
     TRUE),

    -- Plumbing
    ('b2c3d4e5-0002-0002-0002-000000000004',
     'PLUMB_LEAK',
     'Active Water Leak',
     'a1b2c3d4-0001-0001-0001-000000000002',
     'High',
     'Visible water leak from supply, drain, or fixture connection',
     TRUE),

    ('b2c3d4e5-0002-0002-0002-000000000005',
     'PLUMB_NO_TRAP',
     'Missing P-Trap',
     'a1b2c3d4-0001-0001-0001-000000000002',
     'Medium',
     'Drain fixture lacks the required P-trap to prevent sewer gas entry',
     TRUE),

    -- HVAC
    ('b2c3d4e5-0002-0002-0002-000000000006',
     'HVAC_FILTER',
     'Filter Not Installed',
     'a1b2c3d4-0001-0001-0001-000000000003',
     'Low',
     'Air handling unit is operating without a filter, risking equipment damage',
     TRUE),

    ('b2c3d4e5-0002-0002-0002-000000000007',
     'HVAC_DUCT_UNSECURED',
     'Unsecured Ductwork',
     'a1b2c3d4-0001-0001-0001-000000000003',
     'Medium',
     'Duct sections are not properly supported or sealed at joints',
     TRUE),

    -- Structural
    ('b2c3d4e5-0002-0002-0002-000000000008',
     'STRUCT_REBAR_EXPOSED',
     'Exposed Rebar',
     'a1b2c3d4-0001-0001-0001-000000000004',
     'High',
     'Reinforcing steel is exposed to moisture or physical contact, risking corrosion',
     TRUE),

    ('b2c3d4e5-0002-0002-0002-000000000009',
     'STRUCT_LOAD_BEARING_MOD',
     'Unauthorized Load-Bearing Modification',
     'a1b2c3d4-0001-0001-0001-000000000004',
     'Critical',
     'A load-bearing element has been altered without engineering approval',
     TRUE);

-- ---------------------------------------------------------------------------
-- ComplianceRule
-- (§3.9 — Server-evaluated rules; four canonical MVP rules)
-- Weights in ScoringPenalty parameters match §6 of the domain model.
-- ---------------------------------------------------------------------------

INSERT INTO domain.compliance_rule (id, code, name, description, rule_kind, parameters, active) VALUES
    ('c3d4e5f6-0003-0003-0003-000000000001',
     'REQUIRED_INSPECTION_PER_TRADE',
     'Required Inspection Per Trade',
     'Each trade assigned to a project must have at least one Completed inspection with an outcome of Pass or Conditional before the project may be closed.',
     'ClosureGate',
     '{"min_completed": 1, "acceptable_outcomes": ["Pass", "Conditional"]}',
     TRUE),

    ('c3d4e5f6-0003-0003-0003-000000000002',
     'OPEN_VIOLATION_GATE',
     'Open Violation Closure Gate',
     'A project may not be closed while any violation is in state Open or InProgress.',
     'ClosureGate',
     '{"blocking_statuses": ["Open", "InProgress"]}',
     TRUE),

    ('c3d4e5f6-0003-0003-0003-000000000003',
     'OVERDUE_VIOLATION_PENALTY',
     'Overdue Violation Scoring Penalty',
     'Each overdue violation (due_at in the past, status Open or InProgress) reduces the compliance score by the weight for its severity.',
     'ScoringPenalty',
     '{"weights": {"Low": 3, "Medium": 7, "High": 15, "Critical": 30}}',
     TRUE),

    ('c3d4e5f6-0003-0003-0003-000000000004',
     'OPEN_VIOLATION_PENALTY',
     'Open Violation Scoring Penalty',
     'Each open (not yet overdue) violation reduces the compliance score by the weight for its severity.',
     'ScoringPenalty',
     '{"weights": {"Low": 1, "Medium": 3, "High": 8, "Critical": 15}}',
     TRUE);
