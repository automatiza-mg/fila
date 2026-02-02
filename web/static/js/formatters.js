/**
 * Registra os formatadores de input disponíveis.
 *
 * O uso deve ser feito através do atributo [data-format] em
 * tags <input>
 *
 * Exemplo:
 * ```html
 * <input type="text" data-format="cpf" />
 * ```
 */
export function initFormatters() {
  document.addEventListener("input", (e) => {
    const format = e.target.dataset.format;
    if (!format) return;

    switch (format) {
      case "cpf":
        if (e.target.pattern === "") {
          e.target.pattern = "\\d{3}\\.\\d{3}\\.\\d{3}-\\d{2}";
        }
        e.target.value = formatCPF(e.target.value);
        break;
      default:
        console.warn(`Formato desconhecido: ${format}`);
    }
  });
}

/**
 * Aplica a formatação de um CPF para o valor.
 *
 * @param {string} value
 */
export function formatCPF(value) {
  const digits = value.replace(/\D/g, "").slice(0, 11);

  return digits
    .replace(/(\d{3})(\d)/, "$1.$2")
    .replace(/(\d{3})(\d)/, "$1.$2")
    .replace(/(\d{3})(\d{1,2})$/, "$1-$2");
}
