-- +goose Up
-- +goose StatementBegin
ALTER TABLE "processos" ADD COLUMN "metadados_ia" JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE "processos" ADD COLUMN "analisado_em" TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "processos" DROP COLUMN "analisado_em";
ALTER TABLE "processos" DROP COLUMN "metadados_ia";
-- +goose StatementEnd
