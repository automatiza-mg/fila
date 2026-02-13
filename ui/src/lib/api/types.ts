export type Papel = "ADMIN" | "SUBSECRETARIO" | "GESTOR" | "ANALISTA";

export type Escopo = "setup" | "reset-senha" | "auth";

export type StatusProcesso =
  | "ANALISE_PENDENTE"
  | "EM_ANALISE"
  | "EM_DILIGENCIA"
  | "RETORNO_DILIGENCIA"
  | "CONCLUIDO"
  | "LEITURA_INVALIDA";

export type Orgao = "SEPLAG" | "SEE";

export interface PendingAction {
  slug: string;
  titulo: string;
}

export interface Usuario {
  id: number;
  nome: string;
  cpf: string;
  email: string;
  email_verificado: boolean;
  papel?: string;
  pendencias: PendingAction[];
}

export interface Token {
  token: string;
  expira: string;
}

export interface Analista {
  usuario_id: number;
  orgao: Orgao;
  sei_unidade_id: string;
  sei_unidade_sigla: string;
  afastado: boolean;
  ultima_atribuicao_em: string | null;
}

export interface Processo {
  id: string;
  numero: string;
  status: string;
  link_acesso: string;
  sei_unidade_id: string;
  sei_unidade_sigla: string;
  aposentadoria: boolean | null;
  analisado_em: string | null;
  metadados_ia: unknown;
  criado_em: string;
  atualizado_em: string;
}

export interface Assinatura {
  nome: string;
  cpf: string;
}

export interface Documento {
  id: number;
  numero: string;
  tipo: string;
  conteudo: string;
  link_acesso: string;
  data: string;
  unidade_geradora: string;
  assinaturas: Assinatura[];
}

export interface ProcessoAposentadoria {
  id: number;
  processo_id: string;
  numero: string;
  data_requerimento: string;
  cpf_requerente: string;
  data_nascimento_requerente: string;
  invalidez: boolean;
  judicial: boolean;
  prioridade: boolean;
  score: number;
  status: StatusProcesso;
  analista_id: number | null;
  criado_em: string;
  atualizado_em: string;
}

export interface HistoricoStatusProcesso {
  status_anterior: string | null;
  status_novo: string;
  usuario_id: number | null;
  observacao: string | null;
  alterado_em: string;
}

export interface UnidadeSei {
  id: string;
  sigla: string;
  descricao: string;
}

export interface UnidadeGeradora {
  sigla_unidade: string;
  id_unidade: string;
}

export interface DatalakeProcesso {
  numero_processo: string;
  sigla_unidade: string;
  data_recebimento: string;
  unidade_geradora: UnidadeGeradora;
}

export interface Servidor {
  id_pessoa: number;
  nome: string;
  masp: string;
  cpf: string;
  sexo: string;
  data_nascimento: string;
  possui_deficiencia: boolean;
}

export interface EntrarRequest {
  cpf: string;
  senha: string;
}

export interface CadastrarRequest {
  token: string;
  senha: string;
  confirmar_senha: string;
}

export interface RecuperarSenhaRequest {
  cpf: string;
}

export interface RedefinirSenhaRequest {
  token: string;
  senha: string;
  confirmar_senha: string;
}

export interface UsuarioCreateRequest {
  nome: string;
  cpf: string;
  email: string;
  papel: "ANALISTA" | "GESTOR" | "SUBSECRETARIO";
}

export interface AnalistaCreateRequest {
  sei_unidade_id: string;
  orgao: Orgao;
}

export interface ProcessoCreateRequest {
  numero: string;
}

export interface PaginatedResult<T> {
  data: T[];
  limit: number;
  current_page: number;
  total_count: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
}

export interface ApiErrorBody {
  message: string;
  errors?: Record<string, string>;
}
