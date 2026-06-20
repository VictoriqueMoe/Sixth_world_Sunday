-- +goose Up
-- +goose StatementBegin
CREATE TABLE vault_folders (
    id         uuid PRIMARY KEY,
    parent_id  uuid REFERENCES vault_folders(id) ON DELETE CASCADE,
    name       text NOT NULL,
    locked     boolean NOT NULL DEFAULT false,
    created_by uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_vault_folders_parent_id ON vault_folders(parent_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE vault_files (
    id            uuid PRIMARY KEY,
    folder_id     uuid REFERENCES vault_folders(id) ON DELETE CASCADE,
    original_name text NOT NULL,
    stored_name   text NOT NULL,
    mime          text NOT NULL,
    size          bigint NOT NULL,
    locked        boolean NOT NULL DEFAULT false,
    uploaded_by   uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at    timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_vault_files_folder_id ON vault_files(folder_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_vault_files_uploaded_by ON vault_files(uploaded_by);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vault_files;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS vault_folders;
-- +goose StatementEnd
