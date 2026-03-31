export function calcularIdade(nascimento: string): number {
  const nasc = new Date(nascimento);
  const hoje = new Date();
  let idade = hoje.getFullYear() - nasc.getFullYear();
  if (
    hoje.getMonth() < nasc.getMonth() ||
    (hoje.getMonth() === nasc.getMonth() && hoje.getDate() < nasc.getDate())
  ) {
    idade--;
  }
  return idade;
}
