CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION generate_termo_busca(_nome VARCHAR, _apelido VARCHAR, _stack TEXT[])
    RETURNS TEXT AS $$
    BEGIN
    RETURN _nome || _apelido || _stack;
    END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE TABLE IF NOT EXISTS pessoa (
    id uuid UNIQUE NOT NULL,
    apelido TEXT UNIQUE NOT NULL,
    nome TEXT NOT NULL,
    nascimento DATE NOT NULL,
    stack TEXT[],
    termo_busca text GENERATED ALWAYS AS (generate_termo_busca(nome, apelido, stack)) STORED
);

CREATE INDEX IF NOT EXISTS idx_pessoa_termo_busca ON public.pessoa USING gist (termo_busca public.gist_trgm_ops (siglen='64'));

CREATE UNIQUE INDEX IF NOT EXISTS pessoa_apelido_index ON public.pessoa USING btree (apelido);