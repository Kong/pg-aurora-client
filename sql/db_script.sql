BEGIN;
DROP TABLE IF EXISTS "default_runtime_group_relations";

CREATE TABLE "default_runtime_group_relations"
(
    org_id           uuid not null,
    control_plane_id uuid PRIMARY KEY
);

insert into default_runtime_group_relations values('001c5e3c-6086-4c66-b0de-8b5eba9fa655', '6e222976-c148-4fb6-91bd-942a5ab78a92');
insert into default_runtime_group_relations values('001c5e3c-6086-4c66-b0de-8b5eba9fa655', '7e06ff51-d592-4eef-ba89-aa352ed4786a');
insert into default_runtime_group_relations values('001c5e3c-6086-4c66-b0de-8b5eba9fa655', '712a9488-6d3f-487f-ab33-16e01a033ccd');
insert into default_runtime_group_relations values('0028abbd-0494-48e3-b686-44ca7598510c', 'eb79032f-7b2f-42f8-ba5d-ba7fe71e8c96');
insert into default_runtime_group_relations values('0028abbd-0494-48e3-b686-44ca7598510c', '74ae6324-3dac-4ed8-a4f5-df57768a5080');

-- RLS support functions

-- setting the tenant in session scope
CREATE OR REPLACE FUNCTION set_tenant(tenant TEXT) RETURNS VOID AS
$$
DECLARE
    v_value UUID;
BEGIN
    v_value := tenant::UUID;
    PERFORM set_config('app.current_tenant', tenant, false);
END;

$$ LANGUAGE plpgsql SECURITY DEFINER
                    STABLE;


-- un-setting the tenant in session scope
CREATE OR REPLACE FUNCTION unset_tenant() RETURNS VOID AS
$$
BEGIN
    PERFORM set_config('app.current_tenant', '', false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER
                    STABLE;

-- Get the current tenant value that is set in the session context
CREATE OR REPLACE FUNCTION get_tenant() RETURNS UUID AS
$$
DECLARE
    v_value   UUID;
    v_s_value TEXT;
BEGIN
    v_s_value := current_setting('app.current_tenant', true);
    IF v_s_value = '' THEN
        RETURN NULL;
    END IF;
    v_value := current_setting('app.current_tenant', true):: UUID;
    RETURN v_value;

END ;
$$ LANGUAGE plpgsql SECURITY DEFINER
                    STABLE;

-- Create user that is filtered by RLS
CREATE USER koko WITH LOGIN PASSWORD 'koko';
GRANT USAGE ON SCHEMA public TO koko;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO koko;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO koko;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO koko;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO koko;

-- Create user that is exempt by RLS
CREATE USER kokobatch WITH LOGIN PASSWORD 'koko';
GRANT USAGE ON SCHEMA public TO kokobatch;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO kokobatch;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO kokobatch;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO kokobatch;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO kokobatch;
ALTER ROLE kokobatch BYPASSRLS;

-- Enabling RLS on the table and setting the policy
ALTER TABLE default_runtime_group_relations
    ENABLE ROW LEVEL SECURITY;

CREATE POLICY default_runtime_group_relations_policy ON default_runtime_group_relations
    USING (org_id = get_tenant());
commit;
