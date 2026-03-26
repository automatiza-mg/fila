export type Paginated<T> = {
  data: T[];
  limit: number;
  current_page: number;
  total_count: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
};

export type ErrorResponse = {
  message: string;
  errors?: Record<string, string>;
};

export type Credenciais = {
  cpf: string;
  senha: string;
};

export type Token = {
  token: string;
  expira: string;
};

export type Escopo = "reset-senha" | "setup";

export type Papel = "ADMIN" | "ANALISTA" | "GESTOR" | "SUBSECRETARIO";

export type Cadastrar = {
  token: string;
  senha: string;
  confirmar_senha: string;
};

export type CriarUsuario = {
  nome: string;
  cpf: string;
  email: string;
  papel: Papel;
};

export type Pendencia = {
  slug: string;
  titulo: string;
};

export type Usuario = {
  id: number;
  nome: string;
  cpf: string;
  email: string;
  email_verificado: boolean;
  papel?: Papel;
  pendencias: Pendencia[];
};

export type MetadadoIA = {
  judicial: boolean;
  invalidez: boolean;
  aposentadoria: boolean;
  cpf_requerente: string;
  data_requerimento: string;
  cpf_responsavel_diligencia: string;
  data_nascimento_requerente: string;
};

export type Processo = {
  id: string;
  numero: string;
  status: string;
  resumo: string;
  link_acesso: string;
  sei_unidade_id: string;
  sei_unidade_sigla: string;
  aposentadoria: boolean | null;
  preview_hash: string | null;
  analisado_em: string | null;
  metadados_ia: MetadadoIA | null;
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
  data_nascimento_requerente: string;
  judicial: boolean;
  invalidez: boolean;
  prioridade: boolean;
  score: number;
  status: StatusProcessoAposentadoria;
  analista_id: number | null;
  analista: string | null;
  criado_em: string;
  atualizado_em: string;
};

export type Analista = {
  usuario_id: number;
  orgao: string;
  sei_unidade_id: string;
  sei_unidade_sigla: string;
  afastado: boolean;
  ultima_atribuicao_em: string | null;
};

export type Unidade = {
  id: string;
  sigla: string;
  descricao: string;
};

export type CriarAnalista = {
  sei_unidade_id: string;
  orgao: string;
};

export type Assinatura = {
  nome: string;
  cpf: string;
};

export type Documento = {
  id: number;
  numero: string;
  tipo: string;
  conteudo: string;
  content_type: string;
  data: string;
  unidade_geradora: string;
  assinaturas: Assinatura[];
};

export type ProcessoHistorico = {
  status_anterior: StatusProcessoAposentadoria | null;
  status_novo: StatusProcessoAposentadoria;
  usuario_id: number | null;
  observacao: string | null;
  alterado_em: string;
};

export type SolicitacaoPrioridade = {
  id: number;
  numero_processo: string;
  processo_aposentadoria_id: number;
  justificativa: string;
  status: string;
  criado_em: string;
  atualizado_em: string;
};

export type ListProcessosAposentadoriaFilters = {
  page?: number;
  numero?: string;
};

export type ListSolicitacoesPrioridadeFilters = {
  page?: number;
  status?: string;
  numero?: string;
};

export type SolicitarPrioridade = {
  justificativa: string;
};

export type RecuperarSenha = {
  cpf: string;
};

export type RedefinirSenha = {
  token: string;
  senha: string;
  confirmar_senha: string;
};

export type CriarProcesso = {
  numero: string;
};
