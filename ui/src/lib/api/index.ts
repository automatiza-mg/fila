import { env } from "$env/dynamic/public";
import type {
  Paginated,
  Processo,
  ProcessoAposentadoria,
  Usuario,
} from "./types";

type Fetch = (
  input: RequestInfo | URL,
  init?: RequestInit,
) => Promise<Response>;

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
