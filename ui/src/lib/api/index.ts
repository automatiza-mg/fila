import { env } from "$env/dynamic/public";
import type {
  Analista,
  Cadastrar,
  Credenciais,
  Documento,
  ErrorResponse,
  Escopo,
  Paginated,
  Processo,
  ProcessoAposentadoria,
  ProcessoHistorico,
  Token,
  Usuario,
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

export async function authenticate({
  cpf,
  senha,
}: Credenciais): Promise<Token> {
  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/entrar`, {
    method: "POST",
    body: JSON.stringify({
      cpf,
      senha,
    }),
  });

  return await handleResponse<Token>(res);
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

export async function cadastrar(data: Cadastrar) {
  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/cadastrar`, {
    method: "POST",
    body: JSON.stringify(data),
  });
  return await handleResponse(res);
}

export async function getAnalistaAtual(token: string): Promise<Analista> {
  const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/auth/me/analista`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return await handleResponse<Analista>(res);
}

export class Client {
  private readonly baseUrl: string;

  constructor(public authToken: string) {
    this.baseUrl = `${env.PUBLIC_API_URL}/api/v1`;
  }

  private async request<T>(
    url: string,
    options?: RequestInit,
  ): Promise<T> {
    const res = await fetch(url, {
      ...options,
      headers: {
        ...options?.headers,
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    return await handleResponse<T>(res);
  }

  async usuarioAtual(): Promise<Usuario> {
    return await this.request<Usuario>(`${this.baseUrl}/auth/me`);
  }

  async listarProcessos(): Promise<Paginated<Processo>> {
    return await this.request<Paginated<Processo>>(
      `${this.baseUrl}/processos`,
    );
  }

  async listarUsuarios(): Promise<Usuario[]> {
    return await this.request<Usuario[]>(`${this.baseUrl}/usuarios`);
  }

  async listarAposentadoria(): Promise<Paginated<ProcessoAposentadoria>> {
    return await this.request<Paginated<ProcessoAposentadoria>>(
      `${this.baseUrl}/aposentadoria`,
    );
  }

  async getAposentadoria(id: number): Promise<ProcessoAposentadoria> {
    return await this.request<ProcessoAposentadoria>(
      `${this.baseUrl}/aposentadoria/${id}`,
    );
  }

  async getProcesso(id: string): Promise<Processo> {
    return await this.request<Processo>(`${this.baseUrl}/processos/${id}`);
  }

  async getProcessoDocumentos(id: string): Promise<Documento[]> {
    return await this.request<Documento[]>(
      `${this.baseUrl}/processos/${id}/documentos`,
    );
  }

  async getProcessoAposentadoriaHistorico(
    id: number,
  ): Promise<ProcessoHistorico[]> {
    return await this.request<ProcessoHistorico[]>(
      `${this.baseUrl}/aposentadoria/${id}/historico`,
    );
  }
}
