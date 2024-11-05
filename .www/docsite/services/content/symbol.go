package content

import (
	"context"
	"errors"
	"github.com/tylermmorton/torque/.www/docsite/model"
)

var (
	ErrSymbolNotFound = errors.New("SymbolNotFound")
)

func (svc *contentService) GetSymbol(ctx context.Context, name string) (*model.Symbol, error) {
	if sym, ok := svc.symbols[name]; ok {
		return sym, nil
	}
	return nil, ErrSymbolNotFound
}
