-- +goose Up
-- +goose StatementBegin
CREATE TABLE "tokens" (
    "hash" BYTEA PRIMARY KEY,
    "usuario_id" BIGINT NOT NULL,
    "escopo" TEXT NOT NULL,
    "expira_em" TIMESTAMPTZ NOT NULL,
    FOREIGN KEY ("usuario_id") REFERENCES "usuarios"("id") ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "tokens";
-- +goose StatementEnd
