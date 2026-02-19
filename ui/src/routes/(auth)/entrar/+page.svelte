<script>
  import { login } from "./login.remote";
</script>

<h1>Entrar</h1>

<form {...login}>
  <div>
    <label for="cpf">CPF</label>
    <input {...login.fields.cpf.as("text")} />
  </div>

  <div>
    <label for="senha">Senha</label>
    <input {...login.fields.senha.as("password")} />
  </div>

  <button>Enviar</button>
</form>
