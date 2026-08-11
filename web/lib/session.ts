// Nome do cookie httpOnly que guarda o JWT emitido pela API Go. Único lugar
// que define esse nome — tudo que precisa ler/gravar o cookie importa
// daqui, nunca usa a string literal direto.
export const SESSION_COOKIE = "gestorbuy_session";

// Alinhado ao TTL do JWT no backend (internal/config/config.go: 24h).
export const SESSION_MAX_AGE_SECONDS = 60 * 60 * 24;
