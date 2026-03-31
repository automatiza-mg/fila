package auth

import (
	"fmt"
	"strings"

	"github.com/automatiza-mg/fila/internal/validator"
)

// ValidateCreateAdmin valida os parâmetros para criação de um administrador.
func ValidateCreateAdmin(v *validator.Validator, params CreateAdminParams) {
	v.Check(validator.NotBlank(params.Nome), "nome", "Campo obrigatório")
	v.Check(validator.MaxLength(params.Nome, 255), "nome", "Deve possuir até 255 caracteres")

	v.Check(validator.NotBlank(params.CPF), "cpf", "Campo obrigatório")
	v.Check(validator.Matches(params.CPF, validator.CpfRX), "cpf", "Deve ser um CPF válido")

	v.Check(validator.NotBlank(params.Email), "email", "Campo obrigatório")
	v.Check(validator.MaxLength(params.Email, 255), "email", "Deve possuir até 255 caracteres")
	v.Check(validator.Matches(params.Email, validator.EmailRX), "email", "Deve ser um email válido")

	v.Check(validator.NotBlank(params.Senha), "senha", "Campo obrigatório")
	v.Check(validator.MinLength(params.Senha, 8), "senha", "Deve possuir no mínimo 8 caracteres")
	v.Check(validator.StrongPassword(params.Senha), "senha", "Deve possuir pelo menos um número e um caractere especial")
}

// ValidateCreateUsuario valida os parâmetros para criação de um usuário.
func ValidateCreateUsuario(v *validator.Validator, params CreateUsuarioParams) {
	v.Check(validator.NotBlank(params.Nome), "nome", "Campo obrigatório")
	v.Check(validator.MaxLength(params.Nome, 255), "nome", "Deve possuir até 255 caracteres")

	v.Check(validator.NotBlank(params.CPF), "cpf", "Campo obrigatório")
	v.Check(validator.Matches(params.CPF, validator.CpfRX), "cpf", "Deve ser um CPF válido")

	v.Check(validator.NotBlank(params.Email), "email", "Campo obrigatório")
	v.Check(validator.MaxLength(params.Email, 255), "email", "Deve possuir até 255 caracteres")
	v.Check(validator.Matches(params.Email, validator.EmailRX), "email", "Deve ser um email válido")

	v.Check(validator.PermittedValue(params.Papel, AllowedPapeis...), "papel",
		fmt.Sprintf("Deve ser um dos valores: %s", strings.Join(AllowedPapeis, ", ")))
}

// ValidateAlterarSenha valida os parâmetros para alteração de senha pelo próprio usuário.
func ValidateAlterarSenha(v *validator.Validator, senhaAtual, novaSenha, confirmarNovaSenha string) {
	v.Check(validator.NotBlank(senhaAtual), "senha_atual", "Campo obrigatório")
	v.Check(validator.NotBlank(novaSenha), "nova_senha", "Campo obrigatório")
	v.Check(validator.MinLength(novaSenha, 8), "nova_senha", "Deve possuir no mínimo 8 caracteres")
	v.Check(validator.MaxLength(novaSenha, 60), "nova_senha", "Deve possuir no máximo 60 caracteres")
	v.Check(validator.StrongPassword(novaSenha), "nova_senha", "Deve possuir pelo menos um número e um caractere especial")
	v.Check(validator.NotBlank(confirmarNovaSenha), "confirmar_nova_senha", "Campo obrigatório")
	v.Check(novaSenha == confirmarNovaSenha, "confirmar_nova_senha", "Senhas devem ser idênticas")
}

// ValidateResetSenha valida os parâmetros para redefinição de senha.
func ValidateResetSenha(v *validator.Validator, senha, confirmarSenha string) {
	v.Check(validator.NotBlank(senha), "senha", "Campo obrigatório")
	v.Check(validator.MinLength(senha, 8), "senha", "Deve possuir no mínimo 8 caracteres")
	v.Check(validator.MaxLength(senha, 60), "senha", "Deve possuir no máximo 60 caracteres")
	v.Check(validator.StrongPassword(senha), "senha", "Deve possuir pelo menos um número e um caractere especial")
	v.Check(validator.NotBlank(confirmarSenha), "confirmar_senha", "Campo obrigatório")
	v.Check(senha == confirmarSenha, "confirmar_senha", "Senhas devem ser idênticas")
}
