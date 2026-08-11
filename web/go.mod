// Módulo Go vazio só pra marcar limite: sem isso, "go build ./..." rodado
// da raiz do repo desce em web/node_modules e encontra pacotes Go que
// vieram junto de dependências npm (ex.: flatted/golang), poluindo
// build/test/vet à toa. web/ é um projeto Node.js — não tem código Go de
// verdade aqui.
module gestorbuy-web-placeholder

go 1.23
