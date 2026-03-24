-- +goose Up
CREATE UNIQUE INDEX "documentos_numero_processo_id_key" ON "documentos" ("numero", "processo_id");

-- +goose Down
DROP INDEX "documentos_numero_processo_id_key";
