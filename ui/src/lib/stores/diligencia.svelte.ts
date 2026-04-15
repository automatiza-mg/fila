import { getContext, setContext } from "svelte";

export type CategoriaDiligencia = {
  nome: string;
  subcategorias?: string[];
};

export type Diligencia = {
  tipo: string;
  subcategorias: string[];
  detalhe: string;
};

export const categoriasDiligencia: CategoriaDiligencia[] = [
  {
    nome: "Documentos Obrigatórios Ausentes",
    subcategorias: [
      "Dois relatórios de conferência extraídos da Fipa Eletrônica/SISAP: Dados cadastrais e Dados Funcionais",
      "Aposentadoria Voluntária: Requerimento de Aposentadoria",
      "Aposentadoria por incapacidade permanente: Laudo Médico Oficial",
      "Aposentadoria Compulsória: Cópia autenticada da certidão de nascimento ou casamento",
      "Declaração de Acúmulo de Cargos/Proventos",
      "Cópia da publicação constando as informações referentes à licitude de cargos",
      "Cópia da decisão do processo administrativo ou declaração informando a finalização e os termos da decisão do processo administrativo. Cópia da decisão judicial, quando se tratar de direitos reconhecidos judicialmente",
      "Cópia da certidão de nascimento ou casamento ou carteira de identidade ou outro documento público que comprove o nome completo e a idade do(a) servidor(a)",
      "Certidões de tempo de serviço/contribuição averbadas (INSS municipal, outro estado, federal e declarações ou demais documentos inerentes à averbação)",
      "FIPA - Tempo Averbado",
      "FIPA - Matriz de Apuração de Tempo de acordo à regra da aposentadoria",
      "FIPA - Matriz de Contagem de Tempo",
      "FIPA - Dados Cadastrais",
      "Planilha de cálculo de proventos por média e Formulário da Última remuneração nos casos de aposentadoria por média com vigência anterior a 15.09.2020 e direito adquirido da EC 104/20",
      "Planilha de cálculo de proventos por média nos casos de aposentadoria por média após EC 104/20",
      "Demonstrativo de pagamento do mês de vigência da aposentadoria",
      "Declaração do efetivo exercício expedido pelo órgão que recebeu o servidor na situação de adjunção ou disposição",
    ],
  },
  { nome: "Documentos com Informações Incompletas/Faltantes" },
  {
    nome: "Documentos com Erros ou Incompatíveis com o Processo Analisado",
  },
  { nome: "Divergências de Informações entre Processo e SISAP" },
  { nome: "Alteração de Dados Após o Envio" },
  {
    nome: "Documento com Baixa Nitidez",
    subcategorias: [
      "Aposentadoria Voluntária: Requerimento de Aposentadoria",
      "Aposentadoria por incapacidade permanente: Laudo Médico Oficial",
      "Aposentadoria Compulsória: Cópia autenticada da certidão de nascimento ou casamento",
      "Declaração de Acúmulo de Cargos/Proventos",
      "Cópia da publicação constando as informações referentes à licitude de cargos",
      "Cópia da decisão do processo administrativo ou declaração informando a finalização e os termos da decisão do processo administrativo. Cópia da decisão judicial, quando se tratar de direitos reconhecidos judicialmente",
      "Cópia da certidão de nascimento ou casamento ou carteira de identidade ou outro documento público que comprove o nome completo e a idade do(a) servidor(a)",
      "Certidões de tempo de serviço/contribuição averbadas (INSS municipal, outro estado, federal e declarações ou demais documentos inerentes à averbação)",
      "Declaração do efetivo exercício expedido pelo órgão que recebeu o servidor na situação de adjunção ou disposição",
    ],
  },
  { nome: "Inconsistência na Análise do Órgão de Origem" },
];

class DiligenciaStore {
  diligencias = $state<Diligencia[]>([]);

  add(diligencia: Diligencia) {
    this.diligencias.push({
      ...diligencia,
      subcategorias: [...diligencia.subcategorias],
    });
  }

  update(index: number, diligencia: Diligencia) {
    this.diligencias[index] = {
      ...diligencia,
      subcategorias: [...diligencia.subcategorias],
    };
  }

  removeByIndex(index: number) {
    this.diligencias.splice(index, 1);
  }
}

export function setDiligenciaState() {
  setContext("diligencia", new DiligenciaStore());
}

export function getDiligenciaState() {
  return getContext<DiligenciaStore>("diligencia");
}
