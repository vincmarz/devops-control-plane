CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE applications (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name                text NOT NULL,
    argocd_namespace    text NOT NULL,
    argocd_project      text,
    target_namespace    text,
    repo_provider       text,
    repo_url            text,
    repo_project_id     text,
    repo_default_branch text,
    repo_path           text,
    target_revision     text,
    current_revision    text,
    sync_status         text,
    health_status       text,
    last_seen_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (argocd_namespace, name)
);

CREATE TABLE change_requests (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    change_number           text UNIQUE NOT NULL,

    title                   text NOT NULL,
    application_id          uuid REFERENCES applications(id),
    application_name        text NOT NULL,
    target_environment      text NOT NULL DEFAULT 'dev',
    change_type             text NOT NULL,
    status                  text NOT NULL DEFAULT 'draft',
    risk_level              text NOT NULL DEFAULT 'medium',
    requested_by            text NOT NULL,
    description             text,
    request_payload         jsonb NOT NULL DEFAULT '{}'::jsonb,

    git_provider            text,
    gitlab_project_id       text,
    repo_url                text,
    source_branch           text,
    target_branch           text,
    commit_sha              text,
    commit_short_sha        text,
    merge_request_iid       integer,
    merge_request_url       text,
    merge_request_state     text,

    tekton_namespace        text,
    tekton_pipeline_name    text,
    tekton_pipelinerun_name text,
    tekton_status           text,
    tekton_started_at       timestamptz,
    tekton_completed_at     timestamptz,

    argocd_application      text,
    argocd_project          text,
    argocd_sync_revision    text,
    argocd_sync_status      text,
    argocd_health_status    text,
    argocd_operation_phase  text,

    runtime_namespace       text,
    runtime_status          text,

    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    submitted_at            timestamptz,
    approved_at             timestamptz,
    rejected_at             timestamptz,
    executed_at             timestamptz,
    closed_at               timestamptz,
    cancelled_at            timestamptz,
    completed_at            timestamptz,

    CONSTRAINT change_requests_status_check CHECK (
        status IN (
            'draft',
            'submitted',
            'approved',
            'rejected',
            'executing',
            'executed',
            'failed',
            'closed',
            'cancelled'
        )
    ),

    CONSTRAINT change_requests_risk_level_check CHECK (
        risk_level IN (
            'low',
            'medium',
            'high',
            'critical'
        )
    )
);

CREATE TABLE change_events (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    change_request_id   uuid NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,
    event_type          text NOT NULL,
    previous_status     text,
    new_status          text,
    message             text,
    technical_message   text,
    error_code          text,
    source              text,
    payload             jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE evidences (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    change_request_id   uuid NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,
    evidence_type       text NOT NULL,
    name                text,
    summary             text,
    content             text,
    payload             jsonb NOT NULL DEFAULT '{}'::jsonb,
    external_ref        text,
    sanitized           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_applications_name ON applications(name);
CREATE INDEX idx_applications_argocd_project ON applications(argocd_project);
CREATE INDEX idx_applications_target_namespace ON applications(target_namespace);
CREATE INDEX idx_applications_repo_project_id ON applications(repo_project_id);

CREATE INDEX idx_change_requests_application_id ON change_requests(application_id);
CREATE INDEX idx_change_requests_application_name ON change_requests(application_name);
CREATE INDEX idx_change_requests_target_environment ON change_requests(target_environment);
CREATE INDEX idx_change_requests_status ON change_requests(status);
CREATE INDEX idx_change_requests_risk_level ON change_requests(risk_level);
CREATE INDEX idx_change_requests_change_type ON change_requests(change_type);
CREATE INDEX idx_change_requests_created_at ON change_requests(created_at DESC);
CREATE INDEX idx_change_requests_source_branch ON change_requests(source_branch);
CREATE INDEX idx_change_requests_commit_sha ON change_requests(commit_sha);

CREATE INDEX idx_change_events_change_request_id ON change_events(change_request_id);
CREATE INDEX idx_change_events_event_type ON change_events(event_type);
CREATE INDEX idx_change_events_created_at ON change_events(created_at DESC);
CREATE INDEX idx_change_events_source ON change_events(source);

CREATE INDEX idx_evidences_change_request_id ON evidences(change_request_id);
CREATE INDEX idx_evidences_evidence_type ON evidences(evidence_type);
CREATE INDEX idx_evidences_created_at ON evidences(created_at DESC);
