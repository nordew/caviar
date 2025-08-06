package converter

type Converter struct {
	Order   *OrderConverter
	Product *ProductConverter
}

func NewConverter() *Converter {
	return &Converter{
		Order:   NewOrderConverter(),
		Product: NewProductConverter(),
	}
}

var defaultConverter = NewConverter()

func Default() *Converter {
	return defaultConverter
}