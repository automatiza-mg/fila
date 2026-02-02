-- +goose Up
-- +goose StatementBegin
CREATE TABLE "usuarios" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "nome" TEXT NOT NULL,
    "cpf" TEXT NOT NULL UNIQUE,
    "email" TEXT NOT NULL UNIQUE,
    "email_verificado" BOOLEAN NOT NULL DEFAULT FALSE,
    "hash_senha" TEXT,
    "papel" TEXT,
    "criado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "atualizado_em" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "analistas" (
    "usuario_id" BIGINT PRIMARY KEY,
    "orgao" TEXT NOT NULL,
    "sei_unidade_id" TEXT NOT NULL UNIQUE,
    "afastado" BOOLEAN NOT NULL DEFAULT FALSE,
    "ultima_atribuicao_em" TIMESTAMPTZ,
    FOREIGN KEY ("usuario_id") REFERENCES "usuarios"("id") ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "analistas";
DROP TABLE "usuarios";
-- +goose StatementEnd
