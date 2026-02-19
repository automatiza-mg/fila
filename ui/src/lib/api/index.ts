import { env } from "$env/dynamic/public";
import type {
  Credenciais,
  ErrorResponse,
  Paginated,
  Processo,
  ProcessoAposentadoria,
  Token,
  Usuario,
} from "./types";

type Fetch = (
  input: RequestInfo | URL,
  init?: RequestInit,
) => Promise<Response>;

export class ApiError extends Error {
  constructor(
    public message: string,
    public status: number,
    public response: ErrorResponse,
  ) {
    super(message);
  }
}

export async function authenticate(
  { cpf, senha }: Credenciais,
  fetchFn: Fetch = fetch,
): Promise<Token> {
  const res = await fetchFn(`${env.PUBLIC_API_URL}/api/v1/auth/entrar`, {
    method: "POST",
    body: JSON.stringify({
      cpf,
      senha,
    }),
  });

  if (!res.ok) {
    const data = (await res.json()) as ErrorResponse;
    throw new ApiError(data.message, res.status, data);
  }

  return await res.json();
}

export class Client {
  constructor(
    public authToken: string,
    private fetch: Fetch = fetch,
    private baseUrl = `${env.PUBLIC_API_URL}/api/v1`,
  ) {}

  async usuarioAtual(): Promise<Usuario> {
    const res = await this.fetch(`${this.baseUrl}/auth/me`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });

    if (!res.ok) {
      throw new Error("Não foi possível carregar usuário atual");
    }

    return await res.json();
  }

  async listarProcessos(): Promise<Paginated<Processo>> {
    const res = await this.fetch(`${this.baseUrl}/processos`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });
    if (!res.ok) {
      throw new Error("Não foi possível listar processos");
    }

    return await res.json();
  }

  async listarUsuarios(): Promise<Usuario[]> {
    const res = await this.fetch(`${this.baseUrl}/usuarios`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });
    if (!res.ok) {
      throw new Error("Não foi possível listar usuarios");
    }

    return await res.json();
  }

  async listarAposentadoria(): Promise<Paginated<ProcessoAposentadoria>> {
    const res = await this.fetch(`${this.baseUrl}/aposentadoria`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });
    if (!res.ok) {
      throw new Error("Não foi possível listar usuarios");
    }

    return await res.json();
  }

  async getAposentadoria(id: number): Promise<ProcessoAposentadoria> {
    const res = await this.fetch(`${this.baseUrl}/aposentadoria/${id}`, {
      headers: {
        Authorization: `Bearer ${this.authToken}`,
      },
    });
    if (!res.ok) {
      throw new Error("Não foi possível buscar processo");
    }

    return await res.json();
  }
}
