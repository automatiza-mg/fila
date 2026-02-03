package main

import "net/http"

// Contém valores usados por todas as páginas de aplicação.
type basePage struct {
	// HTMLTitle é o valor usado na tag <title> da página.
	HTMLTitle string
}

func (app *application) newBasePage(_ *http.Request, title string) basePage {
	return basePage{
		HTMLTitle: title,
	}
}
