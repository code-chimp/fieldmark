-- fiber_auth bootstrap. Framework-local per ADR-012. Owned by the Go stack.
-- Re-runnable: every statement is IF NOT EXISTS.
-- DO NOT issue any DDL against domain.* here.

CREATE TABLE IF NOT EXISTS fiber_auth.users (
    id           uuid          PRIMARY KEY,
    username     varchar(64)   NOT NULL UNIQUE,
    display_name varchar(128)  NOT NULL,
    created_at   timestamptz   NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fiber_auth.user_roles (
    user_id uuid        NOT NULL REFERENCES fiber_auth.users(id) ON DELETE CASCADE,
    role    varchar(64) NOT NULL,
    PRIMARY KEY (user_id, role),
    CONSTRAINT user_roles_role_check CHECK (role IN (
        'ADMIN', 'COMPLIANCE_OFFICER', 'INSPECTOR', 'SITE_SUPERVISOR', 'EXECUTIVE'
    ))
);

CREATE INDEX IF NOT EXISTS idx_fiber_auth_user_roles_user_id
    ON fiber_auth.user_roles(user_id);
