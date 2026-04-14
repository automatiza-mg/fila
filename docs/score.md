# Score de Priorização de Processos

## Objetivo

O score determina a ordem de prioridade na fila de análise dos processos de aposentadoria. Quanto maior o score, maior a prioridade de análise.

## Regras de Cálculo

O score é calculado com base em quatro critérios, avaliados na seguinte ordem de precedência:

### 1. Processo Judicializado

Se o processo for judicializado, o score é fixado em **9**, independentemente dos demais critérios.

### 2. Prioridade Aprovada

Se o processo possuir uma solicitação de prioridade aprovada, o score é fixado em **8**, independentemente dos critérios de idade ou invalidez.

### 3. Cálculo Padrão

Caso o processo não seja judicializado nem possua prioridade aprovada, o score é calculado pela soma dos seguintes fatores:

| Critério | Pontos |
| :--- | :---: |
| Idade do requerente >= 60 anos | +1 |
| Idade do requerente >= 80 anos | +2 (cumulativo com o anterior) |
| Doença grave ou invalidez | +2 |

O score máximo pelo cálculo padrão é **5** (requerente com 80+ anos e invalidez: 1 + 2 + 2).

## Tabela Resumo

| Situação | Score |
| :--- | :---: |
| Judicializado | 9 |
| Prioridade aprovada | 8 |
| 80+ anos com invalidez | 5 |
| 80+ anos sem invalidez | 3 |
| 60-79 anos com invalidez | 3 |
| 60-79 anos sem invalidez | 1 |
| Menos de 60 anos com invalidez | 2 |
| Menos de 60 anos sem invalidez | 0 |
