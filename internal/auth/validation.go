package auth

import "github.com/automatiza-mg/fila/internal/validator"

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
