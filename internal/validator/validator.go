package validator

// Validator é a struct responsável por acumular erros de validação de campos.
type Validator struct {
	Errors      []string
	FieldErrors map[string]string
}

// SetError adiciona uma nova mensagem de erro não específica a um campo.
func (v *Validator) SetError(msg string) {
	if v.Errors == nil {
		v.Errors = make([]string, 0)
	}
	v.Errors = append(v.Errors, msg)
}

// SetFieldError adiciona um novo erro de campo ao validador. Se o nome do campo já possuir um erro, ignora a mensagem.
func (v *Validator) SetFieldError(name, msg string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	if _, ok := v.FieldErrors[name]; !ok {
		v.FieldErrors[name] = msg
	}
}

// Check adiciona um novo erro de campo ao validador se a condicional ok for false.
func (v *Validator) Check(ok bool, name, msg string) {
	if !ok {
		v.SetFieldError(name, msg)
	}
}

// Valid reporta se o validador não possui erros.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.Errors) == 0
}

// Message retorna a mensagem de erro para algum campo, caso exista.
func (v *Validator) Message(name string) string {
	if v.FieldErrors == nil {
		return ""
	}
	return v.FieldErrors[name]
}
