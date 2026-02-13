import type {
  Analista,
  AnalistaCreateRequest,
  ApiErrorBody,
  CadastrarRequest,
  DatalakeProcesso,
  Documento,
  EntrarRequest,
  Escopo,
  HistoricoStatusProcesso,
  PaginatedResult,
  PaginationParams,
  Processo,
  ProcessoAposentadoria,
  ProcessoCreateRequest,
  RecuperarSenhaRequest,
  RedefinirSenhaRequest,
  Servidor,
  StatusProcesso,
  Token,
  UnidadeSei,
  Usuario,
  UsuarioCreateRequest,
} from "./types";

export class ApiError extends Error {
  status: number;
  body: ApiErrorBody;

  constructor(status: number, body: ApiErrorBody) {
    super(body.message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }

  get fieldErrors(): Record<string, string> | undefined {
    return this.body.errors;
  }
}

export interface ClientOptions {
  baseUrl?: string;
  getToken?: () => string | null;
}

function qs(params: Record<string, string | number | boolean | undefined | null>): string {
  const entries = Object.entries(params).filter(
    ([, v]) => v !== undefined && v !== null && v !== "",
  );
  if (entries.length === 0) return "";
  const search = new URLSearchParams();
  for (const [k, v] of entries) {
    search.set(k, String(v));
  }
  return `?${search.toString()}`;
}

export function createClient(opts: ClientOptions = {}) {
  const baseUrl = (opts.baseUrl ?? "/api/v1").replace(/\/+$/, "");

  async function request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const headers: Record<string, string> = {};

    const token = opts.getToken?.();
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
    }

    const res = await fetch(`${baseUrl}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    if (!res.ok) {
      let errorBody: ApiErrorBody;
      try {
        errorBody = await res.json();
      } catch {
        errorBody = { message: res.statusText };
      }
      throw new ApiError(res.status, errorBody);
    }

    if (res.status === 204 || res.status === 202) {
      return undefined as T;
    }

    return res.json();
  }

  const auth = {
    entrar(data: EntrarRequest): Promise<Token> {
      return request("POST", "/auth/entrar", data);
    },

    tokenInfo(token: string, escopo: Exclude<Escopo, "auth">): Promise<Usuario> {
      return request("GET", `/auth/token${qs({ token, escopo })}`);
    },

    cadastrar(data: CadastrarRequest): Promise<void> {
      return request("POST", "/auth/cadastrar", data);
    },

    recuperarSenha(data: RecuperarSenhaRequest): Promise<void> {
      return request("POST", "/auth/recuperar-senha", data);
    },

    redefinirSenha(data: RedefinirSenhaRequest): Promise<void> {
      return request("POST", "/auth/redefinir-senha", data);
    },

    me(): Promise<Usuario> {
      return request("GET", "/auth/me");
    },

    meAnalista(): Promise<Analista> {
      return request("GET", "/auth/me/analista");
    },
  };

  const usuarios = {
    list(params?: { papel?: string }): Promise<Usuario[]> {
      return request("GET", `/usuarios${qs({ papel: params?.papel })}`);
    },

    create(data: UsuarioCreateRequest): Promise<Usuario> {
      return request("POST", "/usuarios", data);
    },

    get(usuarioID: number): Promise<Usuario> {
      return request("GET", `/usuarios/${usuarioID}`);
    },

    delete(usuarioID: number): Promise<void> {
      return request("DELETE", `/usuarios/${usuarioID}`);
    },

    enviarCadastro(usuarioID: number): Promise<void> {
      return request("POST", `/usuarios/${usuarioID}/enviar-cadastro`);
    },

    getAnalista(usuarioID: number): Promise<Analista> {
      return request("GET", `/usuarios/${usuarioID}/analista`);
    },

    createAnalista(usuarioID: number, data: AnalistaCreateRequest): Promise<Analista> {
      return request("POST", `/usuarios/${usuarioID}/analista`, data);
    },

    afastarAnalista(usuarioID: number): Promise<void> {
      return request("POST", `/usuarios/${usuarioID}/analista/afastar`);
    },

    retornarAnalista(usuarioID: number): Promise<void> {
      return request("POST", `/usuarios/${usuarioID}/analista/retornar`);
    },
  };

  const processos = {
    list(
      params?: PaginationParams & { numero?: string },
    ): Promise<PaginatedResult<Processo>> {
      return request(
        "GET",
        `/processos${qs({
          page: params?.page,
          limit: params?.limit,
          numero: params?.numero,
        })}`,
      );
    },

    create(data: ProcessoCreateRequest): Promise<Processo> {
      return request("POST", "/processos", data);
    },

    get(processoID: string): Promise<Processo> {
      return request("GET", `/processos/${processoID}`);
    },

    documentos(processoID: string): Promise<Documento[]> {
      return request("GET", `/processos/${processoID}/documentos`);
    },
  };

  const aposentadoria = {
    list(
      params?: PaginationParams & { numero?: string; status?: StatusProcesso },
    ): Promise<PaginatedResult<ProcessoAposentadoria>> {
      return request(
        "GET",
        `/aposentadoria${qs({
          page: params?.page,
          limit: params?.limit,
          numero: params?.numero,
          status: params?.status,
        })}`,
      );
    },

    get(paID: number): Promise<ProcessoAposentadoria> {
      return request("GET", `/aposentadoria/${paID}`);
    },

    historico(paID: number): Promise<HistoricoStatusProcesso[]> {
      return request("GET", `/aposentadoria/${paID}/historico`);
    },
  };

  const analistas = {
    list(): Promise<Analista[]> {
      return request("GET", "/analistas");
    },
  };

  const unidades = {
    list(): Promise<UnidadeSei[]> {
      return request("GET", "/unidades");
    },
  };

  const datalake = {
    processos(unidade: string): Promise<DatalakeProcesso[]> {
      return request("GET", `/datalake/processos${qs({ unidade })}`);
    },

    unidadesProcessos(): Promise<string[]> {
      return request("GET", "/datalake/processos/unidades");
    },

    servidor(cpf: string): Promise<Servidor> {
      return request("GET", `/datalake/servidores/${cpf}`);
    },
  };

  return {
    auth,
    usuarios,
    processos,
    aposentadoria,
    analistas,
    unidades,
    datalake,
  };
}

export type ApiClient = ReturnType<typeof createClient>;
