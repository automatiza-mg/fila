import { env } from "$env/dynamic/public";
import type {
  CriarUsuario,
  ErrorResponse,
  Paginated,
  Processo,
  ProcessoAposentadoria,
  ProcessoHistorico,
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

  async usuarioAtual(): Promise<Usuario> {
    return this.request("/api/v1/auth/me");
  }

  async listarProcessos(): Promise<Paginated<Processo>> {
    return await this.request<Paginated<Processo>>("/api/v1/processos");
  }

  async listarUsuarios(): Promise<Usuario[]> {
    return await this.request<Usuario[]>("/api/v1/usuarios");
  }

  async criarUsuario(data: CriarUsuario): Promise<Usuario> {
    return await this.request<Usuario>("/api/v1/usuarios", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async listarAposentadoria(): Promise<Paginated<ProcessoAposentadoria>> {
    return await this.request<Paginated<ProcessoAposentadoria>>(
      `/api/v1/aposentadoria`,
    );
  }

  async getAposentadoria(id: number): Promise<ProcessoAposentadoria> {
    return await this.request<ProcessoAposentadoria>(
      `/api/v1/aposentadoria/${id}`,
    );
  }

  async getHistorico(id: number): Promise<ProcessoHistorico[]> {
    return await this.request<ProcessoHistorico[]>(
      `/api/v1/aposentadoria/${id}/historico`,
    );
  }
}
