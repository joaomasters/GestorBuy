package mining

import "testing"

func TestParseItemID(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "ID puro",
			input: "MLB1234567890",
			want:  "MLB1234567890",
		},
		{
			name:  "ID puro minúsculo",
			input: "mlb1234567890",
			want:  "MLB1234567890",
		},
		{
			name:  "URL de produto com hífen",
			input: "https://produto.mercadolivre.com.br/MLB-1234567890-titulo-do-anuncio-_JM",
			want:  "MLB1234567890",
		},
		{
			name:  "URL de busca/catálogo",
			input: "https://www.mercadolivre.com.br/titulo/p/MLB12345678",
			want:  "MLB12345678",
		},
		{
			name:    "texto sem nenhum ID reconhecível",
			input:   "isso não é um link do mercado livre",
			wantErr: true,
		},
		{
			name:    "string vazia",
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseItemID(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("esperava erro pra input %q, mas passou com %q", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseItemID(%q) retornou erro inesperado: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("ParseItemID(%q) = %q, esperado %q", tc.input, got, tc.want)
			}
		})
	}
}
