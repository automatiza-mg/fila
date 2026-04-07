import { env } from "$env/dynamic/public";
import {
  type AlterarSenha,
  type Analista,
  type Cadastrar,
  type Credenciais,
  type CriarAnalista,
  type CriarProcesso,
  type CriarUsuario,
  type Documento,
  type ErrorResponse,
  type Escopo,
  type ListProcessosAposentadoriaFilters,
  type ListSolicitacoesPrioridadeFilters,
  type Paginated,
  type Processo,
  type ProcessoAposentadoria,
  type ProcessoHistorico,
  type RecuperarSenha,
  type RedefinirSenha,
  type SolicitacaoPrioridade,
  type SolicitarPrioridade,
  type Token,
  type Unidade,
  type Usuario,
} from "./types";

export class ApiError extends Error {
  constructor(
    public message: string,
    public status: number,
    public response: ErrorResponse,
  ) {
    super(message);
  }
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const data = (await res.json()) as ErrorResponse;
    throw new ApiError(data.message, res.status, data);
  }

  return await res.json();
}

export async function tokenInfo(
  token: string,
  escopo: Escopo,
): Promise<Usuario> {
  const q = new URLSearchParams({
    token,
    escopo,
  });

  const res = await fetch(
    `${env.PUBLIC_API_URL}/api/v1/auth/token?${q.toString()}`,
  );
  return await handleResponse<Usuario>(res);
}

export async function entrar(data: Credenciais): Promise<Token> {
  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/entrar`, {
    method: "POST",
    body: JSON.stringify(data),
  });
  return await handleResponse<Token>(res);
}

export async function cadastrar(data: Cadastrar): Promise<void> {
  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/cadastrar`, {
    method: "POST",
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const body = (await res.json()) as ErrorResponse;
    throw new ApiError(body.message, res.status, body);
  }
}

export async function recuperarSenha(data: RecuperarSenha): Promise<void> {
  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/recuperar-senha`, {
    method: "POST",
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const body = (await res.json()) as ErrorResponse;
    throw new ApiError(body.message, res.status, body);
  }
}

export async function redefinirSenha(data: RedefinirSenha): Promise<void> {
  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/redefinir-senha`, {
    method: "POST",
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const body = (await res.json()) as ErrorResponse;
    throw new ApiError(body.message, res.status, body);
  }
}

export class Client {
  constructor(private readonly authToken: string) {}

