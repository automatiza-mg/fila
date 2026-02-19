export type Paginated<T> = {
  data: T[];
  limit: number;
  current_page: number;
  total_count: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;
};

export type Papel = "ADMIN" | "ANALISTA" | "GESTOR" | "SUBSECRETARIO";

export type Usuario = {
  id: number;
  nome: string;
  cpf: string;
  email: string;
  email_verificado: boolean;
  papel?: Papel;
  pendencias: any[];
};

export type Processo = {
  id: string;
  numero: string;
  status: string;
  link_acesso: string;
  sei_unidade_id: string;
  sei_unidade_sigla: string;
  aposentadoria: boolean;
  analisado_em: string;
  metadados_ia?: {
    judicial: boolean;
    invalidez: boolean;
    aposentadoria: boolean;
    cpf_requerente: string;
    data_requerimento: string;
    cpf_responsavel_diligencia: string;
    data_nascimento_requerente: string;
  };
  criado_em: string;
  atualizado_em: string;
};

export type StatusProcessoAposentadoria =
  | "ANALISE_PENDENTE"
  | "EM_ANALISE"
  | "EM_DILIGENCIA"
  | "RETORNO_DILIGENCIA"
  | "LEITURA_INVALIDA"
  | "CONCLUIDO";

export type ProcessoAposentadoria = {
  id: number;
  processo_id: string;
  numero: string;
  data_requerimento: string;
  cpf_requerente: string;
  judicial: boolean;
  invalidez: boolean;
  prioridade: boolean;
  score: number;
  status: StatusProcessoAposentadoria;
  analista_id: number | null;
  criado_em: string;
  atualizado_em: string;
};
