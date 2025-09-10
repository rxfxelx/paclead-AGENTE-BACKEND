package flow

import (
    "context"
    "fmt"
    "strings"

    "pac-lead-agent/internal/clients"
)

func SendProductsCarousel(ctx context.Context, pl *clients.PacLead, whats *clients.Whats, cnpj, number string, ids []string) error {
    if len(ids) == 0 {
        return nil
    }
    cards := make([]map[string]any, 0, len(ids))
    for i := 0; i < len(ids) && i < 5; i++ {
        id := ids[i]
        idCopy := id
        prods, err := pl.Produtos(ctx, cnpj, &idCopy)
        if err != nil || len(prods) == 0 {
            continue
        }
        p := prods[0]
        nome, _ := p["nome"].(string)
        desc, _ := p["descricao"].(string)
        preco := fmt.Sprintf("%v", p["preco"])
        text := strings.ToLower(desc) + "\nPreço:R$" + preco
        image := fmt.Sprintf("http://paclead.com.br:8889/produtos/imagem?id=%s&id_empresa=1", id)
        cards = append(cards, map[string]any{
            "text":  text,
            "image": image,
            "buttons": []map[string]any{{
                "id":   fmt.Sprintf("Vou querer o %s", nome),
                "text": fmt.Sprintf("Vou querer o %s", nome),
                "type": "REPLY",
            }},
        })
    }
    if len(cards) == 0 { return nil }
    return whats.SendCarousel(ctx, number, "Encante-se com os destaques!", cards)
}