  private async request<T>(
    endpoint: string,
    options?: RequestInit,
  ): Promise<T> {
    const url = `${env.PUBLIC_API_URL}${endpoint}`;
    const res = await fetch(url, {
      ...options,
      headers: {
        ...options?.headers,
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    return await handleResponse<T>(res);
  }

  private async requestVoid(
    endpoint: string,
    options?: RequestInit,
  ): Promise<void> {
    const url = `${env.PUBLIC_API_URL}${endpoint}`;
    const res = await fetch(url, {
      ...options,
      headers: {
        ...options?.headers,
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    if (!res.ok) {
      const body = (await res.json()) as ErrorResponse;
      throw new ApiError(body.message, res.status, body);
    }
  }

  private async requestBlob(
    endpoint: string,
    options?: RequestInit,
  ): Promise<Blob> {
    const url = `${env.PUBLIC_API_URL}${endpoint}`;
    const res = await fetch(url, {
      ...options,
      headers: {
        ...options?.headers,
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    if (!res.ok) {
      const body = (await res.json()) as ErrorResponse;
      throw new ApiError(body.message, res.status, body);
    }

    return res.blob();
  }

  async usuarioAtual(): Promise<Usuario> {
    return this.request<Usuario>("/api/v1/auth/me");
  }

  async analistaAtual(): Promise<Analista> {
    return this.request<Analista>("/api/v1/auth/me/analista");
  }

  async alterarSenha(data: AlterarSenha): Promise<void> {
    return this.requestVoid("/api/v1/auth/alterar-senha", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async listarUsuarios(): Promise<Usuario[]> {
    return this.request<Usuario[]>("/api/v1/usuarios");
  }

  async criarUsuario(data: CriarUsuario): Promise<Usuario> {
    return this.request<Usuario>("/api/v1/usuarios", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getUsuario(id: number): Promise<Usuario> {
    return this.request<Usuario>(`/api/v1/usuarios/${id}`);
  }

  async deletarUsuario(id: number): Promise<void> {
    return this.requestVoid(`/api/v1/usuarios/${id}`, {
      method: "DELETE",
    });
  }

  async enviarCadastro(usuarioId: number): Promise<void> {
    return this.requestVoid(`/api/v1/usuarios/${usuarioId}/enviar-cadastro`, {
      method: "POST",
    });
  }

  async getAnalista(usuarioId: number): Promise<Analista> {
    return this.request<Analista>(`/api/v1/usuarios/${usuarioId}/analista`);
  }

  async criarAnalista(
    usuarioId: number,
    data: CriarAnalista,
  ): Promise<Analista> {
    return this.request<Analista>(`/api/v1/usuarios/${usuarioId}/analista`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async afastarAnalista(usuarioId: number): Promise<void> {
    return this.requestVoid(`/api/v1/usuarios/${usuarioId}/analista/afastar`, {
      method: "POST",
    });
  }

  async retornarAnalista(usuarioId: number): Promise<void> {
    return this.requestVoid(`/api/v1/usuarios/${usuarioId}/analista/retornar`, {
      method: "POST",
    });
  }

  async getAnalistaProcessoAtribuido(
    usuarioId: number,
  ): Promise<ProcessoAposentadoria> {
    return this.request<ProcessoAposentadoria>(
      `/api/v1/usuarios/${usuarioId}/analista/processo`,
    );
  }

  async listarAnalistas(): Promise<Analista[]> {
    return this.request<Analista[]>("/api/v1/analistas");
  }

  async listarProcessos(page = 1): Promise<Paginated<Processo>> {
    const q = new URLSearchParams({ page: String(page) });
    return this.request<Paginated<Processo>>(`/api/v1/processos?${q}`);
  }

  async criarProcesso(data: CriarProcesso): Promise<Processo> {
    return this.request<Processo>("/api/v1/processos", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getProcesso(id: string): Promise<Processo> {
    return this.request<Processo>(`/api/v1/processos/${id}`);
  }

  async getProcessoDocumentos(id: string): Promise<Documento[]> {
    return this.request<Documento[]>(`/api/v1/processos/${id}/documentos`);
  }

  async listarAposentadoria(
    filters: ListProcessosAposentadoriaFilters = {},
  ): Promise<Paginated<ProcessoAposentadoria>> {
    const q = new URLSearchParams({ page: String(filters.page ?? 1) });
    if (filters.numero) {
      q.set("numero", filters.numero);
    }
    return this.request<Paginated<ProcessoAposentadoria>>(
      `/api/v1/aposentadoria?${q}`,
    );
  }

  async getAposentadoria(id: number): Promise<ProcessoAposentadoria> {
    return this.request<ProcessoAposentadoria>(`/api/v1/aposentadoria/${id}`);
  }

  async getHistorico(id: number): Promise<ProcessoHistorico[]> {
    return this.request<ProcessoHistorico[]>(
      `/api/v1/aposentadoria/${id}/historico`,
    );
  }

  async solicitarPrioridade(
    paId: number,
    data: SolicitarPrioridade,
  ): Promise<SolicitacaoPrioridade> {
    return this.request<SolicitacaoPrioridade>(
      `/api/v1/aposentadoria/${paId}/prioridade`,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
    );
  }

  async listarSolicitacoesPrioridade(
    filters: ListSolicitacoesPrioridadeFilters = {},
  ): Promise<Paginated<SolicitacaoPrioridade>> {
    const q = new URLSearchParams({ page: String(filters.page ?? 1) });
    if (filters.status) {
      q.set("status", filters.status);
    }
    if (filters.numero) {
      q.set("numero", filters.numero);
    }

    return this.request<Paginated<SolicitacaoPrioridade>>(
      `/api/v1/solicitacoes-prioridade?${q}`,
    );
  }

  async getSolicitacaoPrioridade(id: number): Promise<SolicitacaoPrioridade> {
    return this.request<SolicitacaoPrioridade>(
      `/api/v1/solicitacoes-prioridade/${id}`,
    );
  }

  async aprovarSolicitacaoPrioridade(id: number): Promise<void> {
    return this.requestVoid(`/api/v1/solicitacoes-prioridade/${id}/aprovar`, {
      method: "POST",
    });
  }

  async negarSolicitacaoPrioridade(id: number): Promise<void> {
    return this.requestVoid(`/api/v1/solicitacoes-prioridade/${id}/negar`, {
      method: "POST",
    });
  }

  async listarUnidadesSei(): Promise<Unidade[]> {
    return this.request<Unidade[]>("/api/v1/unidades");
  }

  async meuProcessoAtribuido(): Promise<ProcessoAposentadoria> {
    return this.request<ProcessoAposentadoria>("/api/v1/meu-processo");
  }

  async getAposentadoriaPreview(paId: number): Promise<Blob> {
    return this.requestBlob(`/api/v1/aposentadoria/${paId}/preview`);
  }

  async marcarLeituraInvalida(
    paId: number,
    motivo: string,
  ): Promise<void> {
    return this.requestVoid(
      `/api/v1/aposentadoria/${paId}/leitura-invalida`,
      {
        method: "POST",
        body: JSON.stringify({ motivo }),
      },
    );
  }
}
