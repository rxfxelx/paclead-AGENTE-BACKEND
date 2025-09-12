package clients

// NOTA IMPORTANTE:
// Existe uma implementação ativa do cliente da Plataforma em "plataform.go"
// (histórico do projeto). Este arquivo é propositalmente neutro para evitar
// duplicidade de tipos/funções (Platform, NewPlatform, GetAgentSettings).
//
// O restante do código (dispatcher, etc.) deve continuar usando:
//   clients.NewPlatform(baseURL).GetAgentSettings(ctx, orgID, flowID)
//
// Se em algum momento você quiser migrar a implementação para "platform.go",
// basta mover/renomear a implementação de "plataform.go" para cá e manter
// os mesmos símbolos públicos.
