CREATE TABLE change_runtime_states (
    change_request_id uuid PRIMARY KEY
        REFERENCES change_requests(id)
        ON DELETE CASCADE,

    source_state jsonb NOT NULL DEFAULT '{}'::jsonb,
    gitops_state jsonb NOT NULL DEFAULT '{}'::jsonb,
    tekton_state jsonb NOT NULL DEFAULT '{}'::jsonb,
    argocd_state jsonb NOT NULL DEFAULT '{}'::jsonb,
    runtime_state jsonb NOT NULL DEFAULT '{}'::jsonb,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_change_runtime_states_updated_at
    ON change_runtime_states(updated_at DESC);
